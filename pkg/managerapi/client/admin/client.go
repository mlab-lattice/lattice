package admin

import (
	"io"
)

type Client interface {
	Master() MasterClient
}

type MasterClient interface {
	Components() ([]string, error)

	Component(string) MasterComponentClient
}

type MasterComponentClient interface {
	Logs(nodeID string, follow bool) (io.ReadCloser, error)
	Restart(nodeID string) error
}
