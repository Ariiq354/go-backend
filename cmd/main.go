package main

import (
	"database/sql"
	"fmt"
	"log"
	"test-api/cmd/api"
	"test-api/db"

	"github.com/gin-gonic/gin"
)

func main() {
	db.InitDB()

	database := db.GetDB()

	initStorage(database)

	r := gin.Default()

	api.SetupRoutes(r)

	r.Run(":3001")
}

func initStorage(db *sql.DB) {
	err := db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database: ", err)
	}
	fmt.Println("Connected to MySQL database!")
}
