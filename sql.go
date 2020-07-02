package main

import (
	"database/sql"
	"log"
	"strconv"

	_ "github.com/lib/pq"
)

// DB connection
var DB *sql.DB

func connectToSQL() {
	db, err := sql.Open(
		"postgres",
		"postgres://"+Config.PsqlUser+":"+Config.PsqlPass+"@"+Config.PsqlHost+":"+strconv.Itoa(int(Config.PsqlPort))+"/"+Config.PsqlDb+"?sslmode=disable",
	)
	if err != nil {
		log.Fatal(err)
	}
	DB = db
}
