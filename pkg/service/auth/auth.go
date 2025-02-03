package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// Секретный ключ (используй ENV-переменную в продакшене!)
var jwtSecret = []byte("supersecretkey")

// Создание JWT-токена
func GenerateJWT(login string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": login,
		"exp":   time.Now().Add(time.Hour * 24).Unix(), // Токен на 24 часа
	})

	return token.SignedString(jwtSecret)
}

// Проверка JWT-токена
func ParseJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
}
