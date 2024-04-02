package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/wpcodevo/golang-gorm-postgres/controllers"
	"github.com/wpcodevo/golang-gorm-postgres/middleware"
)

type StationRouteController struct {
	stationController controllers.StationController
}

func NewRouteStationController(stationController controllers.StationController) StationRouteController {
	return StationRouteController{stationController}
}

func (pc *StationRouteController) StationRoute(rg *gin.RouterGroup) {

	router := rg.Group("stations")
	router.Use(middleware.DeserializeUser())
	router.POST("/", pc.stationController.CreateStation)
	router.GET("/", pc.stationController.FindStations)
	router.PUT("/:stationId", pc.stationController.UpdateStation)
	router.GET("/:stationId", pc.stationController.FindStationById)
	router.DELETE("/:stationId", pc.stationController.DeleteStation)
	router.GET("/users/:userId/stations", pc.stationController.FindStationsByUserId)
}
