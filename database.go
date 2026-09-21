package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func connectDB() (*sqlx.DB, error) {
	dsn := "host=localhost port=5432 user=postgres password=ashmit dbname=transactions sslmode=disable"
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// db holds a pointer to sqlx.DB
// sql.DB is a package to manage database
