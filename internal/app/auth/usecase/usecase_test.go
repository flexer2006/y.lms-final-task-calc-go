package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	domainerrors "github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/errord"
	authmodels "github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/auth"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Fatal(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) With(fields ...logger.Field) logger.Logger {
	args := m.Called(fields)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) SetLevel(lvl logger.LogLevel) {
	m.Called(lvl)
}

func (m *MockLogger) GetLevel() logger.LogLevel {
	args := m.Called()
	return args.Get(0).(logger.LogLevel)
}

func (m *MockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockLogger) RawLogger() *zap.Logger {
	args := m.Called()
	return args.Get(0).(*zap.Logger)
}

func setupTestContext() (context.Context, *MockLogger) {
	mockLog := new(MockLogger)
	mockLog.On("With", mock.Anything).Return(mockLog).Maybe()
	mockLog.On("Debug", mock.Anything, mock.Anything).Maybe()
	mockLog.On("Info", mock.Anything, mock.Anything).Maybe()
	mockLog.On("Warn", mock.Anything, mock.Anything).Maybe()
	mockLog.On("Error", mock.Anything, mock.Anything).Maybe()
	mockLog.On("RawLogger").Return(zap.NewNop()).Maybe()

	ctx := logger.WithLogger(context.Background(), mockLog)
	return ctx, mockLog
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *authmodels.User) (*authmodels.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authmodels.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*authmodels.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authmodels.User), args.Error(1)
}

func (m *MockUserRepository) FindByLogin(ctx context.Context, login string) (*authmodels.User, error) {
	args := m.Called(ctx, login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authmodels.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *authmodels.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

type MockTokenRepository struct {
	mock.Mock
}

func (m *MockTokenRepository) Store(ctx context.Context, token *authmodels.Token) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockTokenRepository) FindByID(ctx context.Context, id uuid.UUID) (*authmodels.Token, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authmodels.Token), args.Error(1)
}

func (m *MockTokenRepository) FindByTokenString(ctx context.Context, tokenStr string) (*authmodels.Token, error) {
	args := m.Called(ctx, tokenStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authmodels.Token), args.Error(1)
}

func (m *MockTokenRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*authmodels.Token, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*authmodels.Token), args.Error(1)
}

func (m *MockTokenRepository) RevokeToken(ctx context.Context, tokenStr string) error {
	args := m.Called(ctx, tokenStr)
	return args.Error(0)
}

