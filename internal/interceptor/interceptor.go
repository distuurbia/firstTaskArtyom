// Package interceptor contains all interceptors for grpc
package interceptor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/distuurbia/firstTaskArtyom/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// CustomInterceptor is an interceptor for grpc
type CustomInterceptor struct {
	cfg *config.Config
}

func NewCustomInterceptor (cfg *config.Config) *CustomInterceptor{
	return &CustomInterceptor{cfg: cfg}
}

// UnaryInterceptor globally need to check auth header for excisting or not jwt token and with this info it gives or not access to requested info
//
//nolint:gocyclo //Because there is too many branching when admin exists at logic and we should make a check for admin rights
func (ci *CustomInterceptor) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handl grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	authorization := md.Get("Authorization")
	if (strings.Contains(info.FullMethod, "/SignUpAdmin") || strings.Contains(info.FullMethod, "/Delete")) && ok && (len(authorization) > 0) {
		token, err := tokenParse(authorization[0], ci.cfg)
		if err != nil {
			logrus.Errorf("failed to parse token: %v", err)
			return nil, status.Errorf(codes.Internal, "failed to parse token error: ")
		}
		expired := tokenExpCheck(token)
		if !expired {
			logrus.Error("Token is expired")
			return nil, status.Errorf(codes.Unauthenticated, "Token is expired: ")
		}
		admin := tokenAdminCheck(token)
		if !admin {
			logrus.Error("admin field is false")
			return nil, status.Errorf(codes.PermissionDenied, "You don't have enough rights: ")
		}
		resp, err := handl(ctx, req)
		if err != nil {
			logrus.Fatalf("Failed to run handler method: %v", err)
		}
		return resp, err
	}
	if strings.Contains(info.FullMethod, "/UserService") || strings.Contains(info.FullMethod, "/ImageService") {
		resp, err := handl(ctx, req)
		if err != nil {
			logrus.Fatalf("Failed to run handler method: %v", err)
		}
		return resp, err
	}

	if ok && (len(authorization) > 0) {
		token, err := tokenParse(authorization[0], ci.cfg)
		if err != nil {
			logrus.Errorf("failed to parse token: %v", err)
			return nil, err
		}
		expired := tokenExpCheck(token)
		if !expired {
			logrus.Error("Token is expired")
			return "Token is expired", err
		}
		resp, err := handl(ctx, req)
		if err != nil {
			logrus.Fatalf("Failed to run handler method: %v", err)
		}
		return resp, err
	}
	logrus.Error("not found auth token")
	return nil, status.Errorf(codes.PermissionDenied, "not found auth token")
}

// tokenParse parses token and checks if it valid
func tokenParse(authorization string, cfg *config.Config) (*jwt.Token, error) {
	tokenString := strings.TrimPrefix(authorization, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.AccessTokenSignature), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token %w", err)
	}
	return token, nil
}

// tokenExpCheck checks is token expired or not
func tokenExpCheck(token *jwt.Token) bool {
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		exp := claims["exp"].(float64)
		if exp < float64(time.Now().Unix()) {
			return false
		}
	}
	return true
}

// tokenAdminCheck checks status of admin field in tokens payload
func tokenAdminCheck(token *jwt.Token) bool {
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		admin := claims["admin"].(bool)
		if !admin {
			return false
		}
	}
	return true
}
