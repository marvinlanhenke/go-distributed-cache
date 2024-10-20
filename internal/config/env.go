package config

import (
	"os"
	"strconv"
)

// Retrieves the value of the environment variable identified by 'key'.
// If the environment variable is not set, it returns the provided 'fallback' value.
func getString(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return val
}

// Retrieves the value of the environment variable identified by 'key' and converts it to an integer.
// If the environment variable is not set or the value cannot be parsed as an integer, it returns the provided 'fallback' value.
func getInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}

	return valAsInt
}