func (m *MockTokenRepository) DeleteExpiredTokens(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

func (m *MockTokenRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type MockPasswordService struct {
	mock.Mock
}

func (m *MockPasswordService) Hash(ctx context.Context, password string) (string, error) {
	args := m.Called(ctx, password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordService) Verify(ctx context.Context, password, hash string) (bool, error) {
	args := m.Called(ctx, password, hash)
	return args.Bool(0), args.Error(1)
}

func (m *MockPasswordService) GenerateRandom(ctx context.Context, length int) (string, error) {
	args := m.Called(ctx, length)
	return args.String(0), args.Error(1)
}

type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateTokens(ctx context.Context, userID uuid.UUID, login string) (*authmodels.TokenPair, error) {
	args := m.Called(ctx, userID, login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authmodels.TokenPair), args.Error(1)
}

func (m *MockJWTService) ValidateToken(ctx context.Context, tokenStr string) (uuid.UUID, error) {
	args := m.Called(ctx, tokenStr)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockJWTService) ParseToken(ctx context.Context, tokenStr string) (map[string]interface{}, error) {
	args := m.Called(ctx, tokenStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockJWTService) GetAccessTokenTTL() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *MockJWTService) GetRefreshTokenTTL() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *MockJWTService) GetTokenTTL() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name           string
		login          string
		password       string
		mockSetup      func(*MockUserRepository, *MockPasswordService)
		expectedUserID uuid.UUID
		expectedError  error
	}{
		{
			name:     "Success",
			login:    "testuser",
			password: "password123",
			mockSetup: func(userRepo *MockUserRepository, passwordSvc *MockPasswordService) {
				userRepo.On("FindByLogin", mock.Anything, "testuser").Return(nil, nil)
				passwordSvc.On("Hash", mock.Anything, "password123").Return("hashedpassword", nil)

				userID := uuid.New()
				userRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *authmodels.User) bool {
					return user.Login == "testuser" && user.PasswordHash == "hashedpassword"
				})).Return(&authmodels.User{ID: userID}, nil)
			},
			expectedError: nil,
		},
		{
			name:     "UserAlreadyExists",
			login:    "existinguser",
			password: "password123",
			mockSetup: func(userRepo *MockUserRepository, passwordSvc *MockPasswordService) {
				userRepo.On("FindByLogin", mock.Anything, "existinguser").Return(&authmodels.User{}, nil)
			},
			expectedError: domainerrors.ErrUserAlreadyExists,
		},
		{
			name:     "HashError",
			login:    "testuser",
			password: "password123",
			mockSetup: func(userRepo *MockUserRepository, passwordSvc *MockPasswordService) {
				userRepo.On("FindByLogin", mock.Anything, "testuser").Return(nil, nil)
				passwordSvc.On("Hash", mock.Anything, "password123").Return("", errors.New("hash error"))
			},
			expectedError: errors.New("hash error"),
		},
		{
			name:     "CreateError",
			login:    "testuser",
			password: "password123",
			mockSetup: func(userRepo *MockUserRepository, passwordSvc *MockPasswordService) {
				userRepo.On("FindByLogin", mock.Anything, "testuser").Return(nil, nil)
				passwordSvc.On("Hash", mock.Anything, "password123").Return("hashedpassword", nil)

				userRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *authmodels.User) bool {
					return user.Login == "testuser" && user.PasswordHash == "hashedpassword"
				})).Return(nil, errors.New("db error"))
			},
			expectedError: domainerrors.ErrInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := setupTestContext()
			userRepo := new(MockUserRepository)
			tokenRepo := new(MockTokenRepository)
			passwordSvc := new(MockPasswordService)
			jwtSvc := new(MockJWTService)

			passwordSvc.On("GenerateRandom", mock.Anything, mock.Anything).Return("", nil).Maybe()
			jwtSvc.On("GetTokenTTL").Return(time.Hour).Maybe()

			tt.mockSetup(userRepo, passwordSvc)

			uc := NewAuthUseCase(userRepo, tokenRepo, passwordSvc, jwtSvc)

			userID, err := uc.Register(ctx, tt.login, tt.password)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if tt.expectedError == domainerrors.ErrUserAlreadyExists ||
					tt.expectedError == domainerrors.ErrInternalServerError {
					assert.ErrorIs(t, err, tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, userID)
			}

			userRepo.AssertExpectations(t)
			passwordSvc.AssertExpectations(t)
		})
	}
}

