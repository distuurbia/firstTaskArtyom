// Package repository provides a PostgreSQL repository implementation.
package repository

import (
	"context"
	"fmt"

	"github.com/distuurbia/firstTaskArtyom/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgRepository represents the PostgreSQL repository implementation.
type PgRepository struct {
	pool *pgxpool.Pool
}

// NewPgRepository creates and returns a new instance of PgRepository, using the provided pgxpool.Pool.
func NewPgRepository(pool *pgxpool.Pool) *PgRepository {
	return &PgRepository{
		pool: pool,
	}
}

// Create creates a new car record in the database.
func (p *PgRepository) Create(ctx context.Context, car *model.Car) error {
	_, err := p.pool.Exec(ctx, "INSERT INTO car (id, brand, productionyear, isrunning) VALUES ($1, $2, $3, $4)", car.ID, car.Brand, car.ProductionYear, car.IsRunning)
	if err != nil {
		return fmt.Errorf("PgRepository-Create: error in method r.pool.Exec(): %w", err)
	}
	return nil
}

// Get retrieves a car record from the database based on the provided ID.
func (p *PgRepository) Get(ctx context.Context, id uuid.UUID) (*model.Car, error) {
	var car model.Car
	err := p.pool.QueryRow(ctx, "SELECT id, brand, productionyear, isrunning FROM car WHERE id = $1", id).Scan(&car.ID, &car.Brand, &car.ProductionYear, &car.IsRunning)
	if err != nil {
		return nil, fmt.Errorf("PgRepository-Get: error in method r.pool.QuerryRow(): %w", err)
	}
	return &car, nil
}

// Delete removes a car record from the database based on the provided ID.
func (p *PgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := p.pool.Exec(ctx, "DELETE FROM car WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("PgRepository-Delete: error in method r.pool.Exec(): %w", err)
	}
	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// Update updates a car record in the database.
func (p *PgRepository) Update(ctx context.Context, car *model.Car) error {
	res, err := p.pool.Exec(ctx, "UPDATE car SET brand = $1, productionyear = $2, isrunning = $3 WHERE id = $4", car.Brand, car.ProductionYear, car.IsRunning, car.ID)
	if err != nil {
		return fmt.Errorf("PgRepository-Update: error in method r.pool.Exec(): %w", err)
	}
	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// GetAll retrieves all car records from the database.
func (p *PgRepository) GetAll(ctx context.Context) ([]*model.Car, error) {
	var cars []*model.Car
	rows, err := p.pool.Query(ctx, "SELECT id, brand, productionyear, isrunning FROM car")
	if err != nil {
		return nil, fmt.Errorf("PgRepository-GetAll: error in method r.pool.Query(): %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var car model.Car
		err := rows.Scan(&car.ID, &car.Brand, &car.ProductionYear, &car.IsRunning)
		if err != nil {
			return nil, fmt.Errorf("PgRepository-GetAll: error in method rows.Scan(): %w", err)
		}
		cars = append(cars, &car)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("PgRepository-GetAll: error iterating rows: %w", err)
	}
	return cars, nil
}
