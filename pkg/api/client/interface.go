package client

import (
	"github.com/mlab-lattice/lattice/pkg/api/client/v1"
)

type Interface interface {
	Health() (bool, error)

	V1() v1.Interface
}
