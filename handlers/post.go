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
	if err := c.ShouldBind(&req); err != nil { // Expecting multipart/form-data for text fields and image
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
		Title:      req.Title,
		Subheading: req.Subheading,
		Content:    req.Content,
		ImageURL:   imageURL,
		VideoURL:   []string{}, // Handled by separate endpoint
		AuthorID:   uid,
		Time:       primitive.NewDateTimeFromTime(time.Now()),
		Date:       primitive.NewDateTimeFromTime(time.Now()),
		CreatedAt:  primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:  primitive.NewDateTimeFromTime(time.Now()),
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
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cur, _ := coll.Find(c.Request.Context(), bson.M{}, opts)

	var posts []models.Post
	cur.All(c.Request.Context(), &posts)

	if len(posts) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No posts found"})
		return
	}

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
	// If not the author, check if user is an admin
	userRole, _ := c.Get("userRole")
	if post.AuthorID != uid && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return
	}

	var req models.CreatePostRequest           // Reuse CreatePostRequest for form fields
	if err := c.ShouldBind(&req); err != nil { // Expecting multipart/form-data for text fields and image
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := bson.M{
		"updated_at": primitive.NewDateTimeFromTime(time.Now()),
		"time":       primitive.NewDateTimeFromTime(time.Now()),
		"date":       primitive.NewDateTimeFromTime(time.Now()),
	}

	if req.Title != "" {
		update["title"] = req.Title
	}
	if req.Subheading != "" {
		update["subheading"] = req.Subheading
	}
	if req.Content != "" {
		update["content"] = req.Content
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

	// If not the author, check if user is an admin
	userRole, _ := c.Get("userRole")
	if post.AuthorID != uid && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return
	}

	coll.DeleteOne(c.Request.Context(), bson.M{"_id": pid})
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// UploadPostVideo handles video uploads for a specific post
func UploadPostVideo(c *gin.Context) {
	postID := c.Param("id")
	pid, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	// Verify user authorization (author or admin)
	userID, _ := c.Get("userID")
	uid, _ := primitive.ObjectIDFromHex(userID.(string))
	userRole, _ := c.Get("userRole")

	coll := database.DB.Collection("posts")
	var post models.Post
	if err := coll.FindOne(c.Request.Context(), bson.M{"_id": pid}).Decode(&post); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}
	if post.AuthorID != uid && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to upload video for this post"})
		return
	}

	// Parse the multipart form to get all video files
	if err := c.Request.ParseMultipartForm(512 << 20); err != nil { // 512 MB limit for total video uploads
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse multipart form: " + err.Error()})
		return
	}

	files := c.Request.MultipartForm.File["video"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no video files provided"})
		return
	}

	cld, _ := cloudinary.NewFromURL(config.Load().CloudinaryURL)
	var uploadedVideoURLs []string
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open video file: " + err.Error()})
			return
		}
		defer file.Close()

		resp, err := cld.Upload.Upload(context.Background(), file, uploader.UploadParams{Folder: "blog_videos", ResourceType: "video", PublicID: fileHeader.Filename})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload video to cloudinary: " + err.Error()})
			return
		}
		uploadedVideoURLs = append(uploadedVideoURLs, resp.SecureURL)
	}

	// Append new video URLs to existing ones
	var updatedVideoURLs []string
	if post.VideoURL != nil {
		updatedVideoURLs = append(updatedVideoURLs, post.VideoURL...)
	}
	updatedVideoURLs = append(updatedVideoURLs, uploadedVideoURLs...)

	update := bson.M{"video_url": updatedVideoURLs, "updated_at": primitive.NewDateTimeFromTime(time.Now())}
	coll.UpdateOne(c.Request.Context(), bson.M{"_id": pid}, bson.M{"$set": update})

	c.JSON(http.StatusOK, gin.H{"message": "videos uploaded successfully", "video_urls": uploadedVideoURLs})
}
