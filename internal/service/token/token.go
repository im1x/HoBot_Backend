package token

import (
	"HoBot_Backend/internal/model"
	"os"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenExpireHour  = 12
	RefreshTokenExpireHour = 1440
)

func generateToken(userId string, secret string, expHour time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"id":  userId,
		"exp": time.Now().Add(time.Hour * expHour).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(secret))

	return t, err
}
func GenerateTokens(userId string) (string, string) {
	accessToken, err := generateToken(userId, os.Getenv("JWT_ACCESS_SECRET"), AccessTokenExpireHour)
	if err != nil {
		log.Error("Generate access token error")
	}

	refreshToken, err := generateToken(userId, os.Getenv("JWT_REFRESH_SECRET"), RefreshTokenExpireHour)
	if err != nil {
		log.Error("Generate refresh token error")
	}

	return accessToken, refreshToken
}

func GenerateRefreshToken(userId string) string {
	token, err := generateToken(userId, os.Getenv("JWT_REFRESH_SECRET"), RefreshTokenExpireHour)
	if err != nil {
		log.Error("Generate refresh token error")
		return ""
	}
	return token
}

func isTokenValid(tokenString, secret string) (*model.UserDto, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.UserDto{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}, jwt.WithLeeway(5*time.Second))

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*model.UserDto); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

func ValidateAccessToken(token string) (*model.UserDto, error) {
	return isTokenValid(token, os.Getenv("JWT_ACCESS_SECRET"))
}

func ValidateRefreshToken(token string) (*model.UserDto, error) {
	return isTokenValid(token, os.Getenv("JWT_REFRESH_SECRET"))
}
