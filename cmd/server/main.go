package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"desafio4-rate-limit/internal/config"
	"desafio4-rate-limit/internal/infra/redisstore"
	"desafio4-rate-limit/internal/limiter"
	"desafio4-rate-limit/internal/middleware"

	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer redisClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	rateLimiter := limiter.New(redisstore.New(redisClient), cfg.Limiter)
	rlMiddleware := middleware.NewRateLimiter(rateLimiter)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"ok"}`))
	})

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, rlMiddleware.Handle(mux)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
