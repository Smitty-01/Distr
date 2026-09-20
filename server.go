package main

// ============================================================================
// 1. PACKAGE IMPORTS
// ============================================================================
// Go requires you to import packages to access their standard library utilities.
import (
	"encoding/json" // Used for encoding (converting Go structs -> JSON) and decoding (JSON -> Go structs).
	"fmt"           // Used for I/O formatting like printing to console (Println) or writing to streams (Fprintf).
	"net/http"      // Go's built-in HTTP package providing HTTP server, client, request/response handlers.
	"strconv"       // String conversion library: used here to convert integers to strings (Itoa).
)

// ============================================================================
// 2. DATA STRUCTURES (STRUCT DEFINITIONS)
// ============================================================================

// server struct represents our server configuration.
// Here it holds the network address/port that the server will bind to.
type server struct {
	addr string // e.g., ":8080" means listen on all network interfaces on port 8080.
}

// transaction struct represents a single monetary transaction in our system.
// The `json:"..."` backtick annotations are called "Struct Tags".
// They tell the `encoding/json` package the exact JSON key name to map each field to.
type transaction struct {
	ID     int    `json:"id"`     // Unique numeric ID for the transaction (maps to JSON key "id")
	Name   string `json:"name"`   // Name/description of the transaction (maps to JSON key "name")
	Amount int    `json:"amount"` // Transaction amount in currency units (maps to JSON key "amount")
	Status bool   `json:"status"` // Status flag (true = approved/active, false = rejected/failed)
}

// ============================================================================
// 3. IN-MEMORY STORAGE (GLOBAL STATE)
// ============================================================================

// mytransactions is a package-level slice acting as an in-memory database.
// Note: We use `var` because the short variable declaration `:=` is only allowed inside functions.
// Initial state is an empty slice `nil` / `[]transaction{}`.
var mytransactions []transaction

// ============================================================================
// 4. HELPER / BUSINESS LOGIC FUNCTIONS
// ============================================================================

// createTransaction takes a pointer to a transaction (*transaction) and appends
// the dereferenced transaction value (*t) into the global `mytransactions` slice.
// Using a pointer (*transaction) avoids copying the entire struct in memory when passing it.
func createTransaction(t *transaction) {
	// append() takes an existing slice and one or more elements, returns a new/extended slice.
	mytransactions = append(mytransactions, *t)
}

// transactionValidator performs basic input validation on the transaction.
// Returns:
//   - false: if Name is empty OR Amount is less than or equal to 0.
//   - true:  if the transaction data is valid.
func transactionValidator(t *transaction) bool {
	if t.Name == "" || t.Amount <= 0 {
		return false
	}
	return true
}

// ============================================================================
// 5. HTTP ROUTE HANDLERS
// ============================================================================
// In Go, an HTTP handler function must match the signature:
//    func(w http.ResponseWriter, r *http.Request)
//
// - `w http.ResponseWriter`: An interface used to construct and send the HTTP response back
//                            to the client (status codes, headers, response body).
// - `r *http.Request`: A pointer to the incoming HTTP request struct containing metadata
//                      like HTTP Method (GET, POST), URL path, headers, query parameters,
//                      and the request Body stream.

// ----------------------------------------------------------------------------
// Homehandler handles GET requests to "/"
// ----------------------------------------------------------------------------
func Homehandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprintf formats text and writes it directly to the response writer `w`.
	// The client receives the string "hello from go" with a default status code 200 OK.
	fmt.Fprintf(w, "hello from go")
}

// ----------------------------------------------------------------------------
// TransactionHandler handles GET requests to "/transactions"
// Returns all transactions currently stored in memory as a JSON array.
// ----------------------------------------------------------------------------
func TransactionHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Set the Content-Type header so the client/browser knows the response is JSON.
	w.Header().Set("Content-Type", "application/json")

	// 2. json.NewEncoder(w) creates an encoder that writes directly to the HTTP response stream `w`.
	// 3. .Encode(mytransactions) converts our Go slice of structs into a JSON string
	//    e.g. [{"id":1,"name":"Salary","amount":6000,"status":true}] and flushes it to the client.
	json.NewEncoder(w).Encode(mytransactions)
}

