package main

import "github.com/jmoiron/sqlx"

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
		INSERT INTO transactions (user_id, name, amount, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	// // func (db *DB) Get(dest interface{}, query string, args ...interface{}) error {
	// 	return Get(db, dest, query, args...)
	// }
	// whne the query runs return the result and ut in t.ID
	err := s.db.Get(
		&t.ID,
		query,
		t.UserID,
		t.Name,
		t.Amount, // give value from user to sql
		t.Status,
	)

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
