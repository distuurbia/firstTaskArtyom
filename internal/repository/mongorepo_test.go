package repository

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/distuurbia/firstTaskArtyom/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)


func TestCreateMongo(t *testing.T) {
	err := mrpc.Create(context.Background(), &testModel)
	require.NoError(t, err)

	var newCar *model.Car
	newCar, err = mrpc.Get(context.Background(), testModel.ID)
	require.NoError(t, err)
	require.NotEmpty(t, newCar)
	require.NotZero(t, newCar.ID)
	require.Equal(t, testModel.ID, newCar.ID)
	require.Equal(t, testModel.Brand, newCar.Brand)
	require.Equal(t, testModel.ProductionYear, newCar.ProductionYear)
	require.Equal(t, testModel.IsRunning, newCar.IsRunning)
}

func TestUpdateMongo(t *testing.T) {
	testModel.Brand = "UpdatedTestBrand"
	testModel.ProductionYear = 1999
	testModel.IsRunning = false

	err := mrpc.Update(context.Background(), &testModel)
	require.NoError(t, err)

	var updCar *model.Car
	updCar, err = mrpc.Get(context.Background(), testModel.ID)
	require.NoError(t, err)
	require.Equal(t, testModel.ID, updCar.ID)
	require.Equal(t, testModel.Brand, updCar.Brand)
	require.Equal(t, testModel.ProductionYear, updCar.ProductionYear)
	require.Equal(t, testModel.IsRunning, updCar.IsRunning)
}

func TestGetMongo(t *testing.T) {
	var getcar *model.Car
	getcar, err := mrpc.Get(context.Background(), testModel.ID)
	require.NoError(t, err)
	require.Equal(t, testModel.ID, getcar.ID)
	require.Equal(t, testModel.Brand, getcar.Brand)
	require.Equal(t, testModel.ProductionYear, getcar.ProductionYear)
	require.Equal(t, testModel.IsRunning, getcar.IsRunning)
}

func TestDeleteMongo(t *testing.T) {
	err := mrpc.Delete(context.Background(), testModel.ID)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	_, err = mrpc.Get(context.Background(), testModel.ID)
	require.Error(t, err)
}

func TestGetAllMongo(t *testing.T) {
	getcar, err := mrpc.GetAll(context.Background())
	require.NoError(t, err)

	collection := mrpc.client.Database("mdb").Collection("car")
	count, err := collection.CountDocuments(context.Background(), bson.M{})
	require.NoError(t, err)
	require.Equal(t, len(getcar), int(count))
}

func TestStrangeDataMongo(t *testing.T) {
	newage, _ := strconv.Atoi("age")
	var newbool bool

	testModel.Brand = strconv.Itoa(0)
	testModel.ProductionYear = int64(newage)
	testModel.IsRunning = newbool

	err := mrpc.Create(context.Background(), &testModel)
	if err != nil {
		t.Fatalf("There was an error in creating")
	}

	strangeCar, err := mrpc.Get(context.Background(), testModel.ID)
	require.NoError(t, err)
	require.Equal(t, strangeCar.Brand, "0")
	require.Equal(t, strangeCar.ProductionYear, int64(0))
	require.Equal(t, strangeCar.IsRunning, false)
}

func TestDeleteByFakeIDMongo(t *testing.T) {
	testModel.ID, _ = uuid.Parse("Some UUID")
	err := mrpc.Delete(context.Background(), testModel.ID)
	require.ErrorIs(t, err, mongo.ErrNoDocuments)
}

func TestNotValidIDMongo(t *testing.T) {
	var err error
	testModel.ID, _ = uuid.Parse("1")
	err = mrpc.Update(context.Background(), &testModel)
	require.ErrorIs(t, err, mongo.ErrNoDocuments)
}

func recoveryFunction() {
	if recoveryMessage := recover(); recoveryMessage != nil {
		fmt.Println("Recovered. Error:\n", recoveryMessage)
	}
}
