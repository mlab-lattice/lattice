package networkingprovider

import (
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/networkingprovider/flannel"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/networkingprovider/none"
)

const (
	Flannel = flannel.Flannel
	None    = none.None
)
