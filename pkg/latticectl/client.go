package latticectl

import (
	clientv1 "github.com/mlab-lattice/lattice/pkg/api/client/v1"
)

type ClientFactory func(lattice string) clientv1.Interface
