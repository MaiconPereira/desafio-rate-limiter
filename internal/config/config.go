package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"desafio4-rate-limit/internal/limiter"
)

type Config struct {
	AppPort       string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	Limiter       limiter.Config
}

func Load() (Config, error) {
	redisDB, err := envInt("REDIS_DB", 0)
	if err != nil {
		return Config{}, err
	}

	ipRPS, err := envInt64("RATE_LIMIT_IP_RPS", 10)
	if err != nil {
		return Config{}, err
	}

	tokenRPS, err := envInt64("RATE_LIMIT_TOKEN_RPS", 100)
	if err != nil {
		return Config{}, err
	}

	blockDuration, err := envDuration("RATE_LIMIT_BLOCK_DURATION", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}

	overrides, err := parseTokenOverrides(os.Getenv("RATE_LIMIT_TOKEN_OVERRIDES"))
	if err != nil {
		return Config{}, err
	}

	return Config{
		AppPort:       envString("APP_PORT", "8080"),
		RedisAddr:     envString("REDIS_ADDR", "redis:6379"),
		RedisPassword: envString("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,
		Limiter: limiter.Config{
			IPRequestsPerSecond:    ipRPS,
			TokenRequestsPerSecond: tokenRPS,
			TokenOverrides:         overrides,
			BlockDuration:          blockDuration,
			Window:                 time.Second,
		},
	}, nil
}

func envString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}

	return parsed, nil
}

func envInt64(key string, fallback int64) (int64, error) {
	parsed, err := envInt(key, int(fallback))
	return int64(parsed), err
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration, for example 5m or 10s: %w", key, err)
	}

	return duration, nil
}

func parseTokenOverrides(value string) (map[string]int64, error) {
	overrides := make(map[string]int64)
	value = strings.TrimSpace(value)
	if value == "" {
		return overrides, nil
	}

	items := strings.Split(value, ",")
	for _, item := range items {
		parts := strings.Split(strings.TrimSpace(item), "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("RATE_LIMIT_TOKEN_OVERRIDES item %q must use token=limit", item)
		}

		limit, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("RATE_LIMIT_TOKEN_OVERRIDES item %q has invalid limit: %w", item, err)
		}

		overrides[strings.TrimSpace(parts[0])] = limit
	}

	return overrides, nil
}
