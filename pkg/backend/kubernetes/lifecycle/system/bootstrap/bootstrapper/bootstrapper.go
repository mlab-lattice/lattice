package bootstrapper

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/ghodss/yaml"
)

type Interface interface {
	BootstrapSystemResources(*SystemResources)
}

type SystemResources struct {
	System          *crv1.System
	Namespace       *corev1.Namespace
	ServiceAccounts []*corev1.ServiceAccount
	RoleBindings    []*rbacv1.RoleBinding
	DaemonSets      []*appsv1.DaemonSet
}

func (r *SystemResources) String() (string, error) {
	header := "---\n"
	output := ""

	data, err := yaml.Marshal(r.Namespace)
	if err != nil {
		return "", err
	}

	output += fmt.Sprintf("%v%v", header, string(data))

	for _, serviceAccount := range r.ServiceAccounts {
		data, err := yaml.Marshal(serviceAccount)
		if err != nil {
			return "", err
		}

		output += fmt.Sprintf("%v%v", header, string(data))
	}

	for _, roleBinding := range r.RoleBindings {
		data, err := yaml.Marshal(roleBinding)
		if err != nil {
			return "", err
		}

		output += fmt.Sprintf("%v%v", header, string(data))
	}

	for _, daemonSet := range r.DaemonSets {
		data, err := yaml.Marshal(daemonSet)
		if err != nil {
			return "", err
		}

		output += fmt.Sprintf("%v%v", header, string(data))
	}

	data, err = yaml.Marshal(r.System)
	if err != nil {
		return "", err
	}

	output += fmt.Sprintf("%v%v", header, string(data))

	return output, nil
}
