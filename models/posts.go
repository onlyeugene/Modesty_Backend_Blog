package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Post struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title      string             `bson:"title" json:"title"`
	Subheading string             `bson:"subheading,omitempty" json:"subheading,omitempty"`
	Content    string             `bson:"content" json:"content"`
	ImageURL   string             `bson:"image_url,omitempty" json:"image_url,omitempty"`
	VideoURL   []string           `bson:"video_url,omitempty" json:"video_url,omitempty"`
	AuthorID   primitive.ObjectID `bson:"author_id" json:"author_id"`
	Author     string             `bson:"-" json:"author"`
	Time       primitive.DateTime `bson:"time" json:"time"`
	Date       primitive.DateTime `bson:"date" json:"date"`
	CreatedAt  primitive.DateTime `bson:"created_at" json:"created_at"`
	UpdatedAt  primitive.DateTime `bson:"updated_at" json:"updated_at"`
}

// Used for multipart binding
type CreatePostRequest struct {
	Title      string   `form:"title" binding:"required"`
	Subheading string   `form:"subheading,omitempty"`
	Content    string   `form:"content" binding:"required"`
	VideoURL   []string `form:"video_url,omitempty"`
}
