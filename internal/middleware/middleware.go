// Package middleware provides common middleware functions for request handling.
package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/distuurbia/firstTaskArtyom/internal/config"
	"github.com/caarlos0/env"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// JWTMiddleware is a middleware function that performs JWT.
func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Missing authorization header")
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authorization header")
		}
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.AccessTokenSignature), nil
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
		}
		if !token.Valid {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			exp := claims["exp"].(float64)
			if exp < float64(time.Now().Unix()) {
				return echo.NewHTTPError(http.StatusUnauthorized, "Token expired")
			}
		}
		c.Set("users", token)
		return next(c)
	}
}

// JWTMiddleware is a middleware function that performs JWT.
func JWTMiddlewareAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Missing authorization header")
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authorization header")
		}
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.AccessTokenSignature), nil
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
		}
		if !token.Valid {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			exp := claims["exp"].(float64)
			if exp < float64(time.Now().Unix()) {
				return echo.NewHTTPError(http.StatusUnauthorized, "Token expired")
			}
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			admin := claims["admin"].(bool)
			if !admin {
				return echo.NewHTTPError(http.StatusUnauthorized, "You don't have enough rights to use this method")
			}
		}
		c.Set("users", token)
		return next(c)
	}
}
