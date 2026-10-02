package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
)

var storage *Storage
var rdb *redis.Client

func main() {

	db, err := connectDB()   // db initilization
	storage = NewStorage(db) //  storage struct
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Successfuly connected")
	}
	// redis client
	rdb = NewRedis() // redis client initilization
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Redis connection failed:", err)
	} else {
		log.Println("Successfully connected to Redis")
	}
	// kafka client
	InitKafkaProducer()
	defer kafkaWriter.Close()
	go StartKafkaConsumer(storage)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", Homehandler)
	mux.HandleFunc("GET /transactions", TransactionHandler)
	// mux.HandleFunc("GET /transactions/{id}", SingleTransactionHandler)
	mux.HandleFunc("POST /transactions", addTransactions)

	srv := &server{
		addr: ":8080",
	}

	fmt.Println("server running on http://localhost" + srv.addr)

	http.ListenAndServe(srv.addr, mux) // server start
}
