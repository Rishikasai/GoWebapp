package models

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func Init() (db *sql.DB) {
	db, err := sql.Open("mysql", "root:rishika@(127.0.0.1:3306)/dbname")
	if err != nil {
		panic(err.Error())
	}
	return db
}
