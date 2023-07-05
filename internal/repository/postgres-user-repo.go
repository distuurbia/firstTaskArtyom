package repository

import (
	"context"
	"fmt"

	"github.com/distuurbia/firstTaskArtyom/internal/model"
	"github.com/google/uuid"
)

// SignUpUser creates a new user record in the database.
func (p *PgRepository) SignUpUser(ctx context.Context, user *model.User) error {
	var count int
	err := p.pool.QueryRow(ctx, "SELECT COUNT(id) FROM users WHERE login = $1", user.Login).Scan(&count)
	if err != nil {
		return fmt.Errorf("PgRepository-SignUpUser: error in method r.pool.QuerryRow(): %w", err)
	}
	if count != 0 {
		return fmt.Errorf("PgRepository-SignUpUser: the login is occupied by another user")
	}
	_, err = p.pool.Exec(ctx, "INSERT INTO users (id, login, password) VALUES ($1, $2, $3)", user.ID, user.Login, user.Password)
	if err != nil {
		return fmt.Errorf("PgRepository-SignUpUser: error in method r.pool.Exec(): %w", err)
	}
	return nil
}

// GetByLogin get password and id of user.
func (p *PgRepository) GetByLogin(ctx context.Context, login string) ([]byte, uuid.UUID, error) {
	var id uuid.UUID
	var password []byte
	err := p.pool.QueryRow(ctx, "SELECT password, id FROM users WHERE login = $1", login).Scan(&password, &id)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("PgRepository-GetByLOgin: error in method r.pool.QuerryRow(): %w", err)
	}
	return password, id, nil
}

// AddToken adds a token to the user's record in the database.
func (p *PgRepository) AddToken(ctx context.Context, id uuid.UUID, token string) error {
	_, err := p.pool.Exec(ctx, "UPDATE users SET refreshtoken = $1 WHERE id = $2", token, id)
	if err != nil {
		return fmt.Errorf("PgRepository-AddToken: error in method r.pool.Exec(): %w", err)
	}
	return nil
}

// RefreshToken returns refresh token by id.
func (p *PgRepository) RefreshToken(ctx context.Context, id uuid.UUID) (string, error) {
	var refreshToken string
	err := p.pool.QueryRow(ctx, "SELECT refreshtoken FROM users WHERE id = $1", id).Scan(&refreshToken)
	if err != nil {
		return "", fmt.Errorf("PgRepository-GetByLOgin: error in method r.pool.QuerryRow(): %w", err)
	}
	return refreshToken, nil
}
