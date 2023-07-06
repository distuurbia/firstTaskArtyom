// Package service provides the business logic and services for the application.
package service

import (
	"context"
	"fmt"

	"github.com/distuurbia/firstTaskArtyom/internal/model"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// CarRepository is an interface that defines the methods on entities.
type CarRepository interface {
	Create(ctx context.Context, car *model.Car) error
	Get(ctx context.Context, id uuid.UUID) (*model.Car, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, car *model.Car) error
	GetAll(ctx context.Context) ([]*model.Car, error)
}

// RedisCarRepository is an interface that defines the redis methods on entities.
type RedisCarRepository interface {
	GetCache(ctx context.Context, id uuid.UUID) (*model.Car, error)
	SetCache(ctx context.Context, car *model.Car) error
	DeleteCache(ctx context.Context, id uuid.UUID) error
}

// CarEntity represents the service that interacts with the repository.
type CarEntity struct {
	rpc    CarRepository
	rdsRep RedisCarRepository
}

// NewCarEntity creates a new instance of the service.
func NewCarEntity(rpc CarRepository, rdsRep RedisCarRepository) *CarEntity {
	return &CarEntity{
		rpc:    rpc,
		rdsRep: rdsRep,
	}
}

// Create creates a new car.
func (s *CarEntity) Create(ctx context.Context, car *model.Car) error {
	err := s.rpc.Create(ctx, car)
	if err != nil {
		return fmt.Errorf("CarEntity-Create: error in method s.rpc.Create: %w", err)
	}
	err = s.rdsRep.SetCache(ctx, car)
	if err != nil {
		return fmt.Errorf("CarEntity-Create: error in method s.rdsRep.SetCache: %w", err)
	}
	return nil
}

// Update updates an existing car.
func (s *CarEntity) Update(ctx context.Context, car *model.Car) error {
	_ = s.rdsRep.DeleteCache(ctx, car.ID)
	// if err != nil {
	// 	return fmt.Errorf("CarEntity-Update: error in method s.rpc.DeleteCache: %w", err)
	// }
	_ = s.rdsRep.SetCache(ctx, car)
	// if err != nil {
	// 	return fmt.Errorf("CarEntity-Update: error in method s.rdsRep.SetCache: %w", err)
	// }
	err := s.rpc.Update(ctx, car)
	if err != nil {
		return fmt.Errorf("CarEntity-Update: error in method s.rpc.Update: %w", err)
	}
	return nil
}

// Get retrieves a car by its ID.
func (s *CarEntity) Get(ctx context.Context, id uuid.UUID) (*model.Car, error) {
	car, err := s.rdsRep.GetCache(ctx, id)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("CarEntity-Get: error in method s.rpc.GetCache: %w", err)
	}
	if car == nil {
		car, err = s.rpc.Get(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("CarEntity-Get: error in method s.rpc.Get: %w", err)
		}
		_ = s.rdsRep.SetCache(ctx, car)
		// if err != nil {
		// 	return nil, fmt.Errorf("CarEntity-Get: error in method s.rdsRep.SetCache: %w", err)
		// }
	}
	return car, nil
}

// Delete deletes a car by its ID.
func (s *CarEntity) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.rpc.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("CarEntity-Delete: error in method s.rpc.Delete: %w", err)
	}
	_ = s.rdsRep.DeleteCache(ctx, id)
	// if err != nil {
	// 	return fmt.Errorf("CarEntity-Delete: error in method s.rdsRep.DeleteCache: %w", err)
	// }
	return nil
}

// GetAll retrieves all cars.
func (s *CarEntity) GetAll(ctx context.Context) ([]*model.Car, error) {
	return s.rpc.GetAll(ctx)
}
