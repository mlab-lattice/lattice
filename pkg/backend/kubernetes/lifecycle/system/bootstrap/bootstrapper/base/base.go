package base

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Options struct {
	LatticeID       v1.LatticeID
	NamespacePrefix string
	SystemID        v1.SystemID
	DefinitionURL   string
}

func NewBootstrapper(options *Options) *DefaultBootstrapper {
	return &DefaultBootstrapper{
		latticeID:       options.LatticeID,
		namespacePrefix: options.NamespacePrefix,
		systemID:        options.SystemID,
		definitionURL:   options.DefinitionURL,
	}
}

type DefaultBootstrapper struct {
	latticeID       v1.LatticeID
	namespacePrefix string
	systemID        v1.SystemID
	definitionURL   string
}

func (b *DefaultBootstrapper) BootstrapSystemResources(resources *bootstrapper.SystemResources) {
	namespace := &corev1.Namespace{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeutil.SystemNamespace(b.namespacePrefix, b.systemID),
			Labels: map[string]string{
				kubeconstants.LabelKeyLatticeID: string(b.latticeID),
			},
		},
	}
	resources.Namespace = namespace

	componentBuilderSA := &corev1.ServiceAccount{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: rbacv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.ServiceAccountComponentBuilder,
			Namespace: namespace.Name,
		},
	}
	resources.ServiceAccounts = append(resources.ServiceAccounts, componentBuilderSA)

	componentBuilderCRName := kubeutil.ContainerBuilderClusterRoleName(b.namespacePrefix)
	componentBuilderRB := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: rbacv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.ControlPlaneServiceContainerBuilder,
			Namespace: componentBuilderSA.Namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      componentBuilderSA.Name,
				Namespace: componentBuilderSA.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     componentBuilderCRName,
		},
	}
	resources.RoleBindings = append(resources.RoleBindings, componentBuilderRB)
}
