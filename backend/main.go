package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"strconv"

	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"

	"google.golang.org/api/option"
)

//go:embed dist
var assets embed.FS

func main() {

	ctx := context.Background()

	sa := option.WithCredentialsFile("./svcKey.json")

	db, err := firebase.NewApp(ctx, nil, sa)

	if err != nil {

		log.Fatalln(err)

	}

	client, err := db.Firestore(ctx)

	if err != nil {

		log.Fatalln(err)

	}

	defer client.Close()

	// Now you can use the 'client' to interact with Firestore

	log.Println("Successfully connected to Firestore!")

	r := gin.Default()
	// Enable Gin logging
	gin.DefaultWriter = log.Writer()
	gin.DefaultErrorWriter = log.Writer()

	// Initialize the application
	app := NewApp(client, ctx)

	// API routes
	api := r.Group("/api")
	{
		api.POST("/newtranslation", func(c *gin.Context) {
			data, _ := c.GetRawData()
			text := fmt.Sprintf("%s", data)
			log.Println("Received text:", text)
			result, err := app.GetTranslationNew(text)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, result)
		})

		api.GET("/translations", func(c *gin.Context) {
			translations, err := app.GetAllTranslations()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, translations)
		})

		api.GET("/translations/:id", func(c *gin.Context) {
			id := c.Param("id")
			iid, err := strconv.Atoi(id)
			translation, err := app.GetTranslation(iid)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, translation)
		})

		api.GET("/definitions/:word", func(c *gin.Context) {
			word := c.Param("word")
			definition, err := app.GetWordDefinitions(word)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, definition)
		})

		api.DELETE("/translations/:id", func(c *gin.Context) {
			id := c.Param("id")
			iid, err := strconv.Atoi(id)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
				return
			}
			err = app.DeleteTranslation(iid)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Translation deleted"})
		})
	}

	// Serve frontend static files from the embedded filesystem
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Try to serve the file from the embedded filesystem
		if path == "/" {
			path = "/index.html"
		}

		// Remove leading slash and add frontend/dist prefix
		filePath := "dist" + path

		// Check if the file exists
		content, err := assets.ReadFile(filePath)
		if err != nil {
			// If not found, return the index.html for SPA routing
			content, err = assets.ReadFile("frontend/dist/index.html")
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
		}

		// Set content type based on file extension
		c.Data(http.StatusOK, getContentType(path), content)
	})

	// Run the server
	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getContentType returns the content type based on file extension
func getContentType(path string) string {
	switch {
	case path[len(path)-3:] == ".js":
		return "application/javascript"
	case path[len(path)-4:] == ".css":
		return "text/css"
	case path[len(path)-5:] == ".html":
		return "text/html"
	case path[len(path)-4:] == ".png":
		return "image/png"
	case path[len(path)-4:] == ".jpg", path[len(path)-5:] == ".jpeg":
		return "image/jpeg"
	case path[len(path)-4:] == ".svg":
		return "image/svg+xml"
	case path[len(path)-5:] == ".woff", path[len(path)-6:] == ".woff2":
		return "font/woff"
	default:
		return "text/plain"
	}
}
