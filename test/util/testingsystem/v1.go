package testingsystem

import (
	"time"

	"github.com/mlab-lattice/system/test/util/versionservice"

	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	V1ServiceAVersion = "1.0.0"
)

type V1 struct {
	client versionservice.Client
}

func NewV1(serviceAURL string) *V1 {
	return &V1{
		client: versionservice.NewClient(serviceAURL),
	}
}

func (v *V1) Test() error {
	return v.client.CheckStatusAndVersion(V1ServiceAVersion)
}

func (v *V1) Poll(interval, timeout time.Duration) error {
	err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		return false, v.Test()
	})
	if err == nil || err == wait.ErrWaitTimeout {
		return nil
	}

	return err
}
