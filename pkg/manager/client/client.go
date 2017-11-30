package client

import (
	"github.com/mlab-lattice/system/pkg/manager/client/admin"
	"github.com/mlab-lattice/system/pkg/manager/client/user"
)

func NewAdminClient(managerAPIURL string) *admin.Client {
	return admin.NewClient(managerAPIURL)
}

func NewUserClient(managerAPIURL string) *user.Client {
	return user.NewClient(managerAPIURL)
}
