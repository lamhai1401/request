package tests

import (
	"os"
	"testing"

	"github.com/lamhai1401/gologs/logs"
	"github.com/lamhai1401/request/request"
)

// func BenchmarkAPI(b *testing.B) {
// 	runGet(1, b.N)
// }

func TestAPI(t *testing.T) {
	timeOut := 1
	os.Setenv("MAX_CONNECTION_PER_HOST", "20")
	os.Setenv("MAX_IDLE_CONNECTION_PER_HOST", "20")

	url := "http://localhost:4000/api/info/pc_id"
	api := request.NewAPI(timeOut)

	t.Run("Test Close", func(t *testing.T) {
		api.Close()
		api.GET(&url, nil)
	})

	t.Run("Test Get Timeout", func(t *testing.T) {
		for i := 0; i <= 10000; i++ {
			apiTimeout := request.NewAPI(10)
			resp := apiTimeout.GET(&url, nil)
			result, err := api.ReadResponse(<-resp)
			if err != nil {
				logs.Error(err.Error())
			} else {
				logs.Info(string(result), "\n")
			}
		}
	})
}
