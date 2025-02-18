package db

import (
	"database/sql"
	"log"

	"github.com/go-sql-driver/mysql"
)

func NewMySQL(cfg mysql.Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	return db, nil
}

func GetDB() *sql.DB {
	db, err := NewMySQL(mysql.Config{
		User:   "root",
		Passwd: "root",
		Addr:   "127.0.0.1:3306",
		DBName: "article",
		Net:    "tcp",
	})
	if err != nil {
		log.Fatal(err)
	}

	return db
}
