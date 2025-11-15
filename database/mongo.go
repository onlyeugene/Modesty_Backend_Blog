// database/mongo.go
package database

import (
	"context"
	"log"
	"time"

	"blog-go/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var DB *mongo.Database

func Connect() {
	cfg := config.Load()
	clientOptions := options.Client().ApplyURI(cfg.MONGODB_URL)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	Client = client
	DB = client.Database("blogdb")
	log.Println("Connected to MongoDB")
}
