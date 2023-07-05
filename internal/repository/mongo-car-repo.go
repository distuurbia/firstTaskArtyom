// Package repository provides a MongoDB repository implementation.
package repository

import (
	"context"
	"fmt"

	"github.com/distuurbia/firstTaskArtyom/internal/model"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoRepository represents the MongoDB repository.
type MongoRepository struct {
	client *mongo.Client
}

// NewMongoRepository creates and returns a new instance of MongoRepository, using the passed mongo.Client.
func NewMongoRepository(client *mongo.Client) *MongoRepository {
	return &MongoRepository{
		client: client,
	}
}

// Create inserts a new car record into the MongoDB collection.
func (m *MongoRepository) Create(ctx context.Context, car *model.Car) error {
	collection := m.client.Database("mdb").Collection("car")
	_, err := collection.InsertOne(ctx, car)
	if err != nil {
		return fmt.Errorf("MongoRepository-Create: error in method collection.InsertOne(): %w", err)
	}
	return nil
}

// Get retrieves a car record from the MongoDB collection by ID.
func (m *MongoRepository) Get(ctx context.Context, id uuid.UUID) (*model.Car, error) {
	collection := m.client.Database("mdb").Collection("car")
	filter := bson.M{"_id": id}
	var car model.Car
	err := collection.FindOne(ctx, filter).Decode(&car)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("MongoRepository-Get: car not found")
		}
		return nil, fmt.Errorf("MongoRepository-Get: error in method collection.FindOne(): %w", err)
	}
	return &car, nil
}

// Delete removes a car record from the MongoDB collection by ID.
func (m *MongoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	collection := m.client.Database("mdb").Collection("car")
	filter := bson.M{"_id": id}
	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("MongoRepository-Delete: error in method collection.DeleteOne(): %w", err)
	}
	return nil
}

// Update updates an existing car record in the MongoDB collection.
func (m *MongoRepository) Update(ctx context.Context, car *model.Car) error {
	collection := m.client.Database("mdb").Collection("car")
	filter := bson.M{"_id": car.ID}
	update := bson.M{
		"$set": bson.M{
			"brand":          car.Brand,
			"productionyear": car.ProductionYear,
			"isrunning":      car.IsRunning,
		},
	}
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("MongoRepository-Update: error in method collection.UpdateOne(): %w", err)
	}
	return nil
}

// GetAll retrieves all car records from the MongoDB collection.
func (m *MongoRepository) GetAll(ctx context.Context) ([]*model.Car, error) {
	collection := m.client.Database("mdb").Collection("car")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("MongoRepository-GetAll: error in method collection.Find(): %w", err)
	}
	defer func() {
		err := cursor.Close(ctx)
		if err != nil {
			fmt.Printf("MongoRepository-GetAll: Failed to close cursor: %v", err)
		}
	}()
	var cars []*model.Car
	for cursor.Next(ctx) {
		var car model.Car
		if err := cursor.Decode(&car); err != nil {
			return nil, fmt.Errorf("MongoRepository-GetAll: error decoding car: %w", err)
		}
		cars = append(cars, &car)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("MongoRepository-GetAll: error in cursor: %w", err)
	}
	return cars, nil
}
