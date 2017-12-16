package resources

import (
	"github.com/mlab-lattice/system/pkg/managerapi/client/user"
)

type BuildClient struct {
	RestClient    user.NamespaceClient
	DisplayAsJSON bool
}
