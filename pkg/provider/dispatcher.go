package provider

import (
	"fmt"

	"github.com/mlab-lattice/core/pkg/constants"
	coretypes "github.com/mlab-lattice/core/pkg/types"

	localprovider "github.com/mlab-lattice/kubernetes-integration/pkg/provider/local"
)

func GetProvider(providerName string) Interface {
	provider := coretypes.Provider(providerName)
	switch provider {
	case constants.ProviderLocal:
		return Interface(localprovider.NewProvider())
	default:
		panic(fmt.Sprintf("invalid provider %v", providerName))
	}
}
