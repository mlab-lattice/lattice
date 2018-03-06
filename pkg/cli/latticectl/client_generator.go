package latticectl

import (
	"github.com/mlab-lattice/system/pkg/managerapi/client"
)

type LatticeClientGenerator func(lattice string) client.Interface
