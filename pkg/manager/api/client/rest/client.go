package rest

import (
	"github.com/mlab-lattice/system/pkg/manager/api/client/rest/admin"
	"github.com/mlab-lattice/system/pkg/manager/api/client/rest/user"
)

func NewAdminClient(managerAPIURL string) *admin.Client {
	return admin.NewClient(managerAPIURL)
}

func NewUserClient(managerAPIURL string) *user.Client {
	return user.NewClient(managerAPIURL)
}
