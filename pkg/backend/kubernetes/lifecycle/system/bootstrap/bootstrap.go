package bootstrap

import (
	"fmt"

	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	basebootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper/base"
	"github.com/mlab-lattice/system/pkg/types"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	kubeclientset "k8s.io/client-go/kubernetes"
)

func Bootstrap(
	latticeID types.LatticeID,
	systemID types.SystemID,
	definitionURL string,
	bootstrappers []bootstrapper.Interface,
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
) (*bootstrapper.SystemResources, error) {
	resources := GetBootstrapResources(latticeID, systemID, definitionURL, bootstrappers)

	fmt.Println("seeding namespaces")
	namespace, err := kubeClient.CoreV1().Namespaces().Create(resources.Namespace)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return nil, err
		}

		namespace, err = kubeClient.CoreV1().Namespaces().Get(namespace.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	}

	fmt.Println("seeding service account")
	var serviceAccounts []*corev1.ServiceAccount
	for _, sa := range resources.ServiceAccounts {
		sa, err = kubeClient.CoreV1().ServiceAccounts(namespace.Name).Create(sa)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				return nil, err
			}

			sa, err = kubeClient.CoreV1().ServiceAccounts(namespace.Name).Get(sa.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
		}

		serviceAccounts = append(serviceAccounts, sa)
	}

	fmt.Println("seeding role bindings")
	var roleBindings []*rbacv1.RoleBinding
	for _, roleBinding := range resources.RoleBindings {
		roleBinding, err = kubeClient.RbacV1().RoleBindings(namespace.Name).Create(roleBinding)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				return nil, err
			}

			roleBinding, err = kubeClient.RbacV1().RoleBindings(namespace.Name).Get(roleBinding.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
		}

		roleBindings = append(roleBindings, roleBinding)
	}

	fmt.Println("seeding daemon sets")
	var daemonSets []*appsv1.DaemonSet
	for _, daemonSet := range resources.DaemonSets {
		daemonSet, err = kubeClient.AppsV1().DaemonSets(namespace.Name).Create(daemonSet)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				return nil, err
			}

			daemonSet, err = kubeClient.AppsV1().DaemonSets(namespace.Name).Get(daemonSet.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
		}

		daemonSets = append(daemonSets, daemonSet)
	}

	fmt.Println("seeding system")
	system, err := latticeClient.LatticeV1().Systems(namespace.Name).Create(resources.System)
	if err != nil {
		return nil, err
	}

	resources = &bootstrapper.SystemResources{
		System:          system,
		Namespace:       namespace,
		ServiceAccounts: serviceAccounts,
		RoleBindings:    roleBindings,
		DaemonSets:      daemonSets,
	}
	return resources, err
}

func GetBootstrapResources(
	latticeID types.LatticeID,
	systemID types.SystemID,
	definitionURL string,
	bootstrappers []bootstrapper.Interface,
) *bootstrapper.SystemResources {
	baseOptions := &basebootstrapper.Options{
		LatticeID:     latticeID,
		SystemID:      systemID,
		DefinitionURL: definitionURL,
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
