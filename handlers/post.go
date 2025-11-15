package handlers

import (
	"context"
	"net/http"
	"time"

	"blog-go/config"
	"blog-go/database"
	"blog-go/models"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreatePost(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := primitive.ObjectIDFromHex(userID.(string))

	var req models.CreatePostRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var imageURL string
	if file, err := c.FormFile("image"); err == nil {
		f, _ := file.Open()
		defer f.Close()
		cld, _ := cloudinary.NewFromURL(config.Load().CloudinaryURL)
		resp, _ := cld.Upload.Upload(context.Background(), f, uploader.UploadParams{Folder: "blog_posts"})
		imageURL = resp.SecureURL
	}

	p := models.Post{
		Title:     req.Title,
		Content:   req.Content,
		ImageURL:  imageURL,
		AuthorID:  uid,
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt: primitive.NewDateTimeFromTime(time.Now()),
	}
	coll := database.DB.Collection("posts")
	res, _ := coll.InsertOne(c.Request.Context(), p)
	p.ID = res.InsertedID.(primitive.ObjectID)

	// author name
	var author models.User
	database.DB.Collection("users").FindOne(c.Request.Context(), bson.M{"_id": uid}).Decode(&author)
	p.Author = author.Username

	c.JSON(http.StatusOK, p)
}

func GetPosts(c *gin.Context) {
	coll := database.DB.Collection("posts")
	opts := options.Find().SetSort(bson.D{{"created_at", -1}})
	cur, _ := coll.Find(c.Request.Context(), bson.M{}, opts)

	var posts []models.Post
	cur.All(c.Request.Context(), &posts)

	for i := range posts {
		var u models.User
		database.DB.Collection("users").FindOne(c.Request.Context(), bson.M{"_id": posts[i].AuthorID}).Decode(&u)
		posts[i].Author = u.Username
	}
	c.JSON(http.StatusOK, posts)
}

func GetPostByID(c *gin.Context) {
	postID := c.Param("id")
	pid, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	coll := database.DB.Collection("posts")
	var post models.Post
	if err := coll.FindOne(c.Request.Context(), bson.M{"_id": pid}).Decode(&post); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	var u models.User
	database.DB.Collection("users").FindOne(c.Request.Context(), bson.M{"_id": post.AuthorID}).Decode(&u)
	post.Author = u.Username

	c.JSON(http.StatusOK, post)
}

func UpdatePost(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := primitive.ObjectIDFromHex(userID.(string))
	postID := c.Param("id")
	pid, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	coll := database.DB.Collection("posts")
	var post models.Post
	if err := coll.FindOne(c.Request.Context(), bson.M{"_id": pid}).Decode(&post); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}
	if post.AuthorID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return
	}

	var req models.CreatePostRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := bson.M{
		"title":      req.Title,
		"content":    req.Content,
		"updated_at": primitive.NewDateTimeFromTime(time.Now()),
	}

	if file, err := c.FormFile("image"); err == nil {
		f, _ := file.Open()
		defer f.Close()
		cld, _ := cloudinary.NewFromURL(config.Load().CloudinaryURL)
		resp, _ := cld.Upload.Upload(context.Background(), f, uploader.UploadParams{Folder: "blog_posts"})
		update["image_url"] = resp.SecureURL
	}

	coll.UpdateOne(c.Request.Context(), bson.M{"_id": pid}, bson.M{"$set": update})
	coll.FindOne(c.Request.Context(), bson.M{"_id": pid}).Decode(&post)
	post.Author = "You" // or fetch username
	c.JSON(http.StatusOK, post)
}

func DeletePost(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := primitive.ObjectIDFromHex(userID.(string))
	postID := c.Param("id")
	pid, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	coll := database.DB.Collection("posts")
	var post models.Post
	if err := coll.FindOne(c.Request.Context(), bson.M{"_id": pid}).Decode(&post); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if post.AuthorID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return
	}

	coll.DeleteOne(c.Request.Context(), bson.M{"_id": pid})
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
