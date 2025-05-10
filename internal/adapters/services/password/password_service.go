package password

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/service/password"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	MinPasswordLength = 8
	DefaultCost       = bcrypt.DefaultCost
	lowerChars        = "abcdefghijklmnopqrstuvwxyz"
	upperChars        = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numChars          = "0123456789"
	specChars         = "!@#$%^&*()-_=+[]{}|;:,.<>?/"
)

var (
	allChars             = lowerChars + upperChars + numChars + specChars
	ErrEmptyPassword     = errors.New("password cannot be empty")
	ErrPasswordTooShort  = errors.New("password is too short")
	ErrInvalidHashFormat = errors.New("invalid password hash format")
	ErrGeneratingRandom  = errors.New("error generating random password")
)

type Service struct {
	cost int
}

var _ password.Service = (*Service)(nil)

func NewService(cost int) *Service {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = DefaultCost
	}
	return &Service{cost: cost}
}

func (s *Service) Hash(ctx context.Context, password string) (string, error) {
	const op = "PasswordService.Hash"

	if password == "" {
		return "", ErrEmptyPassword
	}

	if len(password) < MinPasswordLength {
		return "", ErrPasswordTooShort
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		logger.Error(ctx, nil, "failed to hash password", zap.String("op", op), zap.Error(err))
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

func (s *Service) Verify(ctx context.Context, password, hashedPassword string) (bool, error) {
	if password == "" {
		return false, ErrEmptyPassword
	}

	if hashedPassword == "" {
		return false, ErrInvalidHashFormat
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, fmt.Errorf("error verifying password: %w", err)
	}

	return true, nil
}

func (s *Service) GenerateRandom(ctx context.Context, length int) (string, error) {
	if length < MinPasswordLength {
		length = MinPasswordLength
	}

	password := make([]byte, length)
	charSets := []struct {
		set string
		len *big.Int
	}{
		{lowerChars, big.NewInt(int64(len(lowerChars)))},
		{upperChars, big.NewInt(int64(len(upperChars)))},
		{numChars, big.NewInt(int64(len(numChars)))},
		{specChars, big.NewInt(int64(len(specChars)))},
	}

	allCharsLen := big.NewInt(int64(len(allChars)))

	for i := 0; i < length; i++ {
		var randIdx *big.Int
		var err error

		if i < 4 {
			randIdx, err = rand.Int(rand.Reader, charSets[i].len)
			if err != nil {
				return "", fmt.Errorf("%w: %s", ErrGeneratingRandom, err.Error())
			}
			password[i] = charSets[i].set[randIdx.Int64()]
		} else {
			randIdx, err = rand.Int(rand.Reader, allCharsLen)
			if err != nil {
				return "", fmt.Errorf("%w: %s", ErrGeneratingRandom, err.Error())
			}
			password[i] = allChars[randIdx.Int64()]
		}
	}

	for i := length - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", fmt.Errorf("%w: %s", ErrGeneratingRandom, err.Error())
		}
		password[i], password[j.Int64()] = password[j.Int64()], password[i]
	}

	return string(password), nil
}
