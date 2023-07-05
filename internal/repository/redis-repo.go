package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/distuurbia/firstTaskArtyom/internal/model"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// RedisRepository represents the Redis repository implementation.
type RedisRepository struct {
	client *redis.Client
}

// NewRedisRepository creates and returns a new instance of RedisRepository, using the provided redis.Client.
func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{
		client: client,
	}
}

// SetCache stores the provided car object in the Redis cache.
func (r *RedisRepository) SetCache(ctx context.Context, car *model.Car) error {
	carJSON, err := json.Marshal(car)
	if err != nil {
		return fmt.Errorf("RedisRepository-Set: error in method json.Marshal(): %w", err)
	}
	r.client.HSet(ctx, "car", car.ID.String(), carJSON)
	return nil
}

// GetCache retrieves the car object with the specified ID from the Redis cache.
func (r *RedisRepository) GetCache(ctx context.Context, id uuid.UUID) (*model.Car, error) {
	carJSON, err := r.client.HGet(ctx, "car", id.String()).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, err
		}
		return nil, fmt.Errorf("RedisRepository-Get: error in method s.client.HGet(): %w", err)
	}
	var car model.Car
	err = json.Unmarshal([]byte(carJSON), &car)
	if err != nil {
		return nil, fmt.Errorf("RedisRepository-Get: error in method json.Unmarshal(): %w", err)
	}
	return &car, nil
}

// DeleteCache removes the car object with the specified ID from the Redis cache.
func (r *RedisRepository) DeleteCache(ctx context.Context, id uuid.UUID) error {
	_, err := r.client.HDel(ctx, "car", id.String()).Result()
	if err != nil {
		return fmt.Errorf("RedisRepository-Delete: error in method s.client.HDel(): %w", err)
	}
	return nil
}

// // AddToStream adds the provided car object to the Redis stream.
// func (r *RedisRepository) AddToStream(ctx context.Context, car *model.Car) error {
// 	carJSON, err := json.Marshal(car)
// 	if err != nil {
// 		return fmt.Errorf("RedisRepository-AddToStream: error in method json.Marshal(): %w", err)
// 	}
// 	streamData := redis.XAddArgs{
// 		Stream: "carstream",
// 		Values: map[string]interface{}{
// 			"Data":  string(carJSON),
// 			"CarID": car.ID.String(),
// 		},
// 	}
// 	_, err = r.client.XAdd(ctx, &streamData).Result()
// 	if err != nil {
// 		return fmt.Errorf("RedisRepository-AddToStream: error in method rS.client.XAdd(): %w", err)
// 	}
// 	return nil
// }

// // GetFromStream retrieves the car object with the specified ID from the Redis stream.
// func (r *RedisRepository) GetFromStream(ctx context.Context, id uuid.UUID) (*model.Car, error) {
// 	streamData := redis.XReadArgs{
// 		Streams: []string{"carstream", "0"},
// 		Count:   0,
// 		Block:   0,
// 	}
// 	results, err := r.client.XRead(ctx, &streamData).Result()
// 	if err != nil {
// 		return nil, fmt.Errorf("RedisRepository-GetFromStream: error in method rS.client.XRead(): %w", err)
// 	}
// 	if len(results) == 0 || len(results[0].Messages) == 0 {
// 		return nil, redis.Nil
// 	}
// 	var car model.Car
// 	var carJSON string
// 	for _, msg := range results[0].Messages {
// 		if msg.Values["CarID"] == id.String() {
// 			carJSON = msg.Values["Data"].(string)
// 			err = json.Unmarshal([]byte(carJSON), &car)
// 			if err != nil {
// 				return nil, fmt.Errorf("RedisRepository-GetFromStream: error in method json.Unmarshal(): %w", err)
// 			}
// 			return &car, nil
// 		}
// 	}
// 	return nil, nil
// }

// // DeleteFromStream removes the car object with the specified ID from the Redis stream.
// func (r *RedisRepository) DeleteFromStream(ctx context.Context, id uuid.UUID) error {
// 	streamData := redis.XReadArgs{
// 		Streams: []string{"carstream", "0"},
// 		Count:   0,
// 		Block:   0,
// 	}
// 	results, err := r.client.XRead(ctx, &streamData).Result()
// 	if err != nil {
// 		return fmt.Errorf("RedisRepository-DeleteFromStream: error in method rS.client.XRead(): %w", err)
// 	}
// 	if len(results) == 0 || len(results[0].Messages) == 0 {
// 		return redis.Nil
// 	}
// 	var msgID string
// 	for _, msg := range results[0].Messages {
// 		if msg.Values["CarID"] == id.String() {
// 			msgID = msg.ID
// 		}
// 	}
// 	_, err = r.client.XDel(ctx, "carstream", msgID).Result()
// 	if err != nil {
// 		return fmt.Errorf("RedisRepository-DeleteFromStream: error in method rS.client.XDel(): %w", err)
// 	}
// 	return nil
// }
