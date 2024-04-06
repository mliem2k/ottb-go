package main

import (
	"fmt"
	"log"

	"github.com/mliem2k/ottb-go/initializers"
	"github.com/mliem2k/ottb-go/models"
)

func init() {
	config, err := initializers.LoadConfig(".")
	if err != nil {
		log.Fatal("🚀 Could not load environment variables", err)
	}

	initializers.ConnectDB(&config)
}

func main() {
	initializers.DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	initializers.DB.AutoMigrate(&models.User{}, &models.Post{}, &models.Station{})
	fmt.Println("👍 Migration complete")
}
