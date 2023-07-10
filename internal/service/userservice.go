package service

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/distuurbia/firstTaskArtyom/internal/config"
	"github.com/distuurbia/firstTaskArtyom/internal/model"
	"github.com/google/uuid"
)

// UserRepository is an interface that defines the methods on entities.
type UserRepository interface {
	SignUpUser(ctx context.Context, user *model.User) error
	GetByLogin(ctx context.Context, login string) ([]byte, uuid.UUID, bool, error)
	AddToken(ctx context.Context, id uuid.UUID, token string) error
	RefreshToken(ctx context.Context, id uuid.UUID) (string, error)
}

// UserEntity represents the service that interacts with the repository.
type UserEntity struct {
	urpc UserRepository
	cfg  *config.Config
}

// NewUserEntity creates a new instance of the service.
func NewUserEntity(urpc UserRepository, cfg *config.Config) *UserEntity {
	return &UserEntity{
		urpc: urpc,
		cfg:  cfg,
	}
}

// SignUpUser creates a new user.
func (u *UserEntity) SignUpUser(ctx context.Context, user *model.User) (aT, rT string, er error) {
	var err error
	user.Password, err = HashPassword(user.Password)
	if err != nil {
		return "", "", fmt.Errorf("UserEntity-SignUpUser-HashPassword: error in hashing password: %w", err)
	}
	err = u.urpc.SignUpUser(ctx, user)
	if err != nil {
		return "", "", fmt.Errorf("UserEntity-SignUpUser: error in method s.rpc.signupuser: %w", err)
	}
	accessToken, refreshToken, err := GenerateTokens(user.ID, user.Admin, u.cfg)
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
func (u *UserEntity) GetByLogin(ctx context.Context, login string, password []byte) (aT, rT string, er error) {
	hash, id, admin, err := u.urpc.GetByLogin(ctx, login)
	if err != nil {
		return "", "", fmt.Errorf("UserEntity-GetByLogin: error in method u.urpc.GetByLogin: %w", err)
	}
	verify := CheckPasswordHash(password, hash)
	if !verify {
		return "", "", fmt.Errorf("UserEntity-GetByLogin-CheckPasswordHash: passwords not matched: %w", err)
	}
	accessToken, refreshToken, err := GenerateTokens(id, admin, u.cfg)
	if err != nil {
		return "", "", fmt.Errorf("UserEntity-GetByLogin-GenerateTokens: error in generating refresh token: %w", err)
	}
	err = u.AddToken(ctx, id, []byte(refreshToken))
	if err != nil {
		return "", "", fmt.Errorf("UserEntity-GetByLogin-AddToken: error in method u.AddToken: %w", err)
	}
	return accessToken, refreshToken, nil
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
func (u *UserEntity) RefreshToken(ctx context.Context, accessToken, refreshToken string) (aT, rT string, er error) {
	accessID, accessAdmin, err := CheckTokenValidity(accessToken, u.cfg.AccessTokenSignature)
	if err != nil {
		return "", "", fmt.Errorf("UserEntity-RefreshToken-CheckTokenValidity: token expired: %w", err)
	}
	refreshID, refreshAdmin, err := CheckTokenValidity(refreshToken, u.cfg.RefreshTokenSignature)
	if err != nil {
		return "", "", fmt.Errorf("UserEntity-RefreshToken-CheckTokenValidity: token expired: %w", err)
	}
	if accessID != refreshID {
		return "", "", fmt.Errorf("id not matched")
	}
	if accessAdmin != refreshAdmin {
		return "", "", fmt.Errorf("roles not matched")
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
	accessToken, refreshToken, err = GenerateTokens(accessID, accessAdmin, u.cfg)
	if err != nil {
		return "", "", fmt.Errorf("UserEntity-RefreshToken-GenerateTokens: error in generating refresh token: %w", err)
	}
	err = u.AddToken(ctx, accessID, []byte(refreshToken))
	if err != nil {
		return "", "", fmt.Errorf("UserEntity-RefreshToken: error in method u.AddToken: %w", err)
	}
	return accessToken, refreshToken, nil
}
