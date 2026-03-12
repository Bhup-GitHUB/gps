package config

import (
	"os"
	"strconv"
)

func String(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func Int(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	number, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return number
}
