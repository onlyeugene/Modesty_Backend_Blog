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

	// Parse the multipart form manually to handle files and fields
	if err := c.Request.ParseMultipartForm(512 << 20); err != nil { // 512 MB limit
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse multipart form: " + err.Error()})
		return
	}

	var req models.CreatePostRequest
	// Manually bind fields from the parsed form
	req.Title = c.Request.MultipartForm.Value["title"][0]
	req.Subheading = c.Request.MultipartForm.Value["subheading"][0]
	req.Content = c.Request.MultipartForm.Value["content"][0]

	// Handle image upload
	var imageURL string
	if file, handler, err := c.Request.FormFile("image"); err == nil {
		defer file.Close()
		cld, _ := cloudinary.NewFromURL(config.Load().CloudinaryURL)
		resp, _ := cld.Upload.Upload(context.Background(), file, uploader.UploadParams{Folder: "blog_posts", PublicID: handler.Filename})
		imageURL = resp.SecureURL
	} else if err != http.ErrMissingFile {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image upload failed"})
		return
	}

	// Handle video upload
	var videoURL []string
	if files, ok := c.Request.MultipartForm.File["video"]; ok {
		cld, _ := cloudinary.NewFromURL(config.Load().CloudinaryURL)
		for _, fileHeader := range files {
			file, _ := fileHeader.Open()
			defer file.Close()
			resp, _ := cld.Upload.Upload(context.Background(), file, uploader.UploadParams{Folder: "blog_videos", ResourceType: "video", PublicID: fileHeader.Filename})
			videoURL = append(videoURL, resp.SecureURL)
		}
	}

	p := models.Post{
		Title:      req.Title,
		Subheading: req.Subheading,
		Content:    req.Content,
		ImageURL:   imageURL,
		VideoURL:   videoURL,
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

	// Parse the multipart form manually to handle files and fields
	if err := c.Request.ParseMultipartForm(512 << 20); err != nil { // 512 MB limit
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse multipart form: " + err.Error()})
		return
	}

	var req models.CreatePostRequest
	// Manually bind fields from the parsed form
	req.Title = c.Request.MultipartForm.Value["title"][0]
	req.Subheading = c.Request.MultipartForm.Value["subheading"][0]
	req.Content = c.Request.MultipartForm.Value["content"][0]

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
	// No need for req.VideoURL check here, it's handled by file upload logic

	// Handle image upload
	if file, handler, err := c.Request.FormFile("image"); err == nil {
		defer file.Close()
		cld, _ := cloudinary.NewFromURL(config.Load().CloudinaryURL)
		resp, _ := cld.Upload.Upload(context.Background(), file, uploader.UploadParams{Folder: "blog_posts", PublicID: handler.Filename})
		update["image_url"] = resp.SecureURL
	} else if err != http.ErrMissingFile {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image upload failed"})
		return
	}

	// Handle video upload
	if files, ok := c.Request.MultipartForm.File["video"]; ok {
		cld, _ := cloudinary.NewFromURL(config.Load().CloudinaryURL)
		var videoURLs []string
		for _, fileHeader := range files {
			file, _ := fileHeader.Open()
			defer file.Close()
			resp, _ := cld.Upload.Upload(context.Background(), file, uploader.UploadParams{Folder: "blog_videos", ResourceType: "video", PublicID: fileHeader.Filename})
			videoURLs = append(videoURLs, resp.SecureURL)
		}
		update["video_url"] = videoURLs
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
