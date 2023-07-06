package repository

import (
	"context"
	"strconv"
	"testing"

	"github.com/distuurbia/firstTaskArtyom/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

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
	testModel.ID, _ = uuid.Parse("Some UUID")
	err = rpc.Delete(context.Background(), testModel.ID)
	require.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestNotValidID(t *testing.T) {
	defer recoveryFunction()
	var err error
	testModel.ID, _ = uuid.Parse("1")
	err = rpc.Update(context.Background(), &testModel)
	require.ErrorIs(t, err, pgx.ErrNoRows)
}
