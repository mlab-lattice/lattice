package lifecycle

import (
	"net/http"
	"time"

	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
)

func pollForSystemEnvironmentReadiness(address string) error {
	client := &http.Client{
		Timeout: time.Duration(time.Second * 5),
	}

	return wait.Poll(1*time.Second, 300*time.Second, func() (bool, error) {
		req, err := http.NewRequest(http.MethodGet, "http://"+address+"/status", nil)
		if err != nil {
			return false, err
		}

		resp, err := client.Do(req)
		if err != nil {
			// FIXME: print these out at a certain verbosity
			fmt.Printf("Got error polling SystemEnvironmentManager: %v\n", err)
			return false, nil
		}

		if resp.StatusCode != http.StatusOK {
			// FIXME: print these out at a certain verbosity
			fmt.Printf("Got status code %v from /status\n", resp.StatusCode)
			return false, nil
		}

		return true, nil
	})
}
