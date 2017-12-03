package client

import (
	"github.com/mlab-lattice/system/pkg/kubernetes/customresource/client/v1"
)

type Interface interface {
	V1() v1.V1Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	v1 *v1.V1Client
}

// V1 retrieves the V1Client
func (c *Clientset) V1() v1.V1Interface {
	return c.v1
}
