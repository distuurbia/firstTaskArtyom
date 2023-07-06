// Package model provides the data models used in the application.
package model

import (
	"github.com/google/uuid"
)

// Car represents a car entity.
type Car struct {
	ID             uuid.UUID `json:"id,omitempty" bson:"_id"`
	Brand          string    `json:"brand" validate:"required"`
	ProductionYear int64     `json:"productionyear" validate:"gte=1950,lte=2023"`
	IsRunning      bool      `json:"isrunning"`
}

// User represents a user entity.
type User struct {
	ID           uuid.UUID `json:"id" bson:"_id"`
	Login        string    `json:"login" validate:"required,min=4,max=20"`
	Password     []byte    `json:"password" validate:"required,min=4"`
	RefreshToken []byte    `json:"refreshtoken"`
	Admin        bool      `json:"admin"`
}
