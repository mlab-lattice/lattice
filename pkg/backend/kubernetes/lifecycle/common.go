package lifecycle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/util/wait"
)

func pollForSystemEnvironmentReadiness(address string) error {
	client := &http.Client{
		Timeout: time.Duration(time.Second * 5),
	}

	return wait.Poll(1*time.Second, 300*time.Second, func() (bool, error) {
		resp, err := client.Get("http://" + address + "/status")
		if err != nil {
			fmt.Printf("Got error polling /status: %v\n", err)
			return false, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			// FIXME: print these out at a certain verbosity
			fmt.Printf("Got status code %v from /status\n", resp.StatusCode)
			return false, nil
		}

		return true, nil
	})
}

func tearDownAndWaitForSuccess(address string) error {
	client := &http.Client{
		Timeout: time.Duration(time.Second * 5),
	}

	resp, err := client.Post("http://"+address+"/namespaces/lattice-user-system/teardowns", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("got unexpected status code %v when enqueueing teardown", resp.StatusCode)
	}

	teardownResponse := &struct {
		TeardownID string `json:"teardownId"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(teardownResponse)
	if err != nil {
		return err
	}

	return wait.Poll(5*time.Second, 300*time.Second, func() (bool, error) {
		resp, err := client.Get("http://" + address + "/namespaces/lattice-user-system/teardowns/" + teardownResponse.TeardownID)
		if err != nil {
			// FIXME: print these out at a certain verbosity
			fmt.Printf("Got error polling teardown %v: %v\n", teardownResponse.TeardownID, err)
			return false, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			// FIXME: print these out at a certain verbosity
			fmt.Printf("Got status code %v polling teardown %v\n", resp.StatusCode, teardownResponse.TeardownID)
			return false, nil
		}

		teardown := &types.SystemTeardown{}
		err = json.NewDecoder(resp.Body).Decode(teardown)
		if err != nil {
			return false, err
		}

		switch teardown.State {
		case constants.SystemTeardownStateSucceeded:
			return true, nil
		case constants.SystemTeardownStateFailed:
			return false, fmt.Errorf("teardown %v failed", teardownResponse.TeardownID)
		case constants.SystemTeardownStatePending, constants.SystemTeardownStateInProgress:
			fmt.Printf("teardown %v in state %v\n", teardownResponse.TeardownID, teardown.State)
			return false, nil
		default:
			panic("unreachable")
		}
	})
}
