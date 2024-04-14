package main

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/mliem2k/ottb-go/controllers"
	"github.com/mliem2k/ottb-go/initializers"
	"github.com/mliem2k/ottb-go/routes"
)

var (
	server              *gin.Engine
	AuthController      controllers.AuthController
	AuthRouteController routes.AuthRouteController

	UserController      controllers.UserController
	UserRouteController routes.UserRouteController

	PostController      controllers.PostController
	PostRouteController routes.PostRouteController

	StationController      controllers.StationController
	StationRouteController routes.StationRouteController
)

func init() {
	config, err := initializers.LoadConfig(".")
	if err != nil {
		log.Fatal("🚀 Could not load environment variables", err)
	}

	initializers.ConnectDB(&config)

	AuthController = controllers.NewAuthController(initializers.DB)
	AuthRouteController = routes.NewAuthRouteController(AuthController)

	UserController = controllers.NewUserController(initializers.DB)
	UserRouteController = routes.NewRouteUserController(UserController)

	PostController = controllers.NewPostController(initializers.DB)
	PostRouteController = routes.NewRoutePostController(PostController)

	StationController = controllers.NewStationController(initializers.DB)
	StationRouteController = routes.NewRouteStationController(StationController)

	server = gin.Default()
}

func main() {
	config, err := initializers.LoadConfig(".")
	if err != nil {
		log.Fatal("🚀 Could not load environment variables", err)
	}

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*", "http://localhost:3000", config.ClientOrigin}
	//	corsConfig.AllowOrigins = []string{config.ServerOrigin}
	corsConfig.ExposeHeaders = []string{"Access-Control-Allow-Origin"}
	corsConfig.AllowCredentials = true

	server.Use(cors.New(corsConfig))

	router := server.Group("/api")
	router.GET("/healthchecker", func(ctx *gin.Context) {
		message := "Welcome to Golang with Gorm and Postgres"
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": message})
	})

	AuthRouteController.AuthRoute(router)
	UserRouteController.UserRoute(router)
	PostRouteController.PostRoute(router)
	StationRouteController.StationRoute(router)

	// Serve uploaded files for photos
	server.Static("/uploads", "./uploads")

	// Serve over HTTPS
	sslCert := "./certificate.crt"
	sslKey := "./private.key"
	log.Fatal(server.RunTLS(":"+config.ServerPort, sslCert, sslKey))
	// log.Fatal(server.Run(":" + "8000"))
}
