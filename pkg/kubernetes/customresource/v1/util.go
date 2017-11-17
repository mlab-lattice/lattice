package v1

import (
	"fmt"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"
)

func GetProviderFromConfigSpec(config *ConfigSpec) (string, error) {
	if config == nil {
		return "", fmt.Errorf("cannot get provider from nil config")
	}

	providers := []string{}

	if config.Provider.Local != nil {
		providers = append(providers, coreconstants.ProviderLocal)
	}

	if config.Provider.AWS != nil {
		providers = append(providers, coreconstants.ProviderAWS)
	}

	if len(providers) == 0 {
		return "", fmt.Errorf("no provider config set")
	}

	if len(providers) > 1 {
		return "", fmt.Errorf("multiple provider configs set")
	}

	return providers[0], nil
}
