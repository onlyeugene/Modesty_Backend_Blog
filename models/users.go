// models/user.go
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email     string             `bson:"email" json:"email"`
	Username  string             `bson:"username" json:"username"`
	Password  string             `bson:"hashed_password" json:"-"`
	CreatedAt primitive.DateTime `bson:"created_at" json:"created_at"`
}
