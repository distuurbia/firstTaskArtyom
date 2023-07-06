package service

import (
	"fmt"
	"log"
	"time"

	"github.com/distuurbia/firstTaskArtyom/internal/config"
	"github.com/caarlos0/env"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	// AccessTime - time of life access token.
	AccessTime = 15 * time.Minute

	// RefreshTime - time of life refresh token.
	RefreshTime = 72 * time.Hour
)

// HashPassword hashes the password to be written into the database.
func HashPassword(password []byte) ([]byte, error) {
	const cost = 14
	bytes, err := bcrypt.GenerateFromPassword(password, cost)
	return bytes, err
}

// CheckPasswordHash compared hash.
func CheckPasswordHash(password, hash []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, password)
	return err == nil
}

// GenerateTokens created Tokens (Access and Refresh).
func GenerateTokens(id uuid.UUID, admin bool) (aT, rT string, e error) {
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
	accessTokenClaims := jwt.MapClaims{
		"admin": admin, 
		"id":  id.String(),
		"exp": time.Now().Add(AccessTime).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(cfg.AccessTokenSignature))
	if err != nil {
		return "", "", fmt.Errorf("error in generating access token: %w", err)
	}
	refreshTokenClaims := jwt.MapClaims{
		"admin": admin, 
		"id":  id.String(),
		"exp": time.Now().Add(RefreshTime).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(cfg.RefreshTokenSignature))
	if err != nil {
		return "", "", fmt.Errorf("error in generating refresh token: %w", err)
	}
	return accessTokenString, refreshTokenString, nil
}

// CheckTokenValidity returns id by claims.
func CheckTokenValidity(token, signature string) (uuid.UUID, bool, error) {
	var tokenID uuid.UUID
	var admin bool
	thisToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(signature), nil
	})
	if err != nil {
		return uuid.Nil, admin, fmt.Errorf("invalid token: %w", err)
	}
	if !thisToken.Valid {
		return uuid.Nil, admin, fmt.Errorf("invalid token")
	}
	if claims, ok := thisToken.Claims.(jwt.MapClaims); ok && thisToken.Valid {
		tokenID, err = uuid.Parse(claims["id"].(string))
		if err != nil {
			return uuid.Nil, admin, fmt.Errorf("failed to parse id")
		}
		admin = claims["admin"].(bool)
	}
	return tokenID,  admin, nil
}
