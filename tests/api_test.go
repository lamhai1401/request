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

	url := "https://signal-controller-staging.quickom.com/signaling/container/list"
	api := request.NewAPI(timeOut)

	t.Run("Test Close", func(t *testing.T) {
		api.Close()
		api.GET(&url, nil)
	})

	t.Run("Test Get Timeout", func(t *testing.T) {
		apiTimeout := request.NewAPI(0)
		resp := apiTimeout.GET(&url, nil)
		result := api.GetResult(resp)
		if result.Err != nil {
			t.Log(result.Err.Error())
		} else {
			if result.StatusCode != 200 {
				logs.Error("Status code is ", result.StatusCode)
				return
			}
			t.Error("Call api success")
		}
	})
}
