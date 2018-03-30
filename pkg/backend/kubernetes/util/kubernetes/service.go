package kubernetes

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticev1client "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned/typed/lattice/v1"
	"github.com/mlab-lattice/system/pkg/definition/tree"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func GetServiceForPath(client latticev1client.ServicesGetter, namespace string, path tree.NodePath) (*latticev1.Service, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(constants.LabelKeyServicePath, selection.Equals, []string{path.String()})
	if err != nil {
		return nil, err
	}
	selector = selector.Add(*requirement)

	services, err := client.Services(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	if len(services.Items) == 0 {
		return nil, errors.NewNotFound(latticev1.Resource(latticev1.ResourceSingularService), path.ToDomain())
	}

	if len(services.Items) > 1 {
		return nil, fmt.Errorf("found multiple services with path %v", path.String())
	}

	return &services.Items[0], nil
}
