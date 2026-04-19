package env

import "os"

func GetString(key string, fallback string) string {
	if vall := os.Getenv(key); vall != "" {
		return vall
	}
	return fallback
}
