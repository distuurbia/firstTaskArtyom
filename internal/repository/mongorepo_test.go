package repository

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/distuurbia/firstTaskArtyom/internal/model"
	"github.com/google/uuid"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mrpc *MongoRepository

func SetupMongo() (*mongo.Client, func(), error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, fmt.Errorf("could not construct pool: %w", err)
	}
	resource, err := pool.Run("mongo", "latest", []string{
		"MONGO_INITDB_ROOT_USERNAME=artemmdb",
		"MONGO_INITDB_ROOT_PASSWORD=artemmdb",
		"MONGO_INITDB_DATABASE=mdb"})
	if err != nil {
		return nil, nil, fmt.Errorf("could not start resource: %w", err)
	}

	port := resource.GetPort("27017/tcp")
	mongoURL := fmt.Sprintf("mongodb://artemmdb:artemmdb@localhost:%s", port)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect mongoDB: %w", err)
	}
	cleanup := func() {
		client.Disconnect(context.Background())
		pool.Purge(resource)
	}
	return client, cleanup, nil
}

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
	defer recoveryFunction()
	var err error
	testModel.ID, err = uuid.Parse("Some UUID")
	if err != nil {
		t.Fatalf("failed to parse id")
	}
	err = mrpc.Delete(context.Background(), testModel.ID)
	require.NoError(t, err)
}

func TestNotValidIDMongo(t *testing.T) {
	defer recoveryFunction()
	var err error
	testModel.ID, err = uuid.Parse("1")
	if err != nil {
		t.Fatalf("failed to parse id")
	}
	err = mrpc.Update(context.Background(), &testModel)
	require.NoError(t, err)
}

func recoveryFunction() {
	if recoveryMessage := recover(); recoveryMessage != nil {
		fmt.Println("Recovered. Error:\n", recoveryMessage)
	}
}
