package jwt

import (
	"context"
	"errors"
	"fmt"
	"time"

	redisutil "github.com/aiops/AiOpsHub/backend/pkg/redis"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func getJWTSecret() []byte {
	secret := viper.GetString("jwt.secret")
	if secret == "" {
		secret = "aiops-secret-key-change-in-production"
	}
	return []byte(secret)
}

func getTokenExpireDuration() time.Duration {
	expireStr := viper.GetString("jwt.token_expire")
	if expireStr == "" {
		expireStr = "30m"
	}

	duration, err := time.ParseDuration(expireStr)
	if err != nil {
		return 30 * time.Minute
	}
	return duration
}

func GenerateToken(ctx context.Context, userID, username, role string) (string, error) {
	expireDuration := getTokenExpireDuration()

	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "aiops",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(getJWTSecret())
	if err != nil {
		return "", err
	}

	if redisutil.Client != nil {
		info := &redisutil.TokenInfo{
			UserID:    userID,
			Username:  username,
			Role:      role,
			TokenType: "access",
			CreatedAt: time.Now(),
			Source:    "login",
		}
		err = redisutil.SetToken(ctx, tokenString, info, expireDuration)
		if err != nil {
			return "", fmt.Errorf("failed to store token in Redis: %w", err)
		}
	}

	return tokenString, nil
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return getJWTSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func ValidateToken(ctx context.Context, tokenString string) (string, string, string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", "", "", err
	}

	if redisutil.Client != nil {
		exists, err := redisutil.ExistsToken(ctx, tokenString)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to check token in Redis: %w", err)
		}

		if !exists {
			return "", "", "", errors.New("token not found in Redis or expired")
		}
	}

	return claims.UserID, claims.Username, claims.Role, nil
}

func RefreshToken(ctx context.Context, tokenString string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	if redisutil.Client != nil {
		err = redisutil.DeleteToken(ctx, tokenString)
		if err != nil {
			return "", fmt.Errorf("failed to delete old token: %w", err)
		}
	}

	return GenerateToken(ctx, claims.UserID, claims.Username, claims.Role)
}

func Logout(ctx context.Context, tokenString string) error {
	if redisutil.Client != nil {
		return redisutil.DeleteToken(ctx, tokenString)
	}
	return nil
}
