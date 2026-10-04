package main

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

func transactionValidator(t *transaction) bool {
	if t.Name == "" || t.Amount <= 0 {
		return false
	}
	return true
}

type Storage struct {
	db *sqlx.DB
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{
		db: db,
	}
} // recieves a db
func (s *Storage) CreateTransaction(t *transaction) error {

	query := `
		INSERT INTO transactions (reference, user_id, name, amount, status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (reference) DO NOTHING
		RETURNING id
	`
	err := s.db.Get(
		&t.ID,
		query,
		t.Reference,
		t.UserID,
		t.Name,
		t.Amount,
		t.Status,
	)
	if err == sql.ErrNoRows {
		// Duplicate event skipped safely
		return nil
	}
	return err
}
func (s *Storage) GetTransactions() ([]transaction, error) {
	var transactions []transaction
	query := `SELECT * From transactions `
	err := s.db.Select(
		&transactions,
		query,
	)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}
func (s *Storage) GetTransactionsById(id int) (transaction, error) {
	var t transaction
	query := `SELECT * From transactions WHERE id=$1` //Find the row where the database's id column equals $1.
	err := s.db.Get(
		&t,
		query,
		id, //  take the from sql and give it to go
	)
	if err != nil {
		return t, err
	}
	return t, nil

}
func (s *Storage) getUserRules(UserID int) (UserRules, error) {
	var rules UserRules
	query := `
        SELECT user_id, daily_limit, transaction_limit
        FROM user_rules
        WHERE user_id = $1
    `
	err := s.db.Get(&rules, query, UserID)

	if err != nil {
		return rules, err
	}

	return rules, nil

}

func (s *Storage) getSpentToday(userID int) (float64, error) {
	var spentToday float64

	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = $1
		AND status = true
		AND created_at >= CURRENT_DATE
	`

	err := s.db.Get(
		&spentToday,
		query,
		userID,
	)

	if err != nil {
		return 0, err
	}

	return spentToday, nil
}
