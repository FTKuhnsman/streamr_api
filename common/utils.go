package common

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strconv"
)

func GetIntEnvWithDefault(key string, def int) int {
	val, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		log.Printf("Failed to get %s from environment. Defaulting to %d. Error: %v", key, def, err)
		return def
	}
	return val
}

func GetStringEnvWithDefault(key string, def string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Printf("Failed to get %s from environment. Defaulting to %s.", key, def)
		return def
	}
	return val
}

func GenerateRandomHexString(nBytes int) (string, error) {
	bytes := make([]byte, nBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err // return the error if there's a problem generating random bytes
	}
	return hex.EncodeToString(bytes), nil
}
