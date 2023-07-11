package handler

import (
	"context"
	"os"
	"testing"

	"github.com/distuurbia/firstTaskArtyom/internal/handler/mocks"
	"github.com/distuurbia/firstTaskArtyom/internal/model"
	"github.com/distuurbia/firstTaskArtyom/proto_services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gopkg.in/go-playground/validator.v9"
)

var (
	testModel = model.Car{
		ID:             uuid.New(),
		Brand:          "HandlBrand",
		ProductionYear: 2002,
		IsRunning:      false}
	testProtoCar = proto_services.Car{
		ID:             &proto_services.UUID{Value: uuid.New().String()},
		Brand:          "HandlBrand",
		ProductionYear: 2002,
		IsRunning:      false}
)

func TestMain(m *testing.M) {
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestCreateCar(t *testing.T) {
	servCar := new(mocks.CarService)
	servCar.On("Create", mock.Anything, mock.AnythingOfType("*model.Car")).
	Return(nil).
	Once()
	GRPCHandl := NewGRPCHandler(servCar, nil, validator.New())
	protoResponse, err := GRPCHandl.CreateCar(context.Background(), &proto_services.CreateCarRequest{Car: &testProtoCar})
	require.NoError(t, err)
	require.Equal(t, protoResponse.Car.Brand, testProtoCar.Brand)
	require.Equal(t, protoResponse.Car.IsRunning, testProtoCar.IsRunning)
	require.Equal(t, protoResponse.Car.ProductionYear, testProtoCar.ProductionYear)
	servCar.AssertExpectations(t)
}

func TestGetCar(t *testing.T) {
	servCar := new(mocks.CarService)
	servCar.On("Get", mock.Anything, mock.AnythingOfType("uuid.UUID")).
	Return(&testModel, nil).
	Once()
	GRPCHandl := NewGRPCHandler(servCar, nil, validator.New())
	protoResponse, err := GRPCHandl.GetCar(context.Background(), &proto_services.GetCarRequest{ID: testProtoCar.ID})
	require.NoError(t, err)
	require.Equal(t, protoResponse.Car.Brand, testProtoCar.Brand)
	require.Equal(t, protoResponse.Car.IsRunning, testProtoCar.IsRunning)
	require.Equal(t, protoResponse.Car.ProductionYear, testProtoCar.ProductionYear)
	servCar.AssertExpectations(t)
}

func TestDeleteCar(t *testing.T) {
	servCar := new(mocks.CarService)
	servCar.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID")).
		Return(nil).
		Once()
	GRPCHandl := NewGRPCHandler(servCar, nil, validator.New())
	protoResponse, err := GRPCHandl.DeleteCar(context.Background(), &proto_services.DeleteCarRequest{ID: testProtoCar.ID})
	require.NoError(t, err)
	require.Equal(t, protoResponse.ID, testProtoCar.ID)
	servCar.AssertExpectations(t)
}

func TestUpdatecar(t *testing.T) {
	servCar := new(mocks.CarService)
	servCar.On("Update", mock.Anything, mock.AnythingOfType("*model.Car")).
		Return(nil).
		Once()
	GRPCHandl := NewGRPCHandler(servCar, nil, validator.New())
	protoResponse, err := GRPCHandl.UpdateCar(context.Background(), &proto_services.UpdateCarRequest{Car: &testProtoCar})
	require.NoError(t, err)
	require.Equal(t, protoResponse.Car.ID, testProtoCar.ID)
	require.Equal(t, protoResponse.Car.Brand, testProtoCar.Brand)
	require.Equal(t, protoResponse.Car.IsRunning, testProtoCar.IsRunning)
	require.Equal(t, protoResponse.Car.ProductionYear, testProtoCar.ProductionYear)
	servCar.AssertExpectations(t)
}

func TestGetAllCar(t *testing.T) {
	servCar := new(mocks.CarService)
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
	servCar.On("GetAll", mock.Anything).
		Return(expectedCars, nil).
		Once()
	GRPCHandl := NewGRPCHandler(servCar, nil, validator.New())
	protoResponse, err := GRPCHandl.GetAllCars(context.Background(), &proto_services.GetAllCarsRequest{})
	require.NoError(t, err)
	require.Equal(t, len(expectedCars), len(protoResponse.Cars))
	servCar.AssertExpectations(t)
}

func TestSignUpUser(t *testing.T){
	servUser := new(mocks.UserService)
	servUser.On("SignUpUser", mock.Anything, mock.AnythingOfType("*model.User")).
	Return("accessToken", "refreshToken", nil).
	Once()
	GRPCHandl := NewGRPCHandler(nil, servUser, validator.New())
	protoResponse, err := GRPCHandl.SignUpUser(context.Background(), &proto_services.SignUpUserRequest{Login: "testUser", Password: "testUser"})
	require.NoError(t, err)
	require.Equal(t, protoResponse.AccessToken, "accessToken")
	require.Equal(t, protoResponse.RefreshToken, "refreshToken")
	servUser.AssertExpectations(t)
}

func TestSignUpAdmin(t *testing.T){
	servUser := new(mocks.UserService)
	servUser.On("SignUpUser", mock.Anything, mock.AnythingOfType("*model.User")).
	Return("accessToken", "refreshToken", nil).
	Once()
	GRPCHandl := NewGRPCHandler(nil, servUser, validator.New())
	protoResponse, err := GRPCHandl.SignUpAdmin(context.Background(), &proto_services.SignUpAdminRequest{Login: "testUser", Password: "testUser"})
	require.NoError(t, err)
	require.Equal(t, protoResponse.AccessToken, "accessToken")
	require.Equal(t, protoResponse.RefreshToken, "refreshToken")
	servUser.AssertExpectations(t)
}

func TestGetByLogin(t *testing.T){
	servUser := new(mocks.UserService)
	servUser.On("GetByLogin", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
	Return("accessToken", "refreshToken", nil).
	Once()
	GRPCHandl := NewGRPCHandler(nil, servUser, validator.New())
	protoResponse, err := GRPCHandl.GetByLogin(context.Background(), &proto_services.GetByLoginRequest{Login: "testUser", Password: "testUser"})
	require.NoError(t, err)
	require.Equal(t, protoResponse.AccessToken, "accessToken")
	require.Equal(t, protoResponse.RefreshToken, "refreshToken")
	servUser.AssertExpectations(t)
}

func TestRefresh(t *testing.T){
	servUser := new(mocks.UserService)
	servUser.On("RefreshToken", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
	Return("accessToken", "refreshToken", nil).
	Once()
	GRPCHandl := NewGRPCHandler(nil, servUser, validator.New())
	protoResponse, err := GRPCHandl.RefreshToken(context.Background(), &proto_services.RefreshTokenRequest{AccessToken: "testAccess", RefreshToken: "testRefresh"})
	require.NoError(t, err)
	require.Equal(t, protoResponse.AccessToken, "accessToken")
	require.Equal(t, protoResponse.RefreshToken, "refreshToken")
	servUser.AssertExpectations(t)
}