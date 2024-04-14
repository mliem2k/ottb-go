package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/mliem2k/ottb-go/controllers"
)

type StationRouteController struct {
	stationController controllers.StationController
}

func NewRouteStationController(stationController controllers.StationController) StationRouteController {
	return StationRouteController{stationController}
}

func (pc *StationRouteController) StationRoute(rg *gin.RouterGroup) {

	router := rg.Group("stations")
	// router.Use(middleware.DeserializeUser())
	router.POST("", pc.stationController.CreateStation)
	router.GET("", pc.stationController.FindStations)
	router.PUT("/:stationId", pc.stationController.UpdateStation)
	router.GET("/:stationId", pc.stationController.FindStationById)
	router.DELETE("/:stationId", pc.stationController.DeleteStation)
}