// ----------------------------------------------------------------------------
// addTransactions handles POST requests to "/transactions"
// Reads JSON from the request body, validates it, and saves it if amount >= 5000.
// ----------------------------------------------------------------------------
func addTransactions(w http.ResponseWriter, r *http.Request) {
	// 1. Allocate a zero-value transaction struct to hold incoming payload.
	var t transaction

	// 2. Decode the incoming JSON request body into the address of `t` (&t).
	//    json.NewDecoder reads from the io.Reader stream (r.Body).
	//    If the JSON is malformed (e.g. bad syntax), Decode() returns an error.
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		// http.Error sets HTTP 400 Bad Request and writes "invalid JSON" to the response.
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return // Stop handler execution immediately.
	}

	// 3. Validate business rules: Name cannot be empty, Amount must be > 0.
	if !transactionValidator(&t) {
		http.Error(w, "invalid transaction", http.StatusBadRequest)
		return
	}

	// 4. Business Rule check:
	//    If amount is less than 5000, mark status as false (rejected),
	//    return the transaction to the client, but DO NOT save it in `mytransactions`.
	if t.Amount < 5000 {
		t.Status = false
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(t)
		return
	}

	// 5. If amount >= 5000, save the transaction to our in-memory slice.
	createTransaction(&t)

	// 6. Return the saved transaction back to client as JSON confirmation.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// ----------------------------------------------------------------------------
// SingleTransactionHandler handles GET requests to "/transactions/{id}"
// Extracts the path parameter {id} and searches the in-memory slice.
// ----------------------------------------------------------------------------
func SingleTransactionHandler(w http.ResponseWriter, r *http.Request) {
	// 1. r.PathValue("id") is a Go 1.22+ feature that extracts dynamic URL segments
	//    matched by "/transactions/{id}". It returns the value as a string (e.g. "123").
	id := r.PathValue("id")

	// 2. Iterate through every transaction `t` in our `mytransactions` slice.
	for _, t := range mytransactions {
		// strconv.Itoa(t.ID) converts the integer ID (e.g. 101) into string ("101")
		// so we can compare it with the URL path string `id`.
		if strconv.Itoa(t.ID) == id {
			// Found a match: set header, encode the matched transaction as JSON, and return.
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(t)
			return // Return early so we do not execute the "not found" logic below.
		}
	}

	// 3. If the loop completes without finding any matching ID, return HTTP 404 Not Found.
	http.Error(w, "transaction not found", http.StatusNotFound)
}

// ============================================================================
// 6. MAIN FUNCTION - ENTRY POINT
// ============================================================================
func main() {
	// 1. http.NewServeMux() creates a new HTTP request multiplexer (router).
	//    The multiplexer matches incoming request URLs against registered patterns
	//    and calls the corresponding handler.
	mux := http.NewServeMux()

	// 2. Register route patterns using Go 1.22+ method-and-path routing syntax:
	//    - "GET /": Matches GET requests to root path "/"
	//    - "GET /transactions": Matches GET requests to "/transactions"
	//    - "GET /transactions/{id}": Matches GET requests with a dynamic path variable "{id}"
	//    - "POST /transactions": Matches POST requests to "/transactions"
	mux.HandleFunc("GET /", Homehandler)
	mux.HandleFunc("GET /transactions", TransactionHandler)
	mux.HandleFunc("GET /transactions/{id}", SingleTransactionHandler)
	mux.HandleFunc("POST /transactions", addTransactions)

	// 3. Initialize the server configuration struct with port ":8080".
	srv := &server{
		addr: ":8080",
	}

	// 4. Log to standard output that the server is about to listen.
	fmt.Println("server running on http://localhost" + srv.addr)

	// 5. http.ListenAndServe starts the TCP network listener on `srv.addr` (:8080)
	//    and dispatches incoming HTTP requests to our `mux` router.
	//    NOTE: This is a BLOCKING call. It keeps running indefinitely until terminated
	//    or until an unrecoverable error occurs.
	http.ListenAndServe(srv.addr, mux)
}
