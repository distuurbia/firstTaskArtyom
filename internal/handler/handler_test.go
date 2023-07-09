package handler

import (
	"context"
	"os"
	"testing"

	"github.com/distuurbia/firstTaskArtyom/internal/handler/mocks"
	"github.com/distuurbia/firstTaskArtyom/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	testModel = model.Car{
		ID:             uuid.New(),
		Brand:          "HandlBrand",
		ProductionYear: 2002,
		IsRunning:      false}
	serv *mocks.CarService
)

func TestMain(m *testing.M) {
	serv = new(mocks.CarService)
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestCreateCar(t *testing.T) {
	serv.On("Create", mock.Anything, mock.AnythingOfType("*model.Car")).
		Return(nil).
		Once()

	err := serv.Create(context.Background(), &testModel)
	require.Nil(t, err)

	serv.AssertExpectations(t)
}

func TestGetCar(t *testing.T) {
	serv.On("Get", mock.Anything, mock.AnythingOfType("uuid.UUID")).
		Return(&testModel, nil).
		Once()
	car, err := serv.Get(context.Background(), testModel.ID)
	require.Nil(t, err)
	require.Equal(t, testModel.ID, car.ID)
	require.Equal(t, testModel.Brand, car.Brand)
	require.Equal(t, testModel.ProductionYear, car.ProductionYear)
	require.Equal(t, testModel.IsRunning, car.IsRunning)

	serv.AssertExpectations(t)
}

func TestDeleteCar(t *testing.T) {
	serv.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID")).
		Return(nil).
		Once()
	err := serv.Delete(context.Background(), testModel.ID)
	require.Nil(t, err)

	serv.AssertExpectations(t)
}

func TestUpdatecar(t *testing.T) {
	serv.On("Update", mock.Anything, mock.AnythingOfType("*model.Car")).
		Return(nil).
		Once()
	err := serv.Update(context.Background(), &testModel)
	require.Nil(t, err)

	serv.AssertExpectations(t)
}

func TestGetAllCar(t *testing.T) {
	expectedCars := []*model.Car{
		{
			ID:             uuid.New(),
			Brand:          "handlBrand1",
			ProductionYear: 1977,
			IsRunning:      true,
		},
		{
			ID:             uuid.New(),
			Brand:          "handlBrand2",
			ProductionYear: 1988,
			IsRunning:      false,
		},
	}
	serv.On("GetAll", mock.Anything).
		Return(expectedCars, nil).
		Once()

	cars, err := serv.GetAll(context.Background())
	require.Nil(t, err)
	require.Equal(t, len(expectedCars), len(cars))

	serv.AssertExpectations(t)
}
