package cloudprovider

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/aws"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	systembootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
)

type SystemBootstrapperOptions struct {
	AWS   *aws.SystemBootstrapperOptions
	Local *local.SystemBootstrapperOptions
}

func NewSystemBootstrapper(options *SystemBootstrapperOptions) (systembootstrapper.Interface, error) {
	if options.AWS != nil {
		return aws.NewSystemBootstrapper(options.AWS), nil
	}

	if options.Local != nil {
		return local.NewSystemBootstrapper(options.Local), nil
	}

	return nil, fmt.Errorf("must provide cloud provider options")
}
