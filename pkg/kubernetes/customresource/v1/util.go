package v1

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/constants"
)

func GetProviderFromConfigSpec(config *ConfigSpec) (string, error) {
	if config == nil {
		return "", fmt.Errorf("cannot get provider from nil config")
	}

	providers := []string{}

	if config.Provider.Local != nil {
		providers = append(providers, constants.ProviderLocal)
	}

	if config.Provider.AWS != nil {
		providers = append(providers, constants.ProviderAWS)
	}

	if len(providers) == 0 {
		return "", fmt.Errorf("no provider config set")
	}

	if len(providers) > 1 {
		return "", fmt.Errorf("multiple provider configs set")
	}

	return providers[0], nil
}
