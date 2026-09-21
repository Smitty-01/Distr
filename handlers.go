package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Homehandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello from go")
}

func TransactionHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	transactions, err := storage.GetTransactions()
	if err != nil {
		http.Error(w, "failed to create transaction", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(transactions)

}

func addTransactions(w http.ResponseWriter, r *http.Request) {
	var t transaction

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if !transactionValidator(&t) {
		http.Error(w, "invalid transaction", http.StatusBadRequest)
		return
	}

	if t.Amount < 5000 {
		t.Status = false
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(t)
		return
	}

	err := storage.CreateTransaction(&t)
	if err != nil {
		http.Error(w, "failed to create transaction", http.StatusInternalServerError)
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
