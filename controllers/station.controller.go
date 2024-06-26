package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mliem2k/ottb-go/models"
	"gorm.io/gorm"
)

type StationController struct {
	DB *gorm.DB
}

func NewStationController(DB *gorm.DB) StationController {
	return StationController{DB}
}

func (pc *StationController) CreateStation(ctx *gin.Context) {
	// currentUser := ctx.MustGet("currentUser").(models.User)

	// Parse form data
	err := ctx.Request.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Failed to parse form data"})
		return
	}

	// Extract other fields from the form
	name := ctx.Request.FormValue("name")
	latlong := ctx.Request.FormValue("latlong")
	user := ctx.Request.FormValue("user")
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

	// Create new station object
	now := time.Now()
	newStation := models.Station{
		Name:      name,
		UserId:    parsedUser,
		LatLong:   latlong,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save station to database
	result := pc.DB.Create(&newStation)
	if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": newStation})
}

func (pc *StationController) UpdateStation(ctx *gin.Context) {
	stationId := ctx.Param("stationId")
	currentUser := ctx.MustGet("currentUser").(models.User)

	var payload *models.UpdateStation
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "fail", "message": err.Error()})
		return
	}
	var updatedStation models.Station
	result := pc.DB.First(&updatedStation, "id = ?", stationId)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "No station with that title exists"})
		return
	}
	now := time.Now()
	stationToUpdate := models.Station{
		Name:      payload.Name,
		UserId:    currentUser.ID,
		CreatedAt: updatedStation.CreatedAt,
		UpdatedAt: now,
	}

	pc.DB.Model(&updatedStation).Updates(stationToUpdate)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": updatedStation})
}

func (pc *StationController) FindStationById(ctx *gin.Context) {
	stationId := ctx.Param("stationId")

	var station models.Station
	result := pc.DB.First(&station, "id = ?", stationId)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "No station with that title exists"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": station})
}

func (pc *StationController) FindStations(ctx *gin.Context) {
	var page = ctx.DefaultQuery("page", "1")
	var limit = ctx.DefaultQuery("limit", "10")

	intPage, _ := strconv.Atoi(page)
	intLimit, _ := strconv.Atoi(limit)
	offset := (intPage - 1) * intLimit

	var stations []models.Station
	results := pc.DB.Limit(intLimit).Offset(offset).Find(&stations)
	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(stations), "data": stations})
}

func (pc *StationController) DeleteStation(ctx *gin.Context) {
	stationId := ctx.Param("stationId")

	result := pc.DB.Delete(&models.Station{}, "id = ?", stationId)

	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "No station with that title exists"})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}
