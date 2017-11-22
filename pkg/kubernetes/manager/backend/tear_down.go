package backend

import (
	"strings"

	coretypes "github.com/mlab-lattice/core/pkg/types"

	"github.com/mlab-lattice/system/pkg/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func (kb *KubernetesBackend) TearDownSystem(ln coretypes.LatticeNamespace) (coretypes.SystemTeardownId, error) {
	systemTeardown, err := getSystemTeardown(ln)
	if err != nil {
		return "", err
	}

	result := &crv1.SystemTeardown{}
	err = kb.LatticeResourceClient.Post().
		Namespace(constants.NamespaceLatticeInternal).
		Resource(crv1.SystemTeardownResourcePlural).
		Body(systemTeardown).
		Do().
		Into(result)

	return coretypes.SystemTeardownId(result.Name), err
}

func getSystemTeardown(ln coretypes.LatticeNamespace) (*crv1.SystemTeardown, error) {
	labels := map[string]string{
		constants.LatticeNamespaceLabel: string(ln),
	}

	sysT := &crv1.SystemTeardown{
		ObjectMeta: metav1.ObjectMeta{
			Name:   string(uuid.NewUUID()),
			Labels: labels,
		},
		Spec: crv1.SystemTeardownSpec{
			LatticeNamespace: ln,
		},
		Status: crv1.SystemTeardownStatus{
			State: crv1.SystemTeardownStatePending,
		},
	}

	return sysT, nil
}

func (kb *KubernetesBackend) GetSystemTeardown(ln coretypes.LatticeNamespace, tid coretypes.SystemTeardownId) (*coretypes.SystemTeardown, bool, error) {
	result := &crv1.SystemTeardown{}
	err := kb.LatticeResourceClient.Get().
		Namespace(constants.NamespaceLatticeInternal).
		Resource(crv1.SystemTeardownResourcePlural).
		Name(string(tid)).
		Do().
		Into(result)

	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	// TODO: add this to the query
	if strings.Compare(result.Labels[constants.LatticeNamespaceLabel], string(ln)) != 0 {
		return nil, false, nil
	}

	sb := &coretypes.SystemTeardown{
		Id:    tid,
		State: getSystemTeardownState(result.Status.State),
	}

	return sb, true, nil
}

func (kb *KubernetesBackend) ListSystemTeardowns(ln coretypes.LatticeNamespace) ([]coretypes.SystemTeardown, error) {
	result := &crv1.SystemTeardownList{}
	err := kb.LatticeResourceClient.Get().
		Namespace(constants.NamespaceLatticeInternal).
		Resource(crv1.SystemTeardownResourcePlural).
		Do().
		Into(result)

	if err != nil {
		return nil, err
	}

	teardowns := []coretypes.SystemTeardown{}
	for _, b := range result.Items {
		// TODO: add this to the query
		if strings.Compare(b.Labels[constants.LatticeNamespaceLabel], string(ln)) != 0 {
			continue
		}

		teardowns = append(teardowns, coretypes.SystemTeardown{
			Id:    coretypes.SystemTeardownId(b.Name),
			State: getSystemTeardownState(b.Status.State),
		})
	}

	return teardowns, nil
}

func getSystemTeardownState(state crv1.SystemTeardownState) coretypes.SystemTeardownState {
	switch state {
	case crv1.SystemTeardownStatePending:
		return coretypes.SystemTeardownStatePending
	case crv1.SystemTeardownStateSucceeded:
		return coretypes.SystemTeardownStateSucceeded
	case crv1.SystemTeardownStateFailed:
		return coretypes.SystemTeardownStateFailed
	default:
		panic("unreachable")
	}
}
