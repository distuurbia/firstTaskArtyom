package repository

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/distuurbia/firstTaskArtyom/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/require"
)

var rpc *PgRepository

var testModel = model.Car{
	ID:             uuid.New(),
	Brand:          RandBrand(),
	ProductionYear: RandProductionYear(),
	IsRunning:      RandBool(),
}

func SetupPostgres() (*pgxpool.Pool, func(), error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, fmt.Errorf("could not construct pool: %w", err)
	}
	resource, err := pool.Run("postgres", "latest", []string{
		"POSTGRES_USER=artempg",
		"POSTGRESQL_PASSWORD=artempg",
		"POSTGRES_DB=db"})
	if err != nil {
		return nil, nil, fmt.Errorf("could not start resource: %w", err)
	}
	dbURL := fmt.Sprintf("postgres://artempg:artempg@localhost:%s/db", resource.GetPort("5432"))
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse dbURL: %w", err)
	}
	dbpool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect pgxpool: %w", err)
	}
	cleanup := func() {
		dbpool.Close()
		pool.Purge(resource)
	}
	return dbpool, cleanup, nil
}

func TestMain(m *testing.M) {
	dbpool, cleanupPgx, err := SetupPostgres()
	if err != nil {
		fmt.Println("Could not construct the pool: ", err)
		cleanupPgx()
		os.Exit(1)
	}
	rpc = NewPgRepository(dbpool)

	client, cleanupMongo, err := SetupMongo()
	if err != nil {
		fmt.Println(err)
		cleanupMongo()
		os.Exit(1)
	}
	mrpc = NewMongoRepository(client)

	exitCode := m.Run()

	cleanupPgx()
	cleanupMongo()

	os.Exit(exitCode)
}

/*Generating random fields*/

func RandBrand() string {
	currencies := []string{"Toyota", "Honda", "Mazda", "Skoda", "Lada"}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}

func RandInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func RandProductionYear() int64 {
	return RandInt(1980, 2023)
}

func RandBool() bool {
	return rand.Intn(2) == 1
}

func TestCreate(t *testing.T) {
	err := rpc.Create(context.Background(), &testModel)
	require.NoError(t, err)

	var newCar *model.Car
	newCar, err = rpc.Get(context.Background(), testModel.ID)
	require.NoError(t, err)
	require.Equal(t, testModel.ID, newCar.ID)
	require.Equal(t, testModel.Brand, newCar.Brand)
	require.Equal(t, testModel.ProductionYear, newCar.ProductionYear)
	require.Equal(t, testModel.IsRunning, newCar.IsRunning)
}

func TestUpdate(t *testing.T) {
	testModel.Brand = "UpdatedTestBrand"
	testModel.ProductionYear--
	testModel.IsRunning = false
	err := rpc.Update(context.Background(), &testModel)
	require.NoError(t, err)

	var updCar *model.Car
	updCar, err = rpc.Get(context.Background(), testModel.ID)
	require.NoError(t, err)
	require.Equal(t, testModel.ID, updCar.ID)
	require.Equal(t, testModel.Brand, updCar.Brand)
	require.Equal(t, testModel.ProductionYear, updCar.ProductionYear)
	require.Equal(t, testModel.IsRunning, updCar.IsRunning)
}

func TestGet(t *testing.T) {
	getCar, err := rpc.Get(context.Background(), testModel.ID)
	require.NoError(t, err)
	require.Equal(t, testModel.ID, getCar.ID)
	require.Equal(t, testModel.Brand, getCar.Brand)
	require.Equal(t, testModel.ProductionYear, getCar.ProductionYear)
	require.Equal(t, testModel.IsRunning, getCar.IsRunning)
}

func TestDelete(t *testing.T) {
	err := rpc.Delete(context.Background(), testModel.ID)
	require.NoError(t, err)

	var count int
	err = rpc.pool.QueryRow(context.Background(), "SELECT COUNT(brand) FROM car WHERE id = $1", testModel.ID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func TestGetAll(t *testing.T) {
	cars, err := rpc.GetAll(context.Background())
	require.NoError(t, err)

	var count int
	err = rpc.pool.QueryRow(context.Background(), "Select Count(Brand) From Car").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, len(cars), count)
}

func TestStrangeData(t *testing.T) {
	newage, _ := strconv.Atoi("age")
	var newbool bool

	testModel.Brand = strconv.Itoa(0)
	testModel.ProductionYear = int64(newage)
	testModel.IsRunning = newbool

	err := rpc.Create(context.Background(), &testModel)
	if err != nil {
		t.Fatalf("There was an error in creating")
	}

	strangeCar, err := rpc.Get(context.Background(), testModel.ID)
	require.NoError(t, err)
	require.Equal(t, strangeCar.Brand, "0")
	require.Equal(t, strangeCar.ProductionYear, int64(0))
	require.Equal(t, strangeCar.IsRunning, false)

	err = rpc.Delete(context.Background(), testModel.ID)
	require.NoError(t, err)
}

func TestDeleteByFakeID(t *testing.T) {
	defer recoveryFunction()
	var err error
	testModel.ID, err = uuid.Parse("Some UUID")
	if err != nil {
		t.Fatalf("failed to parse id")
	}
	err = rpc.Delete(context.Background(), testModel.ID)
	require.NoError(t, err)
}

func TestNotValidID(t *testing.T) {
	defer recoveryFunction()
	var err error
	testModel.ID, err = uuid.Parse("1")
	if err != nil {
		t.Fatalf("failed to parse id")
	}
	err = rpc.Update(context.Background(), &testModel)
	require.NoError(t, err)
}
