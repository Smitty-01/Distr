package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
)

func Homehandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello from go")
}

func TransactionHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	transactions, err := storage.GetTransactions() // calls
	if err != nil {
		http.Error(w, "failed to create transaction", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(transactions)

}

func addTransactions(w http.ResponseWriter, r *http.Request) {
	var t transaction

	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		log.Println("JSON DECODE ERROR:", err)
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	log.Printf("TRANSACTION RECEIVED: %+v\n", t)

	if !transactionValidator(&t) {
		http.Error(w, "invalid transaction", http.StatusBadRequest)
		return
	}
	dailyLimit, spentToday, txLimit, err := getUserFromRules(t.UserID)

	if err == redis.Nil {

		// Redis doesn't have this user.
		// Get the user's rules from PostgreSQL.

		rules, err := storage.getUserRules(t.UserID)
		if err != nil {
			http.Error(w, "failed to get user rules", http.StatusInternalServerError)
			return
		}

		// Calculate current spending from PostgreSQL.
		spentToday, err = storage.getSpentToday(t.UserID)
		if err != nil {
			http.Error(w, "failed to get spent today", http.StatusInternalServerError)
			return
		}

		dailyLimit = rules.DailyLimit
		txLimit = rules.TransactionLimit

		// Now populate Redis.
		err = setUserRules(
			t.UserID,
			dailyLimit,
			spentToday,
			txLimit,
		)

		if err != nil {
			http.Error(w, "failed to cache user rules", http.StatusInternalServerError)
			return
		}

	} else if err != nil {

		// Some actual Redis error occurred.
		http.Error(w, "redis error", http.StatusInternalServerError)
		return
	}

	approved := decideTransaction(float64(t.Amount), dailyLimit, spentToday, txLimit)
	if !approved {
		t.Status = false
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(t)
		return // Return rejected immediately
	}

	// if t.Amount < 5000 {
	// 	t.Status = false
	// 	w.Header().Set("Content-Type", "application/json")
	// 	json.NewEncoder(w).Encode(t)
	// 	return
	// }

	t.Status = true
	if err := rdb.IncrByFloat(ctx, fmt.Sprintf("user:%d:spent_today", t.UserID), float64(t.Amount)).Err(); err != nil {
		log.Println("Redis Incr Error:", err)
		http.Error(w, "failed to update transaction in redis", http.StatusInternalServerError)
		return
	}
	// we replace creaTransaction with publish transaction movinfg from synchronous to asynchronous
	// err = storage.CreateTransaction(&t)
	// if err != nil {
	// 	log.Println("Database CreateTransaction Error:", err)
	// 	http.Error(w, "failed to create transaction", http.StatusInternalServerError)
	// 	return
	// }

	// w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(t)

	if err = PublishTransaction(&t); err != nil {
		log.Println("Kafka Publish Error:", err)
		http.Error(w, "failed to queue transaction", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// func SingleTransactionHandler(w http.ResponseWriter, r *http.Request) {
// 	id := r.PathValue("id")

// 	for _, t := range mytransactions {
// 		if strconv.Itoa(t.ID) == id {
// 			w.Header().Set("Content-Type", "application/json")
// 			json.NewEncoder(w).Encode(t)
// 			return
// 		}
// 	}

// 	http.Error(w, "transaction not found", http.StatusNotFound)
// }
