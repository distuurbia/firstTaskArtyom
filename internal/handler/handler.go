// Package handler provides the HTTP request handlers for the application's endpoints.
package handler

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/distuurbia/firstTaskArtyom/internal/model"
	"github.com/distuurbia/firstTaskArtyom/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gopkg.in/go-playground/validator.v9"

	log "github.com/sirupsen/logrus"
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
	GetByLogin(ctx context.Context, login string, refreshToken []byte, passw []byte) (bool, uuid.UUID, error)
	AddToken(ctx context.Context, id uuid.UUID, token []byte) error
	RefreshToken(ctx context.Context, accessToken string, refreshToken string) (string, string, error)
}

// Handler is responsible for handling HTTP requests related to entities.
type Handler struct {
	carService  CarService
	userService UserService
	validate    *validator.Validate
}

// NewHandler creates a new instance of the Handler struct.
func NewHandler(carService CarService, userService UserService, v *validator.Validate) *Handler {
	return &Handler{
		carService:  carService,
		userService: userService,
		validate:    v,
	}
}

// Get handles the GET request to retrieve a car by its ID.
// @Summary Get
// @Security ApiKeyAuth
// @Description Get car
// @ID get-car
// @Tags methods
// @Accept json
// @Produce json
// @Param id path string true "car"
// @Success 201 {object} model.Car
// @Failure 400 {object} error
// @Router /car/{id} [get]
func (h *Handler) Get(c echo.Context) error {
	id := c.Param("id")
	err := h.validate.VarCtx(c.Request().Context(), id, "required,uuid")
	if err != nil {
		log.Errorf("error: %v", err)
		return c.JSON(http.StatusBadRequest, "Not valid id or field id is empty")
	}
	idUUID, err := uuid.Parse(id)
	if err != nil {
		log.Errorf("error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Handler-Get: failed to parse id")
	}
	car, err := h.carService.Get(c.Request().Context(), idUUID)
	if err != nil {
		log.WithField(
			"ID", id,
		).Errorf("failed to get data: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Handler-Get: challenge id must have uuid format")
	}
	return c.JSON(http.StatusOK, car)
}

// Create handles the POST request to create a new car.
// @Summary Create
// @Security ApiKeyAuth
// @Description Create a new car
// @ID create-car
// @Tags methods
// @Accept json
// @Produce json
// @Param input body model.Car true "car"
// @Success 201 {object} model.Car
// @Failure 400 {object} error
// @Router /car [post]
func (h *Handler) Create(c echo.Context) error {
	var newCar model.Car
	newCar.ID = uuid.New()
	err := c.Bind(&newCar)
	if err != nil {
		log.Errorf("error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Handler-Create: filling car error")
	}
	err = h.validate.StructCtx(c.Request().Context(), newCar)
	if err != nil {
		log.Errorf("error: %v", err)
		return c.JSON(http.StatusBadRequest, "Not valid data")
	}
	err = h.carService.Create(c.Request().Context(), &newCar)
	if err != nil {
		log.WithFields(log.Fields{
			"Brand":          newCar.Brand,
			"PodusctionYear": newCar.ProductionYear,
			"isRunning":      newCar.IsRunning,
		}).Errorf("failed to get data: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Handler-Create: error in method create()")
	}
	return c.JSON(http.StatusCreated, newCar)
}

// Update handles the PUT request to update an existing car.
// @Summary Update
// @Security ApiKeyAuth
// @Description Update car
// @ID update-car
// @Tags methods
// @Accept json
// @Produce json
// @Param input body model.Car true "car"
// @Success 201 {object} model.Car
// @Failure 400 {object} error
// @Router /car [put]
func (h *Handler) Update(c echo.Context) error {
	var car model.Car
	err := c.Bind(&car)
	if err != nil {
		log.Infof("error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Handler-Update: filling car error")
	}
	err = h.validate.StructCtx(c.Request().Context(), car)
	if err != nil {
		log.Errorf("error: %v", err)
		return c.JSON(http.StatusBadRequest, "Not valid data")
	}
	err = h.carService.Update(c.Request().Context(), &car)
	if err != nil {
		log.WithFields(log.Fields{
			"ID":             car.ID,
			"Brand":          car.Brand,
			"PodusctionYear": car.ProductionYear,
			"isRunning":      car.IsRunning,
		}).Errorf("failed to get data: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Handler-Update: error in method update()")
	}
	return c.JSON(http.StatusOK, car)
}

