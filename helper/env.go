package helper

import "os"

func EnvOrDefault(key, def string) string {
	if env := os.Getenv(key); env != "" {
		return env
	}
	return def
}