func TestLogin(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name          string
		login         string
		password      string
		mockSetup     func(*MockUserRepository, *MockPasswordService, *MockJWTService, *MockTokenRepository)
		expectedError error
	}{
		{
			name:     "Success",
			login:    "testuser",
			password: "password123",
			mockSetup: func(userRepo *MockUserRepository, passwordSvc *MockPasswordService, jwtSvc *MockJWTService, tokenRepo *MockTokenRepository) {
				userRepo.On("FindByLogin", mock.Anything, "testuser").Return(&authmodels.User{
					ID:           userID,
					Login:        "testuser",
					PasswordHash: "hashedpassword",
				}, nil)

				passwordSvc.On("Verify", mock.Anything, "password123", "hashedpassword").Return(true, nil)

				jwtSvc.On("GenerateTokens", mock.Anything, userID, "testuser").Return(&authmodels.TokenPair{
					AccessToken:  "access-token",
					RefreshToken: "refresh-token",
				}, nil)

				jwtSvc.On("GetRefreshTokenTTL").Return(24 * time.Hour)

				tokenRepo.On("Store", mock.Anything, mock.MatchedBy(func(token *authmodels.Token) bool {
					return token.UserID == userID && token.TokenStr == "refresh-token" && !token.IsRevoked
				})).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:     "UserNotFound",
			login:    "nonexistentuser",
			password: "password123",
			mockSetup: func(userRepo *MockUserRepository, passwordSvc *MockPasswordService, jwtSvc *MockJWTService, tokenRepo *MockTokenRepository) {
				userRepo.On("FindByLogin", mock.Anything, "nonexistentuser").Return(nil, nil)
			},
			expectedError: domainerrors.ErrInvalidCredentials,
		},
		{
			name:     "InvalidPassword",
			login:    "testuser",
			password: "wrongpassword",
			mockSetup: func(userRepo *MockUserRepository, passwordSvc *MockPasswordService, jwtSvc *MockJWTService, tokenRepo *MockTokenRepository) {
				userRepo.On("FindByLogin", mock.Anything, "testuser").Return(&authmodels.User{
					ID:           userID,
					Login:        "testuser",
					PasswordHash: "hashedpassword",
				}, nil)

				passwordSvc.On("Verify", mock.Anything, "wrongpassword", "hashedpassword").Return(false, nil)
			},
			expectedError: domainerrors.ErrInvalidCredentials,
		},
		{
			name:     "TokenGenerationFailed",
			login:    "testuser",
			password: "password123",
			mockSetup: func(userRepo *MockUserRepository, passwordSvc *MockPasswordService, jwtSvc *MockJWTService, tokenRepo *MockTokenRepository) {
				userRepo.On("FindByLogin", mock.Anything, "testuser").Return(&authmodels.User{
					ID:           userID,
					Login:        "testuser",
					PasswordHash: "hashedpassword",
				}, nil)

				passwordSvc.On("Verify", mock.Anything, "password123", "hashedpassword").Return(true, nil)

				jwtSvc.On("GenerateTokens", mock.Anything, userID, "testuser").Return(nil, errors.New("token error"))
			},
			expectedError: domainerrors.ErrInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := setupTestContext()
			userRepo := new(MockUserRepository)
			tokenRepo := new(MockTokenRepository)
			passwordSvc := new(MockPasswordService)
			jwtSvc := new(MockJWTService)

			passwordSvc.On("GenerateRandom", mock.Anything, mock.Anything).Return("", nil).Maybe()
			jwtSvc.On("GetTokenTTL").Return(time.Hour).Maybe()

			tt.mockSetup(userRepo, passwordSvc, jwtSvc, tokenRepo)

			uc := NewAuthUseCase(userRepo, tokenRepo, passwordSvc, jwtSvc)

			tokenPair, err := uc.Login(ctx, tt.login, tt.password)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, tokenPair)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tokenPair)
				assert.Equal(t, "access-token", tokenPair.AccessToken)
				assert.Equal(t, "refresh-token", tokenPair.RefreshToken)
			}

			userRepo.AssertExpectations(t)
			passwordSvc.AssertExpectations(t)
			jwtSvc.AssertExpectations(t)
			tokenRepo.AssertExpectations(t)
		})
	}
}

func TestValidateToken(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		token          string
		mockSetup      func(*MockJWTService, *MockUserRepository)
		expectedUserID uuid.UUID
		expectedError  error
	}{
		{
			name:  "Success",
			token: "valid-token",
			mockSetup: func(jwtSvc *MockJWTService, userRepo *MockUserRepository) {
				jwtSvc.On("ValidateToken", mock.Anything, "valid-token").Return(userID, nil)
				userRepo.On("FindByID", mock.Anything, userID).Return(&authmodels.User{ID: userID}, nil)
			},
			expectedUserID: userID,
			expectedError:  nil,
		},
		{
			name:  "InvalidToken",
			token: "invalid-token",
			mockSetup: func(jwtSvc *MockJWTService, userRepo *MockUserRepository) {
				jwtSvc.On("ValidateToken", mock.Anything, "invalid-token").Return(uuid.Nil, errors.New("invalid token"))
			},
			expectedUserID: uuid.Nil,
			expectedError:  domainerrors.ErrInvalidToken,
		},
		{
			name:  "UserNotFound",
			token: "valid-token",
			mockSetup: func(jwtSvc *MockJWTService, userRepo *MockUserRepository) {
				jwtSvc.On("ValidateToken", mock.Anything, "valid-token").Return(userID, nil)
				userRepo.On("FindByID", mock.Anything, userID).Return(nil, nil)
			},
			expectedUserID: uuid.Nil,
			expectedError:  domainerrors.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := setupTestContext()
			userRepo := new(MockUserRepository)
			tokenRepo := new(MockTokenRepository)
			passwordSvc := new(MockPasswordService)
			jwtSvc := new(MockJWTService)

			passwordSvc.On("GenerateRandom", mock.Anything, mock.Anything).Return("", nil).Maybe()
			jwtSvc.On("GetTokenTTL").Return(time.Hour).Maybe()

			tt.mockSetup(jwtSvc, userRepo)

			uc := NewAuthUseCase(userRepo, tokenRepo, passwordSvc, jwtSvc)

			resultUserID, err := uc.ValidateToken(ctx, tt.token)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Equal(t, uuid.Nil, resultUserID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUserID, resultUserID)
			}

			userRepo.AssertExpectations(t)
			jwtSvc.AssertExpectations(t)
		})
	}
}