// Delete handles the DELETE request to delete a car by its ID.
// @Summary Delete
// @Security ApiKeyAuth
// @Description Delete car
// @ID delete-car
// @Tags methods
// @Accept json
// @Produce json
// @Param id path string true "car"
// @Success 201 {object} string
// @Failure 400 {object} error
// @Router /car/{id} [delete]
func (h *Handler) Delete(c echo.Context) error {
	id := c.Param("id")
	err := h.validate.VarCtx(c.Request().Context(), id, "required,uuid")
	if err != nil {
		log.Errorf("error: %v", err)
		return c.JSON(http.StatusBadRequest, "Not valid id or field id is empty")
	}
	idUUID, err := uuid.Parse(id)
	if err != nil {
		log.Errorf("error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Handler-Delete: failed to parse id")
	}
	err = h.carService.Delete(c.Request().Context(), idUUID)
	if err != nil {
		log.WithField(
			"ID", id,
		).Errorf("failed to get data: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Handler-Delete: challenge id must have uuid format")
	}
	return c.String(http.StatusOK, fmt.Sprintf("car with ID %s has been deleted", id))
}

// GetAll handles the GET request to retrieve all cars.
// @Summary GetAll
// @Security ApiKeyAuth
// @Description Get All car
// @ID getall-car
// @Tags methods
// @Accept json
// @Produce json
// @Success 201 {object} model.Car
// @Failure 400 {object} error
// @Router /car [get]
func (h *Handler) GetAll(c echo.Context) error {
	cars, err := h.carService.GetAll(c.Request().Context())
	if err != nil {
		log.Errorf("error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Handler-GetAll: failed to get all cars")
	}
	return c.JSON(http.StatusOK, cars)
}

// InputData is a struct for binding login and password.
type InputData struct {
	Login    string `json:"login" form:"login"`
	Password string `json:"password" form:"password"`
}

// SignUpUser handles the POST request to create a new user.
// @Summary SignUpUser
// @Description Create account
// @ID create-account
// @Tags auth
// @Accept json
// @Produce json
// @Param input body InputData true "info"
// @Success 201 {string} string "token"
// @Failure 400 {object} error
// @Router /signup [post]
func (h *Handler) SignUpUser(c echo.Context) error {
	var newUser model.User
	newUser.ID = uuid.New()
	requestData := &InputData{}
	err := c.Bind(requestData)
	if err != nil {
		log.Errorf("error: %v", err)
		return c.JSON(http.StatusBadRequest, "Handler-SignUpUser: Invalid request payload")
	}
	newUser.Login = requestData.Login
	newUser.Password = []byte(requestData.Password)
	err = h.validate.StructCtx(c.Request().Context(), newUser)
	if err != nil {
		log.Errorf("error: %v", err)
		return c.JSON(http.StatusBadRequest, "Not valid data")
	}
	accessToken, refreshToken, err := h.userService.SignUpUser(c.Request().Context(), &newUser)
	if err != nil {
		log.WithFields(log.Fields{
			"Login":         newUser.Login,
			"Password":      newUser.Password,
			"Access Toke":   accessToken,
			"Refresh Token": refreshToken,
		}).Errorf("failed to get data: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Handler-SignUpUser: error in method h.entityService.SignUpUser() :")
	}
	return c.JSON(http.StatusCreated, echo.Map{
		"Access Token : ":  accessToken,
		"Refresh Token : ": refreshToken,
	})
}

// GetByLogin checked password.
// @Summary GetByLogin
// @Description Log In User
// @ID login
// @Tags auth
// @Accept json
// @Produce json
// @Param input body InputData true "info"
// @Success 200 {string} string "token"
// @Failure 400 {string} error
// @Router /login [post]
func (h *Handler) GetByLogin(c echo.Context) error {
	var requestData InputData
	err := c.Bind(&requestData)
	if err != nil {
		log.Errorf("error: %v", err)
		return c.JSON(http.StatusBadRequest, "Handler-GetByLogin: Invalid request payload")
	}
	var user model.User
	user.Login = requestData.Login
	user.Password = []byte(requestData.Password)
	err = h.validate.VarCtx(c.Request().Context(), requestData.Login, "required")
	if err != nil {
		log.Errorf("error: %v", err)
		return c.JSON(http.StatusBadRequest, "Not valid data: login field is empty")
	}
	err = h.validate.VarCtx(c.Request().Context(), requestData.Password, "required")
	if err != nil {
		log.Errorf("error: %v", err)
		return c.JSON(http.StatusBadRequest, "Not valid data: password field is empty")
	}
	var verify bool
	accessToken, refreshToken, err := service.GenerateTokens(user.ID)
	if err != nil {
		log.Errorf("error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Handler-GetByLogin-GenerateTokens: challenge id must have uuid format")
	}
	verify, user.ID, err = h.userService.GetByLogin(c.Request().Context(), user.Login, []byte(refreshToken), user.Password)
	if err != nil {
		log.WithFields(log.Fields{
			"Login":         user.Login,
			"Password":      user.Password,
			"Access Toke":   accessToken,
			"Refresh Token": refreshToken,
		}).Errorf("failed to get data: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Handler-GetByLogin: error in method GetByLogin() :")
	}
	if !verify {
		return echo.ErrUnauthorized
	}
	return c.JSON(http.StatusOK, echo.Map{
		"Access Token : ": accessToken,
	})
}

