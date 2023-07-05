package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"

	"github.com/distuurbia/firstTaskArtyom/internal/config"
	"github.com/distuurbia/firstTaskArtyom/internal/model"
	"github.com/caarlos0/env"
	"github.com/google/uuid"
)

// UserRepository is an interface that defines the methods on entities.
type UserRepository interface {
	SignUpUser(ctx context.Context, user *model.User) error
	GetByLogin(ctx context.Context, login string) ([]byte, uuid.UUID, error)
	AddToken(ctx context.Context, id uuid.UUID, token string) error
	RefreshToken(ctx context.Context, id uuid.UUID) (string, error)
}

// UserEntity represents the service that interacts with the repository.
type UserEntity struct {
	urpc UserRepository
}

// NewUserEntity creates a new instance of the service.
func NewUserEntity(urpc UserRepository) *UserEntity {
	return &UserEntity{
		urpc: urpc,
	}
}

// SignUpUser creates a new user.
func (u *UserEntity) SignUpUser(ctx context.Context, user *model.User) (aT, rT string, e error) {
	var err error
	user.Password, err = HashPassword(user.Password)
	if err != nil {
		return "", "", fmt.Errorf("UserEntity-SignUpUser-HashPassword: error in hashing password: %w", err)
	}
	err = u.urpc.SignUpUser(ctx, user)
	if err != nil {
		return "", "", fmt.Errorf("UserEntity-SignUpUser: error in method s.rpc.signupuser: %w", err)
	}
	accessToken, refreshToken, err := GenerateTokens(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("UserEntity-SignUpUser-GenerateTokens: error in generating refresh token: %w", err)
	}
	err = u.AddToken(ctx, user.ID, []byte(refreshToken))
	if err != nil {
		return "", "", fmt.Errorf("UserEntity-SignUpUser: error in method u.AddToken: %w", err)
	}
	return accessToken, refreshToken, nil
}

// GetByLogin compare passwords and return id.
func (u *UserEntity) GetByLogin(ctx context.Context, login string, refreshToken, password []byte) (bool, uuid.UUID, error) {
	hash, id, err := u.urpc.GetByLogin(ctx, login)
	if err != nil {
		return false, uuid.Nil, fmt.Errorf("UserEntity-GetByLogin: error in method u.urpc.GetByLogin: %w", err)
	}
	verify := CheckPasswordHash(password, hash)
	if !verify {
		return verify, uuid.Nil, fmt.Errorf("UserEntity-GetByLogin-CheckPasswordHash: passwords not matched: %w", err)
	}
	err = u.AddToken(ctx, id, refreshToken)
	if err != nil {
		return false, uuid.Nil, fmt.Errorf("UserEntity-GetByLogin: error in method u.AddToken: %w", err)
	}
	return verify, id, nil
}

// AddToken insert RefreshToken into database.
func (u *UserEntity) AddToken(ctx context.Context, id uuid.UUID, token []byte) error {
	hashToken := sha256.Sum256(token)
	var err error
	token, err = HashPassword(hashToken[:])
	if err != nil {
		return fmt.Errorf("UserEntity-AddToken: error in hashing password: %w", err)
	}
	return u.urpc.AddToken(ctx, id, string(token))
}

// RefreshToken checks incoming tokens for validity.
func (u *UserEntity) RefreshToken(ctx context.Context, accessToken, refreshToken string) (aT, rT string, e error) {
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
	accessID, err := CheckTokenValidity(accessToken, cfg.AccessTokenSignature)
	if err != nil {
		return "", "", fmt.Errorf("UserEntity-RefreshToken-CheckTokenValidity: token expired: %w", err)
	}
	refreshID, err := CheckTokenValidity(refreshToken, cfg.RefreshTokenSignature)
	if err != nil {
		return "", "", fmt.Errorf("UserEntity-RefreshToken-CheckTokenValidity: token expired: %w", err)
	}
	if accessID != refreshID {
		return "", "", fmt.Errorf("id not matched")
	}
	hash, err := u.urpc.RefreshToken(ctx, refreshID)
	if err != nil {
		return "", "", fmt.Errorf("UserEntity-RefreshToken: error in method u.urpc.RefreshToken: %w", err)
	}
	sum := sha256.Sum256([]byte(refreshToken))
	verified := CheckPasswordHash(sum[:], []byte(hash))
	if !verified {
		return "", "", fmt.Errorf("UserEntity-RefreshToken-CheckPasswordHash: error - refreshToken invalid")
	}
	accessToken, refreshToken, err = GenerateTokens(accessID)
	if err != nil {
		return "", "", fmt.Errorf("UserEntity-RefreshToken-GenerateTokens: error in generating refresh token: %w", err)
	}
	err = u.AddToken(ctx, accessID, []byte(refreshToken))
	if err != nil {
		return "", "", fmt.Errorf("UserEntity-RefreshToken: error in method u.AddToken: %w", err)
	}
	return accessToken, refreshToken, nil
}