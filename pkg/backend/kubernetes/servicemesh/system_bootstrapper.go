package servicemesh

import (
	"fmt"

	systembootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/servicemesh/envoy"
)

type SystemBootstrapperOptions struct {
	Envoy *envoy.SystemBootstrapperOptions
}

func NewSystemBootstrapper(options *SystemBootstrapperOptions) (systembootstrapper.Interface, error) {
	if options.Envoy != nil {
		return envoy.NewSystemBootstrapper(options.Envoy), nil
	}

	return nil, fmt.Errorf("must provide service mesh options")
}
