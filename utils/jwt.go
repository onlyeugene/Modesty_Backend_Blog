// utils/jwt.go
package utils

import (
	"blog-go/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID string, userRole string) (string, error) {
	cfg := config.Load()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  userID,
		"role": userRole,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString([]byte(cfg.JWTSecret))
}
