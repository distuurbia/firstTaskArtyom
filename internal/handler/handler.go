// Package handler provides the HTTP request handlers for the application's endpoints.
package handler

import (
	"bufio"
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/distuurbia/firstTaskArtyom/internal/model"
	"github.com/distuurbia/firstTaskArtyom/proto_services"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
)

// CarService is an interface that defines the methods on Car entity.
type CarService interface {
	Create(ctx context.Context, car *model.Car) error
	Get(ctx context.Context, id uuid.UUID) (*model.Car, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, car *model.Car) error
	GetAll(ctx context.Context) ([]*model.Car, error)
}

// UserService is an interface that defines the methods on User entity.
type UserService interface {
	SignUpUser(ctx context.Context, user *model.User) (string, string, error)
	GetByLogin(ctx context.Context, login string, password []byte) (string, string, error)
	RefreshToken(ctx context.Context, accessToken string, refreshToken string) (string, string, error)
}

// GRPCHandler is responsible for handling gRPC requests related to entities.
type GRPCHandler struct {
	carService  CarService
	userService UserService
	validate    *validator.Validate
	proto_services.UnimplementedCarServiceServer
	proto_services.UnimplementedUserServiceServer
	proto_services.UnimplementedImageServiceServer
}

// NewGRPCHandler creates a new instance of the GRPCHandler struct.
func NewGRPCHandler(carService CarService, userService UserService, v *validator.Validate) *GRPCHandler {
	return &GRPCHandler{
		carService:  carService,
		userService: userService,
		validate:    v,
	}
}

// GetCar handles the GET request to retrieve a car by its ID.
func (h *GRPCHandler) GetCar(ctx context.Context, req *proto_services.GetCarRequest) (*proto_services.GetCarResponse, error) {
	id, err := uuid.Parse(req.ID.Value)
	if err != nil {
		log.Errorf("failed to parse error %v", err)
		return &proto_services.GetCarResponse{}, err
	}
	err = h.validate.VarCtx(ctx, id.String(), "required,uuid")
	if err != nil {
		log.Errorf("failed to validate error: %v", err)
		return &proto_services.GetCarResponse{}, err
	}
	car, err := h.carService.Get(ctx, id)
	if err != nil {
		log.WithField(
			"ID", id,
		).Errorf("failed to get data: %v", err)
		return &proto_services.GetCarResponse{}, err
	}
	protoCar := proto_services.Car{
		ID:             &proto_services.UUID{Value: car.ID.String()},
		Brand:          car.Brand,
		ProductionYear: car.ProductionYear,
		IsRunning:      car.IsRunning,
	}
	return &proto_services.GetCarResponse{Car: &protoCar}, nil
}

// CreateCar handles the POST request to create a new car.
func (h *GRPCHandler) CreateCar(ctx context.Context, req *proto_services.CreateCarRequest) (*proto_services.CreateCarResponse, error) {
	newCar := model.Car{
		ID:             uuid.New(),
		Brand:          req.Car.Brand,
		ProductionYear: req.Car.ProductionYear,
		IsRunning:      req.Car.IsRunning,
	}
	err := h.validate.StructCtx(ctx, newCar)
	if err != nil {
		log.Errorf("failed to validate error: %v", err)
		return &proto_services.CreateCarResponse{}, err
	}
	err = h.carService.Create(ctx, &newCar)
	if err != nil {
		log.WithFields(log.Fields{
			"Brand":          newCar.Brand,
			"PodusctionYear": newCar.ProductionYear,
			"isRunning":      newCar.IsRunning,
		}).Errorf("failed to get data: %v", err)
		return &proto_services.CreateCarResponse{}, err
	}
	protoCar := proto_services.Car{
		ID:             &proto_services.UUID{Value: newCar.ID.String()},
		Brand:          newCar.Brand,
		ProductionYear: newCar.ProductionYear,
		IsRunning:      newCar.IsRunning,
	}
	return &proto_services.CreateCarResponse{Car: &protoCar}, nil
}

