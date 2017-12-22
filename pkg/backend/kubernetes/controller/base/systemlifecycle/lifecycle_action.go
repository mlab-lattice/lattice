package systemlifecycle

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (a *lifecycleAction) String() string {
	if a.rollout != nil {
		return fmt.Sprintf("SystemRollout %v", a.rollout.Name)
	}

	if a.teardown != nil {
		return fmt.Sprintf("SystemTeardown %v", a.teardown.Name)
	}

	// TODO: what should we do here?
	return "unknown"
}

func (a *lifecycleAction) Equal(other *lifecycleAction) bool {
	if a.rollout != nil {
		if other.rollout == nil {
			return false
		}

		return a.rollout.Namespace == other.rollout.Namespace && a.rollout.Name == other.rollout.Name
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

	owningAction, ok := c.owningLifecycleActions[namespace]
	return owningAction, ok
}

func (c *Controller) attemptToClaimRolloutOwningAction(rollout *crv1.SystemRollout) *lifecycleAction {
	action := &lifecycleAction{
		rollout: rollout,
	}

	return c.attemptToClaimOwningAction(rollout.Namespace, action)
}

func (c *Controller) attemptToClaimTeardownOwningAction(teardown *crv1.SystemTeardown) *lifecycleAction {
	action := &lifecycleAction{
		teardown: teardown,
	}

	return c.attemptToClaimOwningAction(teardown.Namespace, action)
}

func (c *Controller) attemptToClaimOwningAction(namespace string, action *lifecycleAction) *lifecycleAction {
	c.owningLifecycleActionsLock.Lock()
	defer c.owningLifecycleActionsLock.Unlock()

	// TODO: should we check to see if the owning action is the same action?
	owningAction, ok := c.owningLifecycleActions[namespace]
	if ok {
		return owningAction
	}

	c.owningLifecycleActions[namespace] = action
	return nil
}

func (c *Controller) relinquishRolloutOwningActionClaim(rollout *crv1.SystemRollout) error {
	action := &lifecycleAction{
		rollout: rollout,
	}

	return c.relinquishOwningActionClaim(rollout.Namespace, action)
}

func (c *Controller) relinquishTeardownOwningActionClaim(teardown *crv1.SystemTeardown) error {
	action := &lifecycleAction{
		teardown: teardown,
	}

	return c.relinquishOwningActionClaim(teardown.Namespace, action)
}

func (c *Controller) relinquishOwningActionClaim(namespace string, action *lifecycleAction) error {
	c.owningLifecycleActionsLock.Lock()
	defer c.owningLifecycleActionsLock.Unlock()

	owningAction, ok := c.owningLifecycleActions[namespace]
	if !ok {
		return fmt.Errorf("expected %v to be owning action for %v namespace but there is no owning action", namespace, action.String())
	}

	if owningAction == nil {
		return fmt.Errorf("expected %v to be owning action %v namespace but owning action is nil", namespace, action.String())
	}

	if !action.Equal(owningAction) {
		return fmt.Errorf("expected %v to be owning action %v namespace but %v is the owning action", namespace, action.String(), owningAction.String())
	}

	delete(c.owningLifecycleActions, namespace)
	return nil
}
