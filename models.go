package main

type server struct {
	addr string
}

type transaction struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Amount int    `json:"amount"`
	Status bool   `json:"status"`
}
