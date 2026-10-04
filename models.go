package main

type server struct {
	addr string
}

type transaction struct {
	ID        int    `json:"id" db:"id"`
	Reference string `json:"reference" db:"reference"`
	UserID    int    `json:"user_id" db:"user_id"`
	Name      string `json:"name" db:"name"`
	Amount    int    `json:"amount" db:"amount"`
	Status    bool   `json:"status" db:"status"`
}
type UserRules struct {
	UserID           int     `db:"user_id"`
	DailyLimit       float64 `db:"daily_limit"`
	TransactionLimit float64 `db:"transaction_limit"`
}

// spenttoday does not make sense redis cache should have acc balance as well
// then we check if amount < balance and send errors
