package main

import (
	"log"
	"os"
	"test-api/db"

	mysqlCfg "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	db, err := db.NewMySQL(mysqlCfg.Config{
		User:   "root",
		Passwd: "root",
		Addr:   "127.0.0.1:3306",
		DBName: "article",
		Net:    "tcp",
	})
	if err != nil {
		log.Fatal(err)
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://cmd/migrate/migrations",
		"mysql",
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	cmd := os.Args[(len(os.Args) - 1)]
	if cmd == "up" {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {

		}
	}
	if cmd == "down" {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {

		}
	}
}