// UpdateCar handles the PUT request to update an existing car.
func (h *GRPCHandler) UpdateCar(ctx context.Context, req *proto_services.UpdateCarRequest) (*proto_services.UpdateCarResponse, error) {
	id, err := uuid.Parse(req.Car.ID.Value)
	if err != nil {
		log.Errorf("failed to parse error %v", err)
		return &proto_services.UpdateCarResponse{}, err
	}
	car := model.Car{
		ID:             id,
		Brand:          req.Car.Brand,
		ProductionYear: req.Car.ProductionYear,
		IsRunning:      req.Car.IsRunning,
	}
	err = h.validate.StructCtx(ctx, car)
	if err != nil {
		log.Errorf("failed to validate error %v", err)
		return &proto_services.UpdateCarResponse{}, err
	}
	err = h.carService.Update(ctx, &car)
	if err != nil {
		log.WithFields(log.Fields{
			"ID":             car.ID,
			"Brand":          car.Brand,
			"PodusctionYear": car.ProductionYear,
			"isRunning":      car.IsRunning,
		}).Errorf("failed to get data: %v", err)
		return &proto_services.UpdateCarResponse{}, err
	}
	protoCar := proto_services.Car{
		ID:             &proto_services.UUID{Value: car.ID.String()},
		Brand:          car.Brand,
		ProductionYear: car.ProductionYear,
		IsRunning:      car.IsRunning,
	}
	return &proto_services.UpdateCarResponse{Car: &protoCar}, nil
}

// DeleteCar handles the DELETE request to delete a car by its ID.
func (h *GRPCHandler) DeleteCar(ctx context.Context, req *proto_services.DeleteCarRequest) (*proto_services.DeleteCarResponse, error) {
	id, err := uuid.Parse(req.ID.Value)
	if err != nil {
		log.Errorf("failed to parse error %v", err)
		return &proto_services.DeleteCarResponse{}, err
	}
	err = h.validate.VarCtx(ctx, id.String(), "required,uuid")
	if err != nil {
		log.Errorf("failed to validate error: %v", err)
		return &proto_services.DeleteCarResponse{}, err
	}
	err = h.carService.Delete(ctx, id)
	if err != nil {
		log.WithField(
			"ID", id,
		).Errorf("failed to get data: %v", err)
		return &proto_services.DeleteCarResponse{}, err
	}
	return &proto_services.DeleteCarResponse{ID: &proto_services.UUID{Value: id.String()}}, nil
}

// GetAllCars handles the GET request to retrieve all cars.
func (h *GRPCHandler) GetAllCars(ctx context.Context, _ *proto_services.GetAllCarsRequest) (*proto_services.GetAllCarsResponse, error) {
	cars, err := h.carService.GetAll(ctx)
	if err != nil {
		log.Errorf("failed to get all cars error: %v", err)
		return &proto_services.GetAllCarsResponse{}, err
	}
	expectedSize := len(cars)
	var protoCars = make([]*proto_services.Car, 0, expectedSize)
	for _, car := range cars {
		protoCars = append(protoCars, &proto_services.Car{
			ID:             &proto_services.UUID{Value: car.ID.String()},
			Brand:          car.Brand,
			ProductionYear: car.ProductionYear,
			IsRunning:      car.IsRunning,
		})
	}
	return &proto_services.GetAllCarsResponse{Cars: protoCars}, nil
}

// InputData is a struct for binding login and password.
type InputData struct {
	Login    string `json:"login" form:"login"`
	Password string `json:"password" form:"password"`
}

