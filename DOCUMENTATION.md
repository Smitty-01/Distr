# Complete Go Backend Server Documentation & Learning Guide

This document is your complete, permanent guide to understanding every single concept, line of code, memory flow, and request lifecycle in `server.go`.

---

## Table of Contents
1. [High-Level Architecture & How Go Handles HTTP](#1-high-level-architecture--how-go-handles-http)
2. [Line-by-Line Code Breakdown](#2-line-by-line-code-breakdown)
3. [The HTTP Request & Response Lifecycle (Trace)](#3-the-http-request--response-lifecycle-trace)
4. [Core Go Concepts Explained (Pointers, Struct Tags, Slices, Multiplexing)](#4-core-go-concepts-explained)
5. [Endpoints Summary & Testing Cheat Sheet](#5-endpoints-summary--testing-cheat-sheet)

---

## 1. High-Level Architecture & How Go Handles HTTP

```
Client (Browser / Postman / cURL)
             │
             │ HTTP Request (e.g. POST /transactions)
             ▼
   [ http.ListenAndServe(":8080", mux) ]   <-- TCP Socket Listener
             │
             │ Creates a new Go Goroutine (Lightweight Thread) per request
             ▼
      [ http.ServeMux ]                    <-- Router / Multiplexer
             │ Matches URL & Method
             ▼
     Handler Function (e.g. addTransactions(w, r))
       ├── Reads & Decodes `r.Body` (JSON -> struct)
       ├── Validates & Applies Business Logic
       ├── Reads / Writes to `mytransactions` (In-Memory Slice)
       └── Encodes JSON & Writes status/headers to `w` (ResponseWriter)
             │
             │ HTTP Response (Status 200/400/404 + JSON)
             ▼
Client Receives Response
```

---

## 2. Line-by-Line Code Breakdown

### Section 1: Package & Imports
```go
package main
```
- **What it means:** Defines this file as part of the `main` package. The `main` package produces an executable binary when compiled and contains the starting function `main()`.

```go
import (
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
)
```
- `"encoding/json"`: Provides `json.NewDecoder` and `json.NewEncoder` to convert between JSON text and Go structs.
- `"fmt"`: Provides I/O formatting (e.g., `fmt.Println` to print to the terminal, and `fmt.Fprintf` to write directly to an HTTP stream).
- `"net/http"`: Go's standard library for building HTTP servers and clients.
- `"strconv"`: String conversion library (e.g., `strconv.Itoa` converts integer `int` to `string`).

---

### Section 2: Data Structures (Structs)

```go
type server struct {
    addr string
}
```
- **What it means:** A custom data structure (struct) to hold server settings. `addr` is a string specifying `:8080` (port 8080 on all local interfaces).

```go
type transaction struct {
    ID     int    `json:"id"`
    Name   string `json:"name"`
    Amount int    `json:"amount"`
    Status bool   `json:"status"`
}
```
- **What it means:** Defines a blueprint for a transaction.
- **Why fields start with Capital Letters (`ID`, `Name`, `Amount`, `Status`):** In Go, capitalized identifiers are **exported** (public). The `json` package can only read and write exported fields.
- **What are Struct Tags (`json:"id"`)?** They instruct Go's JSON parser to map JSON keys like `"id"` to Go field `ID`.

---

### Section 3: In-Memory Storage

```go
var mytransactions []transaction
```
- **What it means:** A global slice of `transaction` structs. Slices in Go are dynamically-sized arrays.
- **Why `var` instead of `:=`?** The shorthand `:=` is only valid inside function bodies. At the package level, variable declarations must use `var`.

---

### Section 4: Helper & Business Logic Functions

```go
func createTransaction(t *transaction) {
    mytransactions = append(mytransactions, *t)
}
```
- **Parameter `t *transaction`:** `t` is a pointer (memory address) to a `transaction`. Passing a pointer avoids copying the entire struct in memory.
- `*t` dereferences the pointer to get the underlying `transaction` value.
- `append(mytransactions, *t)` adds the item to the slice and updates `mytransactions`.

```go
func transactionValidator(t *transaction) bool {
    if t.Name == "" || t.Amount <= 0 {
        return false
    }
    return true
}
```
- **Purpose:** Verifies required data. A transaction is invalid if the name is empty or the amount is 0 or negative.

---

### Section 5: HTTP Route Handlers

All Go HTTP handlers have the signature:
`func(w http.ResponseWriter, r *http.Request)`
- `w http.ResponseWriter`: Used to construct and send the HTTP response (headers, HTTP status code, and response body).
- `r *http.Request`: Contains everything sent by the client (HTTP method, URL, headers, URL parameters, body stream).

#### 1. Home Handler
```go
func Homehandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "hello from go")
}
```
- `fmt.Fprintf(w, ...)` writes the string directly to the response writer. The client receives `"hello from go"` with HTTP 200 OK.

#### 2. Get All Transactions
```go
func TransactionHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(mytransactions)
}
```
- Sets the HTTP Header `Content-Type: application/json`.
- `json.NewEncoder(w).Encode(mytransactions)` turns the slice into a JSON array and writes it directly to the response stream.

#### 3. Create / Add Transaction
```go
func addTransactions(w http.ResponseWriter, r *http.Request) {
    var t transaction
    if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }
```
- Creates an empty `t transaction`.
- `json.NewDecoder(r.Body).Decode(&t)` parses the incoming JSON request body into `t`.
- If decoding fails (invalid JSON format), it responds with `400 Bad Request` and stops execution (`return`).

```go
    if !transactionValidator(&t) {
        http.Error(w, "invalid transaction", http.StatusBadRequest)
        return
    }
```
- Runs validation; rejects empty names or amount <= 0.

```go
    if t.Amount < 5000 {
        t.Status = false
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(t)
        return
    }

    createTransaction(&t)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(t)
}
```
- **Business Rule:** If amount is under 5000, set `Status = false` and return it without adding it to `mytransactions`.
- If amount >= 5000, calls `createTransaction(&t)` to append it to `mytransactions`, and returns the saved object.

#### 4. Get Single Transaction by ID
```go
func SingleTransactionHandler(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")

    for _, t := range mytransactions {
        if strconv.Itoa(t.ID) == id {
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(t)
            return
        }
    }
    http.Error(w, "transaction not found", http.StatusNotFound)
}
```
- `r.PathValue("id")`: Extracts the path variable from `{id}` in the URL (Go 1.22+ routing).
- Loops through all transactions; `strconv.Itoa(t.ID)` converts integer ID to string to compare with `id`.
- If found, writes JSON and returns.
- If loop finishes with no match, returns `404 Not Found`.

---

### Section 6: Main Entrypoint & Server Startup

```go
func main() {
    mux := http.NewServeMux()
```
- `http.NewServeMux()` creates the router (multiplexer).

```go
    mux.HandleFunc("GET /", Homehandler)
    mux.HandleFunc("GET /transactions", TransactionHandler)
    mux.HandleFunc("GET /transactions/{id}", SingleTransactionHandler)
    mux.HandleFunc("POST /transactions", addTransactions)
```
- Registers URL routes and HTTP methods to their respective handler functions.

```go
    srv := &server{
        addr: ":8080",
    }
    fmt.Println("server running on http://localhost" + srv.addr)
    http.ListenAndServe(srv.addr, mux)
}
```
- `http.ListenAndServe(":8080", mux)`: Starts listening on TCP port 8080 and passes incoming connections to `mux`. This blocks indefinitely while the server is active.

---

## 3. Core Go Concepts Reference

| Concept | Why & How It Is Used |
| :--- | :--- |
| **Pointers (`*` and `&`)** | `&t` gives the memory address of `t`. `*transaction` is a pointer type. Using pointers prevents copying structs across function calls and allows in-place mutation. |
| **Struct Tags (`json:"..."`)** | Metadata attached to struct fields. Tells Go's JSON parser what keys to use when encoding/decoding JSON. |
| **`http.ResponseWriter`** | An interface that implements writing headers (`WriteHeader`), body (`Write`), and setting headers (`Header().Set()`). |
| **`*http.Request`** | A struct that contains all incoming request information (method, URL, headers, body stream). |
| **Goroutines in HTTP** | Go automatically spawns a separate lightweight thread (goroutine) for every incoming HTTP request. Multiple requests run concurrently. |

---

## 4. Endpoints Summary & Testing Cheat Sheet

### 1. Test Home Route
- **Request:** `GET http://localhost:8080/`
- **Response:** `hello from go` (Text)

### 2. Add New Transaction (Valid & Approved)
- **Request:** `POST http://localhost:8080/transactions`
- **Headers:** `Content-Type: application/json`
- **Body:**
```json
{
  "id": 1,
  "name": "Salary",
  "amount": 7500,
  "status": true
}
```
- **Response Code:** `200 OK`
- **Response Body:** `{"id":1,"name":"Salary","amount":7500,"status":true}`

### 3. Add Transaction (Under 5000 -> Rejected)
- **Request:** `POST http://localhost:8080/transactions`
- **Body:**
```json
{
  "id": 2,
  "name": "Coffee",
  "amount": 150
}
```
- **Response Body:** `{"id":2,"name":"Coffee","amount":150,"status":false}` *(Not saved to storage)*

### 4. Fetch All Transactions
- **Request:** `GET http://localhost:8080/transactions`
- **Response:** `[{"id":1,"name":"Salary","amount":7500,"status":true}]`

### 5. Fetch Single Transaction by ID
- **Request:** `GET http://localhost:8080/transactions/1`
- **Response:** `{"id":1,"name":"Salary","amount":7500,"status":true}`
