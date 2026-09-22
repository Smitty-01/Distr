package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func NewRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

// fetch transaction from redis
func getUserRules(userID int) (dailyLimit, spentToday, txLimit float64, err error) {
	dailyLimit, err = rdb.Get(ctx, fmt.Sprintf("user:%d:daily_limit", userID)).Float64()
	if err != nil {
		return 0, 0, 0, err
	}
	spentToday, err = rdb.Get(ctx, fmt.Sprintf("user:%d:spent_today", userID)).Float64()
	if err != nil {
		return 0, 0, 0, err
	}
	txLimit, err = rdb.Get(ctx, fmt.Sprintf("user:%d:transaction_limit", userID)).Float64()
	if err != nil {
		return 0, 0, 0, err
	}
	return dailyLimit, spentToday, txLimit, nil
}

func decideTransaction(amount, dailyLimit, spentToday, txLimit float64) bool {
	if amount > txLimit {
		return false
	}
	if spentToday+amount > dailyLimit {
		return false
	}
	return true
}