// SignUpUser handles the POST request to create a new user.
func (h *GRPCHandler) SignUpUser(ctx context.Context, req *proto_services.SignUpUserRequest) (*proto_services.SignUpUserResponse, error) {
	var newUser model.User
	newUser.ID = uuid.New()
	newUser.Login = req.Login
	newUser.Password = []byte(req.Password)
	err := h.validate.StructCtx(ctx, newUser)
	if err != nil {
		log.Errorf("failed to validate error: %v", err)
		return &proto_services.SignUpUserResponse{}, err
	}
	accessToken, refreshToken, err := h.userService.SignUpUser(ctx, &newUser)
	if err != nil {
		log.WithFields(log.Fields{
			"Login":         newUser.Login,
			"Password":      newUser.Password,
			"Access Toke":   accessToken,
			"Refresh Token": refreshToken,
			"Admin":         newUser.Admin,
		}).Errorf("failed to get data: %v", err)
		return &proto_services.SignUpUserResponse{}, err
	}
	return &proto_services.SignUpUserResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

// SignUpAdmin handles the POST request to create a new admin.
func (h *GRPCHandler) SignUpAdmin(ctx context.Context, req *proto_services.SignUpAdminRequest) (*proto_services.SignUpAdminResponse, error) {
	var newUser model.User
	newUser.ID = uuid.New()

	newUser.Login = req.Login
	newUser.Password = []byte(req.Password)
	newUser.Admin = true
	err := h.validate.StructCtx(ctx, newUser)
	if err != nil {
		log.Errorf("failed to validate error: %v", err)
		return &proto_services.SignUpAdminResponse{}, err
	}
	accessToken, refreshToken, err := h.userService.SignUpUser(ctx, &newUser)
	if err != nil {
		log.WithFields(log.Fields{
			"Login":         newUser.Login,
			"Password":      newUser.Password,
			"Access Token":  accessToken,
			"Refresh Token": refreshToken,
			"Admin":         newUser.Admin,
		}).Errorf("failed to get data: %v", err)
		return &proto_services.SignUpAdminResponse{}, err
	}
	return &proto_services.SignUpAdminResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

// GetByLogin checked password.
func (h *GRPCHandler) GetByLogin(ctx context.Context, req *proto_services.GetByLoginRequest) (*proto_services.GetByLoginResponse, error) {
	var user model.User
	user.Login = req.Login
	user.Password = []byte(req.Password)
	err := h.validate.StructCtx(ctx, user)
	if err != nil {
		log.Errorf("failed to validate error: %v", err)
		return &proto_services.GetByLoginResponse{}, err
	}
	accessToken, refreshToken, err := h.userService.GetByLogin(ctx, user.Login, user.Password)
	if err != nil {
		log.WithFields(log.Fields{
			"Login":    user.Login,
			"Password": user.Password,
		}).Errorf("failed to get data: %v", err)
		return &proto_services.GetByLoginResponse{}, err
	}
	return &proto_services.GetByLoginResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

// InputTokens is a struct for binding access and refresh tokens.
type InputTokens struct {
	AccessToken  string `json:"accessToken" form:"accessToken"`
	RefreshToken string `json:"refreshToken" form:"refreshToken"`
}

// RefreshToken refreshed tokens by tokens.
func (h *GRPCHandler) RefreshToken(ctx context.Context, req *proto_services.RefreshTokenRequest) (*proto_services.RefreshTokenResponse, error) {
	accessToken, refreshToken, err := h.userService.RefreshToken(ctx, req.AccessToken, req.RefreshToken)
	if err != nil {
		log.WithFields(log.Fields{
			"Access Toke":   req.AccessToken,
			"Refresh Token": req.RefreshToken,
		}).Errorf("failed to get data: %v", err)
		return &proto_services.RefreshTokenResponse{}, err
	}
	return &proto_services.RefreshTokenResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

// DownloadImage downloads image from given path
func (h *GRPCHandler) DownloadImage(_ context.Context, req *proto_services.DownloadImageRequest) (*proto_services.DownloadImageResponse, error) {
	imgname := req.ImgName
	imgpath := filepath.Join("images", imgname)
	cleanPath := filepath.Clean(imgpath)
	file, err := os.Open(cleanPath)
	if err != nil {
		log.Errorf("failed to open file error: %v", err)
		return &proto_services.DownloadImageResponse{}, err
	}
	defer func() {
		errClose := file.Close()
		if errClose != nil {
			log.Errorf("failed to close file error: %v", errClose)
		}
	}()

	stat, err := file.Stat()
	if err != nil {
		log.Errorf("failed to check file stat error: %v", err)
		return &proto_services.DownloadImageResponse{}, err
	}

	img := make([]byte, stat.Size())
	_, err = bufio.NewReader(file).Read(img)
	if err != nil && err != io.EOF {
		log.Errorf("failed to check file stat error: %v", err)
		return &proto_services.DownloadImageResponse{}, err
	}
	protoImage := &proto_services.DownloadImageResponse{Img: img}
	return protoImage, nil
}

// UploadImage uploads image from given path
func (h *GRPCHandler) UploadImage(_ context.Context, req *proto_services.UploadImageRequest) (*proto_services.UploadImageResponse, error) {
	imgname := req.ImgName
	imgpath := filepath.Join("images", imgname)
	cleanPath := filepath.Clean(imgpath)
	file, err := os.Open(cleanPath)
	if err != nil {
		log.Errorf("failed to open file error: %v", err)
		return &proto_services.UploadImageResponse{}, err
	}
	defer func() {
		errClose := file.Close()
		if errClose != nil {
			log.Errorf("failed to close file error: %v", errClose)
		}
	}()
	dst, err := os.Create(filepath.Join("images", "upload", "uploadedSmile.png"))
	if err != nil {
		log.Errorf("error: %v", err)
		return &proto_services.UploadImageResponse{}, err
	}
	defer func() {
		errClose := dst.Close()
		if errClose != nil {
			log.Errorf("error: %v", errClose)
		}
	}()
	_, err = io.Copy(dst, file)
	if err != nil {
		log.Errorf("error: %v", err)
		return &proto_services.UploadImageResponse{}, err
	}
	return &proto_services.UploadImageResponse{}, nil
}
