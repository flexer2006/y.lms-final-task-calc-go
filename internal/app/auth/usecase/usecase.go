// Package usecase реализует бизнес-логику аутентификации пользователей,
// включая регистрацию, вход, валидацию и обновление токенов, а также выход из системы.
// Пакет обеспечивает безопасное взаимодействие между пользовательскими запросами и хранилищем данных.
package usecase

import (
	"context"
	"fmt"
	"time"

	domainerrors "github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/errord"
	authmodels "github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/auth"
	authapi "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/api/auth"
	authrepo "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/repository/auth"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/service/jwt"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/service/password"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AuthUseCase представляет сервис авторизации, инкапсулирующий бизнес-логику
// аутентификации и управления сессиями пользователей.
// Структура следует принципам чистой архитектуры, используя репозитории
// и сервисы через их интерфейсы для обеспечения гибкости и тестируемости.
type AuthUseCase struct {
	userRepo    authrepo.UserRepository  // Репозиторий для работы с данными пользователей
	tokenRepo   authrepo.TokenRepository // Репозиторий для работы с токенами аутентификации
	passwordSvc password.Service         // Сервис для хеширования и проверки паролей
	jwtSvc      jwt.Service              // Сервис для создания и валидации JWT токенов
}

// Проверка, что AuthUseCase реализует интерфейс UseCaseUser
var _ authapi.UseCaseUser = (*AuthUseCase)(nil)

// NewAuthUseCase создает новый экземпляр сервиса авторизации с необходимыми зависимостями.
// Этот конструктор следует принципу инверсии зависимостей, принимая репозитории и сервисы
// в качестве интерфейсов, что повышает гибкость и тестируемость системы.
//
// Параметры:
//   - userRepo: репозиторий для работы с пользователями
//   - tokenRepo: репозиторий для работы с токенами
//   - passwordSvc: сервис для работы с паролями
//   - jwtSvc: сервис для работы с JWT токенами
//
// Возвращает:
//   - экземпляр AuthUseCase, готовый к использованию
func NewAuthUseCase(
	userRepo authrepo.UserRepository,
	tokenRepo authrepo.TokenRepository,
	passwordSvc password.Service,
	jwtSvc jwt.Service,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		passwordSvc: passwordSvc,
		jwtSvc:      jwtSvc,
	}
}

