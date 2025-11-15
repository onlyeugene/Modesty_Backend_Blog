// handlers/auth.go
package handlers

import (
	"blog-go/database"
	"blog-go/models"
	"blog-go/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	coll := database.DB.Collection("users")

	var existing models.User
	err := coll.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existing)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
		return
	}

	hashed, _ := utils.HashPassword(req.Password)
	user := models.User{
		Email:     req.Email,
		Username:  req.Username,
		Password:  hashed,
		Role:      "admin", // Default role for new users
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
	}

	res, _ := coll.InsertOne(ctx, user)
	user.ID = res.InsertedID.(primitive.ObjectID)

	token, _ := utils.GenerateToken(user.ID.Hex(), user.Role)
	c.JSON(http.StatusOK, gin.H{"token": token, "user": gin.H{
		"id": user.ID.Hex(), "email": user.Email, "username": user.Username, "role": user.Role,
	}})
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	coll := database.DB.Collection("users")

	var user models.User
	err := coll.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil || !utils.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, _ := utils.GenerateToken(user.ID.Hex(), user.Role)
	c.JSON(http.StatusOK, gin.H{"token": token})
}
