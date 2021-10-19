package request

import (
	"os"
	"strconv"
)

func getMaxConnsPerHost() int {
	i := 20
	if interval := os.Getenv("MAX_CONNECTION_PER_HOST"); interval != "" {
		j, err := strconv.Atoi(interval)
		if err == nil {
			i = j
		}
	}
	return i
}

func getMaxIdleConnsPerHost() int {
	i := 20
	if interval := os.Getenv("MAX_IDLE_CONNECTION_PER_HOST"); interval != "" {
		j, err := strconv.Atoi(interval)
		if err == nil {
			i = j
		}
	}
	return i
}
