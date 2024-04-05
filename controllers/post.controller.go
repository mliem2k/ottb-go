package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wpcodevo/golang-gorm-postgres/initializers"
	"github.com/wpcodevo/golang-gorm-postgres/models"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostController struct {
	DB *gorm.DB
}

func NewPostController(DB *gorm.DB) PostController {
	return PostController{DB}
}

func (pc *PostController) CreatePost(ctx *gin.Context) {
	// currentUser := ctx.MustGet("currentUser").(models.User)

	// Parse form data
	err := ctx.Request.ParseMultipartForm(10 << 20) // 10 MB max

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Failed to parse form data"})
		return
	}

	// Extract other fields from the form
	title := ctx.Request.FormValue("title")
	content := ctx.Request.FormValue("content")
	user := ctx.Request.FormValue("user")
	parsedUUID, err := uuid.Parse(user)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid user ID"})
		return
	}

	// Handle file uploads
	files := ctx.Request.MultipartForm.File["image"]
	var imageNames []string
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Failed to get image file"})
			return
		}
		defer file.Close()

		// Save uploaded file to server with timestamp
		currentTime := time.Now().UnixNano()
		imageExtension := filepath.Ext(fileHeader.Filename)
		imageName := fmt.Sprintf("%d%s", currentTime, imageExtension)
		filePath := "uploads/" + imageName
		outFile, err := os.Create(filePath)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to save image file"})
			return
		}
		defer outFile.Close()
		_, err = io.Copy(outFile, file)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to save image file"})
			return
		}
		imageNames = append(imageNames, imageName)
	}

	// Create new post object
	now := time.Now()
	newPost := models.Post{
		Title:     title,
		Content:   content,
		Image:     strings.Join(imageNames, ","),
		UserId:    parsedUUID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save post to database
	result := pc.DB.Create(&newPost)
	if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": newPost})
}

func (pc *PostController) UpdatePost(ctx *gin.Context) {
	postId := ctx.Param("postId")
	currentUser := ctx.MustGet("currentUser").(models.User)

	var payload *models.UpdatePost
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "fail", "message": err.Error()})
		return
	}
	var updatedPost models.Post
	result := pc.DB.First(&updatedPost, "id = ?", postId)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "No post with that title exists"})
		return
	}
	now := time.Now()
	postToUpdate := models.Post{
		Title:     payload.Title,
		Content:   payload.Content,
		Image:     payload.Image,
		UserId:    currentUser.ID,
		CreatedAt: updatedPost.CreatedAt,
		UpdatedAt: now,
	}

	pc.DB.Model(&updatedPost).Updates(postToUpdate)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": updatedPost})
}

func (pc *PostController) FindPostById(ctx *gin.Context) {
	postId := ctx.Param("postId")

	var post models.Post
	result := pc.DB.First(&post, "id = ?", postId)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "No post with that title exists"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": post})
}

func (pc *PostController) FindPostsByUserId(ctx *gin.Context) {
	config, err := initializers.LoadConfig(".")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to load config"})
		return
	}

	userId := ctx.Param("userId")

	// Set logger to write logs to os.Stdout
	pc.DB.Logger.LogMode(logger.Info)

	var posts []models.Post
	result := pc.DB.Debug().Where("user_id = ?", userId).Find(&posts)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to find posts"})
		return
	}

	// Create a new JSON structure without using struct images
	var responseData []map[string]interface{}
	for _, post := range posts {
		images := strings.Split(post.Image, ",")
		var formattedImages []string
		for _, image := range images {
			formattedImages = append(formattedImages, config.ClientOrigin+"/uploads/"+image)
		}

		responseData = append(responseData, map[string]interface{}{
			"id":         post.ID,
			"title":      post.Title,
			"content":    post.Content,
			"images":     formattedImages,
			"user_id":    post.UserId,
			"created_at": post.CreatedAt.Format(time.RFC3339),
			"updated_at": post.UpdatedAt.Format(time.RFC3339),
		})
	}
	var updatedResponseData []map[string]interface{}
	for _, data := range responseData {
		// Extract the images slice
		images, ok := data["images"].([]string)
		if !ok {
			// If the `images` key is not present or not a []string, skip this item
			continue
		}
		// Iterate over each image URL
		for _, img := range images {
			// Create a new map for each image URL
			newData := map[string]interface{}{
				"id":         data["id"],
				"title":      data["title"],
				"content":    data["content"],
				"image":      img, // Set the image URL
				"user_id":    data["user_id"],
				"created_at": data["created_at"],
				"updated_at": data["updated_at"],
			}

			// Append the new map to the updated response data slice
			updatedResponseData = append(updatedResponseData, newData)
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": updatedResponseData})
}

func (pc *PostController) FindPosts(ctx *gin.Context) {
	var page = ctx.DefaultQuery("page", "1")
	var limit = ctx.DefaultQuery("limit", "10")

	intPage, _ := strconv.Atoi(page)
	intLimit, _ := strconv.Atoi(limit)
	offset := (intPage - 1) * intLimit

	var posts []models.Post
	results := pc.DB.Limit(intLimit).Offset(offset).Find(&posts)
	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(posts), "data": posts})
}

func (pc *PostController) DeletePost(ctx *gin.Context) {
	postId := ctx.Param("postId")

	result := pc.DB.Delete(&models.Post{}, "id = ?", postId)

	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "No post with that title exists"})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}
