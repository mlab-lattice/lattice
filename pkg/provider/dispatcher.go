package provider

import (
	"fmt"

	"github.com/mlab-lattice/core/pkg/constants"

	localprovider "github.com/mlab-lattice/kubernetes-integration/pkg/provider/local"
)

func GetProvider(providerName string) Interface {
	switch providerName {
	case constants.ProviderLocal:
		return Interface(localprovider.NewProvider())
	default:
		panic(fmt.Sprintf("invalid provider %v", providerName))
	}
}
