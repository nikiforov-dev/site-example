package handler

import (
	"backend-example/pkg/service/auth"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware проверяет JWT-токен
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка Authorization
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			tokenString, _ = c.Cookie("jwtToken")
		}
		if tokenString == "" {
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		// Если токен передан в формате "Bearer <token>", извлекаем сам токен
		parts := strings.Split(tokenString, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString = parts[1]
		}

		// Проверяем токен
		token, err := auth.ParseJWT(tokenString)
		if err != nil || !token.Valid {
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		// Получаем login.html пользователя из claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		login, exists := claims["login"].(string)
		if !exists {
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		fmt.Println("----------- User is:", login)

		// Сохраняем login.html в контексте запроса для использования в обработчиках
		c.Set("login", login)

		// Передаём управление следующему обработчику
		c.Next()
	}
}
