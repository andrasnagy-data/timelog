package auth

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/andrasnagy-data/timelog/internal/shared/config"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
)

type (
	servicer interface {
		ValidateCredentials(context.Context, string, string) (*User, error)
		GetUserByID(context.Context, uuid.UUID) (*User, error)
		GetSecretKey() []byte
	}
	service struct {
		config *config.Config
	}
)

func NewAuthService(config *config.Config) servicer {
	return &service{
		config: config,
	}
}

// ValidateCredentials checks username and password
func (s *service) ValidateCredentials(_ context.Context, username, password string) (*User, error) {
	if username != s.config.Username {
		return nil, ErrInvalidCredentials
	}

	err := bcrypt.CompareHashAndPassword([]byte(s.config.PasswordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return &User{
		ID:       uuid.MustParse(s.config.UserId),
		Username: s.config.Username,
	}, nil
}

// GetUserByID returns user by ID (for context validation)
func (s *service) GetUserByID(_ context.Context, userID uuid.UUID) (*User, error) {
	if userID.String() == s.config.UserId {
		return &User{
			ID:       uuid.MustParse(s.config.UserId),
			Username: s.config.Username,
		}, nil
	}
	return nil, errors.New("user not found")
}

// GetSecretKey returns the secret key for cookie encryption
func (s *service) GetSecretKey() []byte {
	key, err := hex.DecodeString(s.config.SecretKey)
	if err != nil {
		panic("Invalid hex secret key: " + err.Error())
	}
	return key
}
