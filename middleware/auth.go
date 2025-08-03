package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error_messages": "missing required token",
			})
			c.Abort()

			return
		}

		// format token
		// Authorization: Bearer xxx
		// Split = [Bearer] [xxx]
		tokenString := strings.Split(authHeader, " ")

		if len(tokenString) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error_messages": "Invalid token.",
			})
			c.Abort()

			return
		}

		token, err := jwt.Parse(tokenString[1], func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error_messages": "Invalid token.",
			})
			c.Abort()

			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)

		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error_messages": "Invalid token.",
			})
			c.Abort()

			return
		}

		c.Set("user_id", claims["user_id"].(float64))
		c.Next()
	}
}
