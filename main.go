package main

import (
	"fmt"
	"log"
	"net/http"
)

var storage *Storage

func main() {

	db, err := connectDB()
	storage = NewStorage(db)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Successfuly connected")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", Homehandler)
	mux.HandleFunc("GET /transactions", TransactionHandler)
	// mux.HandleFunc("GET /transactions/{id}", SingleTransactionHandler)
	mux.HandleFunc("POST /transactions", addTransactions)

	srv := &server{
		addr: ":8080",
	}

	fmt.Println("server running on http://localhost" + srv.addr)

	http.ListenAndServe(srv.addr, mux)
}
