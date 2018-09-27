package bootstrap

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	basebootstrapper "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper/base"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	kubeclientset "k8s.io/client-go/kubernetes"
)

func Bootstrap(
	namespacePrefix string,
	latticeID v1.LatticeID,
	systemID v1.SystemID,
	definitionURL string,
	bootstrappers []bootstrapper.Interface,
	kubeClient kubeclientset.Interface,
) (*bootstrapper.SystemResources, error) {
	resources := GetBootstrapResources(namespacePrefix, latticeID, systemID, definitionURL, bootstrappers)
	namespace, err := kubeClient.CoreV1().Namespaces().Create(resources.Namespace)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return nil, err
		}

		namespace, err = kubeClient.CoreV1().Namespaces().Get(resources.Namespace.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	}

	var serviceAccounts []*corev1.ServiceAccount
	for _, sa := range resources.ServiceAccounts {
		result, err := kubeClient.CoreV1().ServiceAccounts(namespace.Name).Create(sa)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				return nil, err
			}

			result, err = kubeClient.CoreV1().ServiceAccounts(namespace.Name).Get(sa.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
		}

		serviceAccounts = append(serviceAccounts, result)
	}

	var roleBindings []*rbacv1.RoleBinding
	for _, roleBinding := range resources.RoleBindings {
		result, err := kubeClient.RbacV1().RoleBindings(namespace.Name).Create(roleBinding)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				return nil, err
			}

			result, err = kubeClient.RbacV1().RoleBindings(namespace.Name).Get(roleBinding.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
		}

		roleBindings = append(roleBindings, result)
	}

	resources = &bootstrapper.SystemResources{
		Namespace:       namespace,
		ServiceAccounts: serviceAccounts,
		RoleBindings:    roleBindings,
	}
	return resources, err
}

func GetBootstrapResources(
	namespacePrefix string,
	latticeID v1.LatticeID,
	systemID v1.SystemID,
	definitionURL string,
	bootstrappers []bootstrapper.Interface,
) *bootstrapper.SystemResources {
	baseOptions := &basebootstrapper.Options{
		NamespacePrefix: namespacePrefix,
		LatticeID:       latticeID,
		SystemID:        systemID,
		DefinitionURL:   definitionURL,
	}

	baseBootstrapper := basebootstrapper.NewBootstrapper(baseOptions)

	resources := &bootstrapper.SystemResources{}
	baseBootstrapper.BootstrapSystemResources(resources)

	for _, b := range bootstrappers {
		if b == nil {
			continue
		}

		b.BootstrapSystemResources(resources)
	}

	return resources
}
