package systemlifecycle

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FIXME: add proper errors etc to all this
func (a *lifecycleAction) String() string {
	if a.deploy != nil {
		return fmt.Sprintf("Deploy %v", a.deploy.Name)
	}

	if a.teardown != nil {
		return fmt.Sprintf("SystemTeardown %v", a.teardown.Name)
	}

	// TODO: what should we do here?
	return "unknown"
}

func (a *lifecycleAction) Equal(other *lifecycleAction) bool {
	if a.deploy != nil {
		if other.deploy == nil {
			return false
		}

		return a.deploy.Namespace == other.deploy.Namespace && a.deploy.Name == other.deploy.Name
	}

	if a.teardown != nil {
		if other.teardown == nil {
			return false
		}

		return a.teardown.Namespace == other.teardown.Namespace && a.teardown.Name == other.teardown.Name
	}

	// TODO: what should we do here?
	return false
}

func (c *Controller) getOwningAction(namespace string) (*lifecycleAction, bool) {
	c.owningLifecycleActionsLock.RLock()
	defer c.owningLifecycleActionsLock.RUnlock()

	ns, err := c.kubeClient.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
	// FIXME: think about this, probably wrong
	if err != nil {
		return nil, false
	}

	owningAction, ok := c.owningLifecycleActions[ns.UID]
	return owningAction, ok
}

func (c *Controller) attemptToClaimDeployOwningAction(rollout *latticev1.Deploy) (*lifecycleAction, error) {
	action := &lifecycleAction{
		deploy: rollout,
	}

	namespace, err := c.kubeClient.CoreV1().Namespaces().Get(rollout.Namespace, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return c.attemptToClaimOwningAction(namespace, action), nil
}

func (c *Controller) attemptToClaimTeardownOwningAction(teardown *latticev1.Teardown) (*lifecycleAction, error) {
	action := &lifecycleAction{
		teardown: teardown,
	}

	namespace, err := c.kubeClient.CoreV1().Namespaces().Get(teardown.Namespace, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return c.attemptToClaimOwningAction(namespace, action), nil
}

func (c *Controller) attemptToClaimOwningAction(namespace *corev1.Namespace, action *lifecycleAction) *lifecycleAction {
	c.owningLifecycleActionsLock.Lock()
	defer c.owningLifecycleActionsLock.Unlock()

	// TODO: should we check to see if the owning action is the same action?
	owningAction, ok := c.owningLifecycleActions[namespace.UID]
	if ok {
		return owningAction
	}

	c.owningLifecycleActions[namespace.UID] = action
	return nil
}

func (c *Controller) relinquishDeployOwningActionClaim(rollout *latticev1.Deploy) error {
	action := &lifecycleAction{
		deploy: rollout,
	}

	namespace, err := c.kubeClient.CoreV1().Namespaces().Get(rollout.Namespace, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return c.relinquishOwningActionClaim(namespace, action)
}

func (c *Controller) relinquishTeardownOwningActionClaim(teardown *latticev1.Teardown) error {
	action := &lifecycleAction{
		teardown: teardown,
	}

	namespace, err := c.kubeClient.CoreV1().Namespaces().Get(teardown.Namespace, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return c.relinquishOwningActionClaim(namespace, action)
}

func (c *Controller) relinquishOwningActionClaim(namespace *corev1.Namespace, action *lifecycleAction) error {
	c.owningLifecycleActionsLock.Lock()
	defer c.owningLifecycleActionsLock.Unlock()

	owningAction, ok := c.owningLifecycleActions[namespace.UID]
	if !ok {
		return fmt.Errorf("expected %v to be owning action for %v namespace but there is no owning action", namespace, action.String())
	}

	if owningAction == nil {
		return fmt.Errorf("expected %v to be owning action %v namespace but owning action is nil", namespace, action.String())
	}

	if !action.Equal(owningAction) {
		return fmt.Errorf("expected %v to be owning action %v namespace but %v is the owning action", namespace, action.String(), owningAction.String())
	}

	delete(c.owningLifecycleActions, namespace.UID)
	return nil
}
