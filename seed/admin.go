// seed/admin.go
package seed

import (
	"blog-go/config"
	"blog-go/database"
	"blog-go/models"
	"blog-go/utils"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateAdmin() {
	cfg := config.Load()
	coll := database.DB.Collection("users")

	var existing models.User
	err := coll.FindOne(context.Background(), bson.M{"email": cfg.AdminEmail}).Decode(&existing)
	if err == nil {
		log.Println("Admin already exists")
		return
	}

	hashed, _ := utils.HashPassword(cfg.AdminPass)
	admin := models.User{
		Email:     cfg.AdminEmail,
		Username:  "admin",
		Password:  hashed,
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
	}

	coll.InsertOne(context.Background(), admin)
	log.Printf("Admin created: %s / %s", cfg.AdminEmail, cfg.AdminPass)
}
