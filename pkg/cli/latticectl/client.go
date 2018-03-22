package latticectl

import (
	"github.com/mlab-lattice/system/pkg/apiserver/client"
)

type ClientFactory func(lattice string) client.Interface