// InputTokens is a struct for binding access and refresh tokens.
type InputTokens struct {
	AccessToken  string `json:"accessToken" form:"accessToken"`
	RefreshToken string `json:"refreshToken" form:"refreshToken"`
}

// RefreshToken refreshed tokens by tokens.
// @Summary RefreshToken
// @Description Refresh Token
// @ID refresh-token
// @Tags auth
// @Accept json
// @Produce json
// @Param input body InputTokens true "info"
// @Success 200 {string} string "tokens"
// @Failure 400 {string} error
// @Router /refresh [post]
func (h *Handler) RefreshToken(c echo.Context) error {
	var requestData InputTokens
	err := c.Bind(&requestData)
	if err != nil {
		log.Errorf("error: %v", err)
		return c.JSON(http.StatusBadRequest, "Handler-RefreshToken: Invalid request payload")
	}
	accessToken, refreshToken, err := h.userService.RefreshToken(c.Request().Context(), requestData.AccessToken, requestData.RefreshToken)
	if err != nil {
		log.WithFields(log.Fields{
			"Access Toke":   requestData.AccessToken,
			"Refresh Token": requestData.RefreshToken,
		}).Errorf("failed to get data: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Handler-RefreshToken: error in method h.userService.RefreshToken :")
	}
	return c.JSON(http.StatusOK, echo.Map{
		"Access Token : ":  accessToken,
		"Refresh Token : ": refreshToken,
	})
}

// UploadImage upload a picture to server.
// @Summary UploadImage
// @Description Upload Image
// @ID upload-image
// @Tags image
// @Accept json
// @Produce json
// @Param image formData file true "Image file"
// @Success 201 {object} string
// @Failure 400 {object} error
// @Router /upload [post]
func (h *Handler) UploadImage(c echo.Context) error {
	file, err := c.FormFile("image")
	if err != nil {
		log.Errorf("error: %v", err)
		return c.String(http.StatusBadRequest, "Failed to retrieve file")
	}
	src, err := file.Open()
	if err != nil {
		log.Errorf("error: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to open file")
	}
	defer func() {
		errClose := src.Close()
		if errClose != nil {
			log.Errorf("error: %v", errClose)
		}
	}()
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))
	//nolint:gosec // Disabled for this line as security is verified elsewhere.
	dst, err := os.Create(filepath.Join("images", "upload", filename))
	if err != nil {
		log.Errorf("error: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to create file")
	}
	defer func() {
		errClose := dst.Close()
		if errClose != nil {
			log.Errorf("error: %v", errClose)
		}
	}()
	_, err = io.Copy(dst, src)
	if err != nil {
		log.Errorf("error: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to copy file")
	}
	return c.JSON(http.StatusOK, echo.Map{
		"filename": filename,
	})
}

// DownloadImage show picture in request.
// @Summary DownloadImage
// @Description Download Image
// @ID download-image
// @Tags image
// @Accept json
// @Produce json
// @Param filename path string true "Image filename"
// @Success 200 {object} object
// @Failure 400 {object} error
// @Router /download/{filename} [get]
func (h *Handler) DownloadImage(c echo.Context) error {
	imgname := c.Param("filename")
	imgpath := filepath.Join("images", "upload", imgname)
	cleanPath := filepath.Clean(imgpath)
	file, err := os.Open(cleanPath)
	if err != nil {
		log.Errorf("error: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to open file")
	}
	defer func() {
		errClose := file.Close()
		if errClose != nil {
			log.Errorf("error: %v", errClose)
		}
	}()
	contentType := mime.TypeByExtension(filepath.Ext(imgname))
	c.Response().Header().Set("Content-Type", contentType)
	if _, err := io.Copy(c.Response(), file); err != nil {
		log.Errorf("error: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to copy file")
	}
	return nil
}
