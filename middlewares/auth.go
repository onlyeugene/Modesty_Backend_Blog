// middleware/auth.go
package middleware

import (
	"net/http"
	"strings"

	"blog-go/config"
	"blog-go/database"
	"blog-go/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			c.Abort()
			return
		}

		tokenStr := strings.Replace(authHeader, "Bearer ", "", 1)
		cfg := config.Load()

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, _ := token.Claims.(jwt.MapClaims)
		userID := claims["sub"].(string)

		// Fetch user from DB to get role
		var user models.User
		objID, _ := primitive.ObjectIDFromHex(userID)
		err = database.DB.Collection("users").FindOne(c.Request.Context(), bson.M{"_id": objID}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Set("userRole", user.Role)
		c.Next()
	}
}
