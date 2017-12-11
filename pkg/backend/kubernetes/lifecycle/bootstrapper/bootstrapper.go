package bootstrapper

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/bootstrapper/base"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/bootstrapper/cloud"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/bootstrapper/local"

	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Interface interface {
	Bootstrap() ([]interface{}, error)
}

type BaseBootstrapper interface {
	BaseBootstrap() ([]interface{}, error)
}

type LocalBootstrapper interface {
	LocalBootstrap() ([]interface{}, error)
}

type CloudBootstrapper interface {
	CloudBootstrap() ([]interface{}, error)
}

type Options struct {
	DryRun           bool
	Config           crv1.ConfigSpec
	MasterComponents base.MasterComponentOptions
	Networking       *cloud.NetworkingOptions
}

func NewBootstrapper(options *Options, kubeConfig *rest.Config) (Interface, error) {
	if options == nil {
		return nil, fmt.Errorf("options required")
	}

	var kubeClient kubeclientset.Interface
	var latticeClient latticeclientset.Interface
	var err error
	if !options.DryRun {
		kubeClient, err = kubeclientset.NewForConfig(kubeConfig)
		if err != nil {
			return nil, err
		}

		latticeClient, err = latticeclientset.NewForConfig(kubeConfig)
		if err != nil {
			return nil, err
		}
	}

	if options.Config.Provider.Local != nil {
		return NewLocalBootstrapper(options, kubeConfig, kubeClient, latticeClient)
	}

	if options.Config.Provider.AWS != nil {
		return NewCloudBootstrapper(options, kubeConfig, kubeClient, latticeClient)
	}

	return nil, fmt.Errorf("must specify Provider in Config")
}

type DefaultLocalBootstrapper struct {
	BaseBootstrapper  BaseBootstrapper
	LocalBootstrapper LocalBootstrapper
}

func NewLocalBootstrapper(
	options *Options,
	kubeConfig *rest.Config,
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
) (*DefaultLocalBootstrapper, error) {
	if options == nil {
		return nil, fmt.Errorf("options required")
	}

	baseOptions := &base.Options{
		DryRun:           options.DryRun,
		Config:           options.Config,
		MasterComponents: options.MasterComponents,
	}
	baseBootstrapper, err := base.NewBootstrapper(baseOptions, kubeConfig, kubeClient, latticeClient)
	if err != nil {
		return nil, err
	}

	localOptions := &local.Options{
		DryRun: options.DryRun,
	}
	localBootstrapper := local.NewBootstrapper(localOptions, kubeClient)

	b := &DefaultLocalBootstrapper{
		BaseBootstrapper:  baseBootstrapper,
		LocalBootstrapper: localBootstrapper,
	}
	return b, nil
}

func (b *DefaultLocalBootstrapper) Bootstrap() ([]interface{}, error) {
	objects := []interface{}{}
	additionalObjects, err := b.BaseBootstrapper.BaseBootstrap()
	if err != nil {
		return nil, err
	}
	objects = append(objects, additionalObjects...)

	additionalObjects, err = b.LocalBootstrapper.LocalBootstrap()
	if err != nil {
		return nil, err
	}
	objects = append(objects, additionalObjects...)

	return objects, nil
}

type DefaultCloudBootstrapper struct {
	BaseBootstrapper  BaseBootstrapper
	CloudBootstrapper CloudBootstrapper
}

func NewCloudBootstrapper(
	options *Options,
	kubeConfig *rest.Config,
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
) (*DefaultCloudBootstrapper, error) {
	if options == nil {
		return nil, fmt.Errorf("options required")
	}

	baseOptions := &base.Options{
		DryRun:           options.DryRun,
		Config:           options.Config,
		MasterComponents: options.MasterComponents,
	}
	baseBootstrapper, err := base.NewBootstrapper(baseOptions, kubeConfig, kubeClient, latticeClient)
	if err != nil {
		return nil, err
	}

	cloudOptions := &cloud.Options{
		DryRun:     options.DryRun,
		Networking: options.Networking,
	}
	cloudBootstrapper, err := cloud.NewBootstrapper(cloudOptions, kubeClient)
	if err != nil {
		return nil, err
	}

	b := &DefaultCloudBootstrapper{
		BaseBootstrapper:  baseBootstrapper,
		CloudBootstrapper: cloudBootstrapper,
	}
	return b, nil
}

func (b *DefaultCloudBootstrapper) Bootstrap() ([]interface{}, error) {
	objects := []interface{}{}
	additionalObjects, err := b.BaseBootstrapper.BaseBootstrap()
	if err != nil {
		return nil, err
	}
	objects = append(objects, additionalObjects...)

	additionalObjects, err = b.CloudBootstrapper.CloudBootstrap()
	if err != nil {
		return nil, err
	}
	objects = append(objects, additionalObjects...)

	return objects, nil
}
