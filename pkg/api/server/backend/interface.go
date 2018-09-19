package backend

import "github.com/mlab-lattice/lattice/pkg/api/server/backend/v1"

type Interface interface {
	V1() v1.Interface
}