func TestRefreshToken(t *testing.T) {
	userID := uuid.New()
	expirationTime := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name          string
		token         string
		mockSetup     func(*MockJWTService, *MockTokenRepository, *MockUserRepository)
		expectedError error
	}{
		{
			name:  "Success",
			token: "valid-refresh-token",
			mockSetup: func(jwtSvc *MockJWTService, tokenRepo *MockTokenRepository, userRepo *MockUserRepository) {
				jwtSvc.On("ParseToken", mock.Anything, "valid-refresh-token").Return(map[string]interface{}{"user_id": userID.String()}, nil)

				tokenRepo.On("FindByTokenString", mock.Anything, "valid-refresh-token").Return(&authmodels.Token{
					ID:        uuid.New(),
					UserID:    userID,
					TokenStr:  "valid-refresh-token",
					ExpiresAt: expirationTime,
					IsRevoked: false,
				}, nil)

				userRepo.On("FindByID", mock.Anything, userID).Return(&authmodels.User{
					ID:    userID,
					Login: "testuser",
				}, nil)

				tokenRepo.On("RevokeToken", mock.Anything, "valid-refresh-token").Return(nil)

				jwtSvc.On("GenerateTokens", mock.Anything, userID, "testuser").Return(&authmodels.TokenPair{
					AccessToken:  "new-access-token",
					RefreshToken: "new-refresh-token",
				}, nil)

				jwtSvc.On("GetRefreshTokenTTL").Return(24 * time.Hour)

				tokenRepo.On("Store", mock.Anything, mock.MatchedBy(func(token *authmodels.Token) bool {
					return token.UserID == userID && token.TokenStr == "new-refresh-token" && !token.IsRevoked
				})).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "InvalidToken",
			token: "invalid-token",
			mockSetup: func(jwtSvc *MockJWTService, tokenRepo *MockTokenRepository, userRepo *MockUserRepository) {
				jwtSvc.On("ParseToken", mock.Anything, "invalid-token").Return(nil, errors.New("invalid token"))
			},
			expectedError: domainerrors.ErrInvalidToken,
		},
		{
			name:  "TokenNotFound",
			token: "nonexistent-token",
			mockSetup: func(jwtSvc *MockJWTService, tokenRepo *MockTokenRepository, userRepo *MockUserRepository) {
				jwtSvc.On("ParseToken", mock.Anything, "nonexistent-token").Return(map[string]interface{}{"user_id": userID.String()}, nil)
				tokenRepo.On("FindByTokenString", mock.Anything, "nonexistent-token").Return(nil, nil)
			},
			expectedError: domainerrors.ErrTokenNotFound,
		},
		{
			name:  "RevokedToken",
			token: "revoked-token",
			mockSetup: func(jwtSvc *MockJWTService, tokenRepo *MockTokenRepository, userRepo *MockUserRepository) {
				jwtSvc.On("ParseToken", mock.Anything, "revoked-token").Return(map[string]interface{}{"user_id": userID.String()}, nil)
				tokenRepo.On("FindByTokenString", mock.Anything, "revoked-token").Return(&authmodels.Token{
					ID:        uuid.New(),
					UserID:    userID,
					TokenStr:  "revoked-token",
					ExpiresAt: expirationTime,
					IsRevoked: true,
				}, nil)
			},
			expectedError: domainerrors.ErrTokenRevoked,
		},
		{
			name:  "ExpiredToken",
			token: "expired-token",
			mockSetup: func(jwtSvc *MockJWTService, tokenRepo *MockTokenRepository, userRepo *MockUserRepository) {
				jwtSvc.On("ParseToken", mock.Anything, "expired-token").Return(map[string]interface{}{"user_id": userID.String()}, nil)
				tokenRepo.On("FindByTokenString", mock.Anything, "expired-token").Return(&authmodels.Token{
					ID:        uuid.New(),
					UserID:    userID,
					TokenStr:  "expired-token",
					ExpiresAt: time.Now().Add(-24 * time.Hour),
					IsRevoked: false,
				}, nil)
			},
			expectedError: domainerrors.ErrTokenExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := setupTestContext()
			userRepo := new(MockUserRepository)
			tokenRepo := new(MockTokenRepository)
			passwordSvc := new(MockPasswordService)
			jwtSvc := new(MockJWTService)

			passwordSvc.On("GenerateRandom", mock.Anything, mock.Anything).Return("", nil).Maybe()
			jwtSvc.On("GetTokenTTL").Return(time.Hour).Maybe()

			tt.mockSetup(jwtSvc, tokenRepo, userRepo)

			uc := NewAuthUseCase(userRepo, tokenRepo, passwordSvc, jwtSvc)

			tokenPair, err := uc.RefreshToken(ctx, tt.token)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, tokenPair)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tokenPair)
				assert.Equal(t, "new-access-token", tokenPair.AccessToken)
				assert.Equal(t, "new-refresh-token", tokenPair.RefreshToken)
			}

			userRepo.AssertExpectations(t)
			tokenRepo.AssertExpectations(t)
			jwtSvc.AssertExpectations(t)
		})
	}
}

