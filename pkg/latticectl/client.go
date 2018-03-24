package latticectl

import (
	clientv1 "github.com/mlab-lattice/system/pkg/api/client/v1"
)

type ClientFactory func(lattice string) clientv1.Interface
