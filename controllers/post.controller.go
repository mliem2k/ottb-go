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
	"github.com/mliem2k/ottb-go/initializers"
	"github.com/mliem2k/ottb-go/models"
	"gorm.io/gorm"
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
	user := ctx.Request.FormValue("user")

	parsedUser, err := uuid.Parse(user)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid user ID"})
		return
	}
	stationId := ctx.Request.FormValue("stationId")
	parsedStationId, err := uuid.Parse(stationId)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid station ID"})
		return
	}

	// Create new post object
	now := time.Now()
	newPost := models.Post{
		Title: title,
		// Content:   content,
		// Image:     strings.Join(imageNames, ","),
		Developed: false,
		UserId:    parsedUser,
		StationId: parsedStationId,
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
	// currentUser := ctx.MustGet("currentUser").(models.User)
	err := ctx.Request.ParseMultipartForm(10 << 20) // 10 MB max

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Failed to parse form data"})
		return
	}
	user := ctx.Request.FormValue("user")
	title := ctx.Request.FormValue("title")
	// image := ctx.Request.FormValue("image")
	developedStr := ctx.Request.FormValue("developed")
	developed, _ := strconv.ParseBool(developedStr)

	parsedUser, err := uuid.Parse(user)

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
	// var payload *models.UpdatePost
	var postModel models.Post
	now := time.Now()

	result := pc.DB.First(&postModel, "id = ?", postId)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "No post with that title exists"})
		return
	}
	// fmt.Println(result.)
	postToUpdate := map[string]interface{}{
		"title":      title,
		"image":      strings.Join(imageNames, ","),
		"developed":  developed,
		"user_id":    parsedUser,
		"created_at": postModel.CreatedAt,
		"updated_at": now,
	}
	for key, value := range postToUpdate {
		if strValue, ok := value.(string); ok && strValue == "" {
			// If value is an empty string, remove it from the map
			delete(postToUpdate, key)
		}
	}
	pc.DB.Debug().Model(&postModel).Updates(postToUpdate)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": postToUpdate})
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

	var posts []models.Post
	result := pc.DB.Where("user_id = ? AND image <> '' AND developed = ?", userId, true).Find(&posts)

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
			formattedImages = append(formattedImages, config.ServerOrigin+"/uploads/"+image)
		}

		responseData = append(responseData, map[string]interface{}{
			"id":         post.ID,
			"title":      post.Title,
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
	config, err := initializers.LoadConfig(".")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to load config"})
		return
	}

	// Get query parameters
	page := ctx.DefaultQuery("page", "1")
	limit := ctx.DefaultQuery("limit", "10")
	developed := ctx.Query("developed")

	// Convert page and limit to integers
	intPage, _ := strconv.Atoi(page)
	intLimit, _ := strconv.Atoi(limit)
	offset := (intPage - 1) * intLimit

	var posts []map[string]interface{}       // Using a map to store posts and station_name
	var posts_total []map[string]interface{} // Using a map to store posts and station_name

	baseQuery := pc.DB.Debug().
		Table("posts").
		Select("posts.*, stations.name as station_name").
		Joins("JOIN stations ON posts.station_id::uuid = stations.id")

	// Apply pagination
	baseQueryTotal := baseQuery.Find(&posts_total)
	baseQueryLO := baseQuery.Limit(intLimit).Offset(offset)

	if baseQueryTotal.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": baseQueryTotal.Error})
		return
	}

	// Apply additional filter if "developed" parameter is present
	if developed != "" {
		baseQueryLO = baseQueryLO.Where("posts.developed = ?", developed)
	}

	// Execute the query
	results := baseQueryLO.Find(&posts)

	// Check for errors in query execution
	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	}
	for _, img := range posts {
		images := strings.Split(img["image"].(string), ",")
		var formattedImages []string
		for _, image := range images {
			formattedImages = append(formattedImages, config.ServerOrigin+"/uploads/"+image)
		}
		img["image"] = formattedImages
	}
	// Return JSON response with results
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "page": page, "results": len(posts), "total_results": len(posts_total), "data": posts})
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
