package repository

import (
	"context"
	"fmt"

	"github.com/distuurbia/firstTaskArtyom/internal/model"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// SignUpUser creates a new user record in the database.
func (m *MongoRepository) SignUpUser(ctx context.Context, user *model.User) error {
	collection := m.client.Database("mdb").Collection("users")
	count, err := collection.CountDocuments(ctx, bson.M{"login": user.Login})
	if err != nil {
		return fmt.Errorf("MongoRepository-SignUpUser: error in CountDocuments: %w", err)
	}
	if count != 0 {
		return fmt.Errorf("MongoRepository-SignUpUser: the login is occupied by another user")
	}
	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("MongoRepository-SignUpUser: error in InsertOne: %w", err)
	}
	return nil
}

// GetByLogin retrieves the user's password from the database by login.
func (m *MongoRepository) GetByLogin(ctx context.Context, login string) ([]byte, uuid.UUID, bool, error) {
	collection := m.client.Database("mdb").Collection("users")
	var result struct {
		ID       uuid.UUID        `bson:"_id"`
		Password primitive.Binary `bson:"password"`
		Admin	 bool			  `bson:"admin"`
	}
	err := collection.FindOne(ctx, bson.M{"login": login}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, uuid.Nil, false, fmt.Errorf("MongoRepository-GetByLogin: user not found")
		}
		return nil, uuid.Nil, false, fmt.Errorf("MongoRepository-GetByLogin: error in FindOne: %w", err)
	}
	passwordCopy := make([]byte, len(result.Password.Data))
	copy(passwordCopy, result.Password.Data)
	return passwordCopy, result.ID, result.Admin, nil
}

// AddToken adds a token to the user's record in the database.
func (m *MongoRepository) AddToken(ctx context.Context, id uuid.UUID, token string) error {
	collection := m.client.Database("mdb").Collection("users")
	_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"refreshtoken": token}})
	if err != nil {
		return fmt.Errorf("MongoRepository-AddToken: error in UpdateOne: %w", err)
	}
	return nil
}

// RefreshToken returns refresh token by id.
func (m *MongoRepository) RefreshToken(ctx context.Context, id uuid.UUID) (string, error) {
	collection := m.client.Database("mdb").Collection("users")
	filter := bson.M{"_id": id}
	var result struct {
		RefreshToken string `bson:"refreshtoken"`
	}
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("MongoRepository-RefreshToken: user not found")
		}
		return "", fmt.Errorf("MongoRepository-RefreshToken: error in method collection.FindOne(): %w", err)
	}
	return result.RefreshToken, nil
}