func TestLogout(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name          string
		token         string
		mockSetup     func(*MockJWTService, *MockTokenRepository)
		expectedError error
	}{
		{
			name:  "Success",
			token: "valid-token",
			mockSetup: func(jwtSvc *MockJWTService, tokenRepo *MockTokenRepository) {
				jwtSvc.On("ParseToken", mock.Anything, "valid-token").Return(map[string]interface{}{"user_id": userID.String()}, nil)
				tokenRepo.On("FindByTokenString", mock.Anything, "valid-token").Return(&authmodels.Token{
					ID:       uuid.New(),
					TokenStr: "valid-token",
				}, nil)
				tokenRepo.On("RevokeToken", mock.Anything, "valid-token").Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "InvalidToken",
			token: "invalid-token",
			mockSetup: func(jwtSvc *MockJWTService, tokenRepo *MockTokenRepository) {
				jwtSvc.On("ParseToken", mock.Anything, "invalid-token").Return(nil, errors.New("invalid token"))
			},
			expectedError: domainerrors.ErrInvalidToken,
		},
		{
			name:  "TokenNotFound",
			token: "nonexistent-token",
			mockSetup: func(jwtSvc *MockJWTService, tokenRepo *MockTokenRepository) {
				jwtSvc.On("ParseToken", mock.Anything, "nonexistent-token").Return(map[string]interface{}{"user_id": userID.String()}, nil)
				tokenRepo.On("FindByTokenString", mock.Anything, "nonexistent-token").Return(nil, nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := setupTestContext()
			userRepo := new(MockUserRepository)
			tokenRepo := new(MockTokenRepository)
			passwordSvc := new(MockPasswordService)
			jwtSvc := new(MockJWTService)

			passwordSvc.On("GenerateRandom", mock.Anything, mock.Anything).Return("", nil).Maybe()
			jwtSvc.On("GetTokenTTL").Return(time.Hour).Maybe()

			tt.mockSetup(jwtSvc, tokenRepo)

			uc := NewAuthUseCase(userRepo, tokenRepo, passwordSvc, jwtSvc)

			err := uc.Logout(ctx, tt.token)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			tokenRepo.AssertExpectations(t)
			jwtSvc.AssertExpectations(t)
		})
	}
}

func TestCleanupExpiredTokens(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*MockTokenRepository)
		expectedError error
	}{
		{
			name: "Success",
			mockSetup: func(tokenRepo *MockTokenRepository) {
				tokenRepo.On("DeleteExpiredTokens", mock.Anything, mock.MatchedBy(func(t time.Time) bool {
					now := time.Now()
					diff := now.Sub(t)
					return diff >= 0 && diff < time.Second
				})).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Error",
			mockSetup: func(tokenRepo *MockTokenRepository) {
				tokenRepo.On("DeleteExpiredTokens", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedError: domainerrors.ErrInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := setupTestContext()
			userRepo := new(MockUserRepository)
			tokenRepo := new(MockTokenRepository)
			passwordSvc := new(MockPasswordService)
			jwtSvc := new(MockJWTService)

			passwordSvc.On("GenerateRandom", mock.Anything, mock.Anything).Return("", nil).Maybe()
			jwtSvc.On("GetTokenTTL").Return(time.Hour).Maybe()

			tt.mockSetup(tokenRepo)

			uc := NewAuthUseCase(userRepo, tokenRepo, passwordSvc, jwtSvc)

			err := uc.CleanupExpiredTokens(ctx)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			tokenRepo.AssertExpectations(t)
		})
	}
}
