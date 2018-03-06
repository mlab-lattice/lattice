package latticectl

import (
	"github.com/mlab-lattice/system/pkg/managerapi/client"
)

type ClientFactory func(lattice string) client.Interface
