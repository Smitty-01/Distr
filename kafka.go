// Producer and Consumer in go

package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

var (
	kafkaTopic  = "transactions"
	kafkaBroker = "localhost:9092"
	kafkaWriter *kafka.Writer
)

// InitKafkaProducer initializes the Kafka writer connection pool
func InitKafkaProducer() {
	kafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(kafkaBroker),
		Topic:        kafkaTopic,
		Balancer:     &kafka.LeastBytes{}, //LeastBytes tries to distribute messages based on the amount of data being sent.
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		BatchTimeout: 5 * time.Millisecond, // Flushes in 5ms instead of waiting 1000ms!
	}
}

// PublishTransaction sends an approved transaction event to Kafka in <1ms
func PublishTransaction(t *transaction) error {
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(string(rune(t.UserID))),
		Value: data,
	}
	return kafkaWriter.WriteMessages(context.Background(), msg)
}

// StartKafkaConsumer runs in a background goroutine, reading messages with manual commits and retries
func StartKafkaConsumer(storage *Storage) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaBroker},
		Topic:    kafkaTopic,
		GroupID:  "transaction-audit-group",
		MinBytes: 1,    // Read immediately without waiting for 10KB buffer!
		MaxBytes: 10e6, // 10MB
	})

	log.Println("Kafka Consumer started (with Manual Commits & Retry)...")

	for {
		// 1. Fetch message WITHOUT committing offset
		ctx := context.Background()
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			log.Println("Error fetching Kafka message:", err)
			continue
		}

		var t transaction
		if err := json.Unmarshal(m.Value, &t); err != nil {
			log.Println("Error unmarshaling transaction from Kafka:", err)
			// Commit malformed message so it doesn't block the queue
			reader.CommitMessages(ctx, m)
			continue
		}

		// 2. Retry loop for PostgreSQL insertion (up to 3 attempts with backoff)
		var dbErr error
		for attempt := 1; attempt <= 3; attempt++ {
			dbErr = storage.CreateTransaction(&t)
			if dbErr == nil {
				break
			}
			log.Printf("Attempt %d: DB insert failed for user %d: %v. Retrying in %dms...\n",
				attempt, t.UserID, dbErr, attempt*200)
			time.Sleep(time.Duration(attempt*200) * time.Millisecond)
		}

		// 3. ONLY commit offset if PostgreSQL write succeeded!
		if dbErr == nil {
			if err := reader.CommitMessages(ctx, m); err != nil {
				log.Println("Failed to commit Kafka offset:", err)
			} else {
				kafkaAuditSavedTotal.Inc()
				log.Printf("DB Audit Log Saved & Offset Committed: Txn %s for User %d\n", t.Reference, t.UserID)
			}
		} else {
			log.Printf("CRITICAL: Failed to write to DB after 3 attempts. Offset NOT committed for txn %s\n", t.Reference)
		}
	}
}
