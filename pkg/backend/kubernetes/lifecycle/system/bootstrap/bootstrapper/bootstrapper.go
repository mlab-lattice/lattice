package bootstrapper

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/ghodss/yaml"
)

type Interface interface {
	BootstrapSystemResources(*SystemResources)
}

type SystemResources struct {
	Namespace       *corev1.Namespace
	ServiceAccounts []*corev1.ServiceAccount
	RoleBindings    []*rbacv1.RoleBinding
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

	return output, nil
}
