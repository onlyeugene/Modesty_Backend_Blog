// main.go
package main

import (
	"embed"
	"net/http"

	"blog-go/database"
	"blog-go/handlers"
	middleware "blog-go/middlewares"
	"blog-go/seed"

	"github.com/gin-gonic/gin"
)

//go:embed docs/*
var docsFS embed.FS // ← MUST be at package level

// main.go (fixed)
func main() {
	database.Connect()
	seed.CreateAdmin()

	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Public routes
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)
	// r.StaticFS("/docs", http.FS(docsFS))

	// PUBLIC: Anyone can read posts
	r.GET("/posts", handlers.GetPosts)

	// PROTECTED: Only logged-in users can create
	auth := r.Group("/")
	auth.Use(middleware.AuthRequired())
	{
		{
			auth.POST("/posts", handlers.CreatePost)
			auth.PUT("/posts/:id", handlers.UpdatePost)
			auth.DELETE("/posts/:id", handlers.DeletePost)
		}
	}
	// Add this BEFORE r.Run()
	r.GET("/docs/openapi.yaml", func(c *gin.Context) {
		data, _ := docsFS.ReadFile("docs/openapi.yaml")
		c.Data(200, "application/yaml", data)
	})
	r.GET("/docs", func(c *gin.Context) {
		data, _ := docsFS.ReadFile("docs/index.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})
	r.GET("/docs/", func(c *gin.Context) {
		data, _ := docsFS.ReadFile("docs/index.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	r.Run(":8000")
}
