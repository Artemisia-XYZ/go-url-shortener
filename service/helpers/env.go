package helpers

import "os"

func Getenv(key string, def string) string {
	if val, found := os.LookupEnv(key); found {
		return val
	}
	return def
}