// Register регистрирует нового пользователя в системе.
// Процесс включает проверку существования пользователя с таким логином,
// хеширование пароля и сохранение данных нового пользователя в базе данных.
//
// Включает следующие этапы:
//  1. Проверка существования пользователя с указанным логином
//  2. Хеширование пароля с использованием безопасного алгоритма
//  3. Создание новой записи пользователя в хранилище
//  4. Возврат идентификатора созданного пользователя
//
// Параметры:
//   - ctx: контекст выполнения операции
//   - login: логин пользователя (должен быть уникальным)
//   - password: пароль пользователя в открытом виде (будет хешироваться)
//
// Возвращает:
//   - uuid.UUID: идентификатор созданного пользователя
//   - error: ошибка операции или nil при успехе
func (uc *AuthUseCase) Register(ctx context.Context, login, password string) (uuid.UUID, error) {
	const op = "AuthUseCase.Register"
	log := logger.ContextLogger(ctx, nil).With(zap.String("op", op), zap.String("login", login))

	existingUser, err := uc.userRepo.FindByLogin(ctx, login)
	if err != nil {
		log.Error("Failed to check user existence", zap.Error(err))
		return uuid.Nil, fmt.Errorf("%s: %w", op, domainerrors.ErrInternalServerError)
	}

	if existingUser != nil {
		log.Warn("User already exists")
		return uuid.Nil, domainerrors.ErrUserAlreadyExists
	}

	hashedPassword, err := uc.passwordSvc.Hash(ctx, password)
	if err != nil {
		log.Error("Failed to hash password", zap.Error(err))
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	user := &authmodels.User{
		ID:           uuid.New(),
		Login:        login,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	createdUser, err := uc.userRepo.Create(ctx, user)
	if err != nil {
		log.Error("Failed to create user", zap.Error(err))
		return uuid.Nil, fmt.Errorf("%s: %w", op, domainerrors.ErrInternalServerError)
	}

	log.Info("User registered successfully", zap.String("userId", createdUser.ID.String()))
	return createdUser.ID, nil
}

// Login выполняет аутентификацию пользователя и создает пару токенов (access и refresh).
// В случае успешной авторизации генерируются новые токены, и refresh токен сохраняется
// в базу данных для последующего использования и возможности его отзыва.
//
// Процесс аутентификации:
//  1. Поиск пользователя по логину
//  2. Проверка хешированного пароля
//  3. Генерация пары токенов доступа
//  4. Сохранение refresh токена в базе данных
//
// Параметры:
//   - ctx: контекст выполнения операции
//   - login: логин пользователя
//   - password: пароль пользователя в открытом виде
//
// Возвращает:
//   - *authmodels.TokenPair: пара токенов (access и refresh) при успешной аутентификации
//   - error: ошибка операции или nil при успехе
func (uc *AuthUseCase) Login(ctx context.Context, login, password string) (*authmodels.TokenPair, error) {
	const op = "AuthUseCase.Login"
	log := logger.ContextLogger(ctx, nil).With(zap.String("op", op), zap.String("login", login))

	user, err := uc.userRepo.FindByLogin(ctx, login)
	if err != nil {
		log.Error("Failed to find user", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, domainerrors.ErrInternalServerError)
	}

	if user == nil {
		log.Warn("User not found")
		return nil, domainerrors.ErrInvalidCredentials
	}

	valid, err := uc.passwordSvc.Verify(ctx, password, user.PasswordHash)
	if err != nil {
		log.Error("Password verification error", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, domainerrors.ErrInternalServerError)
	}

	if !valid {
		log.Warn("Invalid password")
		return nil, domainerrors.ErrInvalidCredentials
	}

	tokenPair, err := uc.jwtSvc.GenerateTokens(ctx, user.ID, user.Login)
	if err != nil {
		log.Error("Failed to generate tokens", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, domainerrors.ErrInternalServerError)
	}

	token := &authmodels.Token{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenStr:  tokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(uc.jwtSvc.GetRefreshTokenTTL()),
		CreatedAt: time.Now(),
		IsRevoked: false,
	}

	if err := uc.tokenRepo.Store(ctx, token); err != nil {
		log.Error("Failed to store refresh token", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, domainerrors.ErrInternalServerError)
	}

	log.Info("User logged in successfully", zap.String("userId", user.ID.String()))
	return tokenPair, nil
}

// ValidateToken проверяет действительность access токена и возвращает ID пользователя.
// Выполняет криптографическую проверку подписи токена и проверяет существование
// пользователя в системе.
//
// Этапы проверки:
//  1. Криптографическая валидация токена
//  2. Извлечение ID пользователя из токена
//  3. Проверка существования пользователя в системе
//
// Параметры:
//   - ctx: контекст выполнения операции
//   - tokenStr: строка токена для проверки
//
// Возвращает:
//   - uuid.UUID: идентификатор пользователя, которому принадлежит токен
//   - error: ошибка операции или nil при успешной валидации
func (uc *AuthUseCase) ValidateToken(ctx context.Context, tokenStr string) (uuid.UUID, error) {
	const op = "AuthUseCase.ValidateToken"
	log := logger.ContextLogger(ctx, nil).With(zap.String("op", op))

	userID, err := uc.jwtSvc.ValidateToken(ctx, tokenStr)
	if err != nil {
		log.Debug("Token validation failed", zap.Error(err))
		return uuid.Nil, domainerrors.ErrInvalidToken
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		log.Error("Failed to find user", zap.Error(err))
		return uuid.Nil, fmt.Errorf("%s: %w", op, domainerrors.ErrInternalServerError)
	}

	if user == nil {
		log.Warn("User not found", zap.String("userId", userID.String()))
		return uuid.Nil, domainerrors.ErrUserNotFound
	}

	log.Debug("Token validated successfully", zap.String("userId", userID.String()))
	return userID, nil
}

// RefreshToken обновляет пару токенов (access и refresh) при наличии
// действительного refresh токена. При успешном обновлении, старый refresh токен
// отзывается и создается новая пара токенов.
//
// Процесс обновления включает:
//  1. Парсинг refresh токена и извлечение идентификатора пользователя
//  2. Поиск токена в базе данных и проверка его статуса (не отозван, не просрочен)
//  3. Отзыв старого токена
//  4. Генерация новой пары токенов
//  5. Сохранение нового refresh токена в базе данных
//
// Параметры:
//   - ctx: контекст выполнения операции
//   - refreshTokenStr: строка refresh токена
//
// Возвращает:
//   - *authmodels.TokenPair: новая пара токенов при успешном обновлении
//   - error: ошибка операции или nil при успехе
func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshTokenStr string) (*authmodels.TokenPair, error) {
	const op = "AuthUseCase.RefreshToken"
	log := logger.ContextLogger(ctx, nil).With(zap.String("op", op))

	claims, err := uc.jwtSvc.ParseToken(ctx, refreshTokenStr)
	if err != nil {
		log.Debug("Failed to parse refresh token", zap.Error(err))
		return nil, domainerrors.ErrInvalidToken
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		log.Debug("Invalid user_id claim")
		return nil, domainerrors.ErrInvalidToken
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Debug("Invalid user_id format", zap.String("userId", userIDStr))
		return nil, domainerrors.ErrInvalidToken
	}

	token, err := uc.tokenRepo.FindByTokenString(ctx, refreshTokenStr)
	if err != nil {
		log.Error("Failed to find token", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, domainerrors.ErrInternalServerError)
	}

	if token == nil {
		log.Debug("Token not found in database")
		return nil, domainerrors.ErrTokenNotFound
	}

	if token.IsRevoked {
		log.Debug("Token is revoked")
		return nil, domainerrors.ErrTokenRevoked
	}

	if token.ExpiresAt.Before(time.Now()) {
		log.Debug("Token is expired")
		return nil, domainerrors.ErrTokenExpired
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		log.Error("Failed to find user", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, domainerrors.ErrInternalServerError)
	}

	if user == nil {
		log.Warn("User not found", zap.String("userId", userID.String()))
		return nil, domainerrors.ErrUserNotFound
	}

	if err := uc.tokenRepo.RevokeToken(ctx, refreshTokenStr); err != nil {
		log.Error("Failed to revoke old token", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, domainerrors.ErrInternalServerError)
	}

	newTokenPair, err := uc.jwtSvc.GenerateTokens(ctx, user.ID, user.Login)
	if err != nil {
		log.Error("Failed to generate new tokens", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, domainerrors.ErrInternalServerError)
	}

	newToken := &authmodels.Token{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenStr:  newTokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(uc.jwtSvc.GetRefreshTokenTTL()),
		CreatedAt: time.Now(),
		IsRevoked: false,
	}

	if err := uc.tokenRepo.Store(ctx, newToken); err != nil {
		log.Error("Failed to store new refresh token", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, domainerrors.ErrInternalServerError)
	}

	log.Info("Tokens refreshed successfully", zap.String("userId", user.ID.String()))
	return newTokenPair, nil
}

// Logout завершает сессию пользователя путем отзыва refresh токена.
// После успешного выхода токен становится недействительным и не может быть
// использован для обновления пары токенов.
//
// Процесс выхода включает:
//  1. Парсинг токена и извлечение идентификатора пользователя
//  2. Поиск токена в базе данных
//  3. Отзыв токена из системы
//
// Параметры:
//   - ctx: контекст выполнения операции
//   - tokenStr: строка refresh токена для отзыва
//
// Возвращает:
//   - error: ошибка операции или nil при успешном выходе
func (uc *AuthUseCase) Logout(ctx context.Context, tokenStr string) error {
	const op = "AuthUseCase.Logout"
	log := logger.ContextLogger(ctx, nil).With(zap.String("op", op))

	claims, err := uc.jwtSvc.ParseToken(ctx, tokenStr)
	if err != nil {
		log.Debug("Failed to parse token", zap.Error(err))
		return domainerrors.ErrInvalidToken
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		log.Debug("Invalid user_id claim")
		return domainerrors.ErrInvalidToken
	}

	token, err := uc.tokenRepo.FindByTokenString(ctx, tokenStr)
	if err != nil {
		log.Error("Failed to find token", zap.Error(err))
		return fmt.Errorf("%s: %w", op, domainerrors.ErrInternalServerError)
	}

	if token == nil {
		log.Debug("Token not found in database")
		return nil
	}

	if err := uc.tokenRepo.RevokeToken(ctx, tokenStr); err != nil {
		log.Error("Failed to revoke token", zap.Error(err))
		return fmt.Errorf("%s: %w", op, domainerrors.ErrInternalServerError)
	}

	log.Info("User logged out successfully", zap.String("userId", userIDStr))
	return nil
}

// CleanupExpiredTokens выполняет очистку истекших токенов из базы данных.
// Эта операция может выполняться периодически для поддержания базы данных в актуальном
// состоянии и предотвращения её избыточного роста.
//
// Параметры:
//   - ctx: контекст выполнения операции
//
// Возвращает:
//   - error: ошибка операции или nil при успешной очистке
func (uc *AuthUseCase) CleanupExpiredTokens(ctx context.Context) error {
	const op = "AuthUseCase.CleanupExpiredTokens"
	log := logger.ContextLogger(ctx, nil).With(zap.String("op", op))

	if err := uc.tokenRepo.DeleteExpiredTokens(ctx, time.Now()); err != nil {
		log.Error("Failed to delete expired tokens", zap.Error(err))
		return fmt.Errorf("%s: %w", op, domainerrors.ErrInternalServerError)
	}

	log.Info("Expired tokens cleaned up successfully")
	return nil
}

// Close освобождает ресурсы, используемые AuthUseCase. В текущей реализации
// этот метод не выполняет никаких действий, но может быть расширен в будущем
// для корректного закрытия соединений или очистки ресурсов.
//
// Возвращает:
//   - error: ошибка при освобождении ресурсов или nil при успехе
func (uc *AuthUseCase) Close() error {
	return nil
}
