package initializers

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/denisenkom/go-mssqldb"
)

var DB *sql.DB

func ConnectDB(config *Config) {
	// Build connection string
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;",
		config.DBHost, config.DBUsername, config.DBUserPassword, config.DBPort, config.DBName)

	var err error
	// Create connection pool
	DB, err = sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal("Error creating connection pool: ", err.Error())
	}

	// Ping the database to check the connection
	ctx := context.Background()
	err = DB.PingContext(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println("🚀 Connected Successfully to the Database")
}
