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
		Balancer:     &kafka.LeastBytes{},
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

// StartKafkaConsumer runs in a background goroutine, reading messages and saving to DB
func StartKafkaConsumer(storage *Storage) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaBroker},
		Topic:    kafkaTopic,
		GroupID:  "transaction-audit-group",
		MinBytes: 1, // Read immediately without waiting for 10KB buffer!
		MaxBytes: 10e6, // 10MB
	})

	log.Println("Kafka Consumer started and listening for transactions...")

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Println("Error reading Kafka message:", err)
			continue
		}

		var t transaction
		if err := json.Unmarshal(m.Value, &t); err != nil {
			log.Println("Error unmarshaling transaction from Kafka:", err)
			continue
		}

		// Asynchronously save to PostgreSQL!
		if err := storage.CreateTransaction(&t); err != nil {
			log.Printf("Failed to insert transaction to DB for user %d: %v\n", t.UserID, err)
		} else {
			log.Printf("DB Audit Log Saved: Transaction ID %d for User %d\n", t.ID, t.UserID)
		}
	}
}
