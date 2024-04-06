package main

import (
	"fmt"
	"log"

	"github.com/mliem2k/ottb-go/initializers"
)

func init() {
	config, err := initializers.LoadConfig(".")
	if err != nil {
		log.Fatal("🚀 Could not load environment variables", err)
	}

	initializers.ConnectDB(&config)
}

func main() {
	initializers.DB.Exec("CREATE DATABASE your_database_name")
	initializers.DB.Exec("USE your_database_name")
	initializers.DB.Exec("CREATE TABLE IF NOT EXISTS users (id INT PRIMARY KEY, name VARCHAR(255), email VARCHAR(255))")
	initializers.DB.Exec("CREATE TABLE IF NOT EXISTS posts (id INT PRIMARY KEY, title VARCHAR(255), content TEXT)")
	initializers.DB.Exec("CREATE TABLE IF NOT EXISTS stations (id INT PRIMARY KEY, name VARCHAR(255), location VARCHAR(255))")
	fmt.Println("👍 Migration complete")
}
