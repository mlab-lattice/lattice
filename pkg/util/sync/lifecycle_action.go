package sync

import (
	"fmt"
	"strings"
	"sync"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

func NewLifecycleActionManager() *LifecycleActionManager {
	return &LifecycleActionManager{systems: make(map[v1.SystemID]*lifecycleActionTree)}
}

type LifecycleActionManager struct {
	// this could be a RWMutex but trying to keep things "simple" to start
	sync.Mutex

	// important: systemIDs are assumed to be unique across time and space.
	// that is, if a system is deleted and then another system is created,
	// it is impossible that the new system will have the same name
	systems map[v1.SystemID]*lifecycleActionTree
}

func (a *LifecycleActionManager) InProgressActions(system v1.SystemID) ([]v1.DeployID, *v1.TeardownID) {
	a.Lock()
	defer a.Unlock()

	t, ok := a.systems[system]
	if !ok {
		return nil, nil
	}

	var deploys []v1.DeployID
	for id := range t.deploys {
		deploys = append(deploys, id)
	}

	return deploys, t.teardown
}

func (a *LifecycleActionManager) AcquireDeploy(system v1.SystemID, deploy v1.DeployID, path tree.Path) error {
	a.Lock()
	defer a.Unlock()

	t, ok := a.systems[system]
	if !ok {
		t = newLifecycleActionTree()
		a.systems[system] = t
	}

	return t.AcquireDeploy(deploy, path)
}

func (a *LifecycleActionManager) ReleaseDeploy(system v1.SystemID, deploy v1.DeployID) {
	a.Lock()
	defer a.Unlock()

	t, ok := a.systems[system]
	if !ok {
		// this shouldn't happen
		// TODO(kevindrosendahl): send warn event
		return
	}

	t.ReleaseDeploy(deploy)
}

func (a *LifecycleActionManager) AcquireTeardown(system v1.SystemID, teardown v1.TeardownID) error {
	a.Lock()
	defer a.Unlock()

	t, ok := a.systems[system]
	if !ok {
		t = newLifecycleActionTree()
		a.systems[system] = t
	}

	return t.AcquireTeardown(teardown)
}

func (a *LifecycleActionManager) ReleaseTeardown(system v1.SystemID, teardown v1.TeardownID) {
	a.Lock()
	defer a.Unlock()

	t, ok := a.systems[system]
	if !ok {
		// this shouldn't happen
		// TODO(kevindrosendahl): send warn event
		return
	}

	t.ReleaseTeardown(teardown)
}

func newLifecycleActionTree() *lifecycleActionTree {
	return &lifecycleActionTree{
		inner:   tree.NewRadix(),
		deploys: make(map[v1.DeployID]tree.Path),
	}
}

// lifecycleActionTree allows deploys to attempt to acquire intention locks
// down the tree to its path and exclusive locks at its full path, and teardowns
// to attempt to acquire an exclusive lock on the entire system
// important note: there is no synchronization inside the lifecycleActionTree,
// it is assumed that the lifecycleActionTree is accessed via a LifecycleActionManager
// which is coordinating synchronization
type lifecycleActionTree struct {
	inner *tree.Radix

	deploys  map[v1.DeployID]tree.Path
	teardown *v1.TeardownID
}

func newLifecycleActionTreeNode() *lifecycleActionTreeNode {
	return &lifecycleActionTreeNode{
		Deploys: make(map[v1.DeployID]*IntentionLockUnlocker),
	}
}

type lifecycleActionTreeNode struct {
	Lock     IntentionLock
	Deploys  map[v1.DeployID]*IntentionLockUnlocker
	Teardown *lifecycleActionTreeNodeTeardown
}

type lifecycleActionTreeNodeTeardown struct {
	ID       v1.TeardownID
	Unlocker *IntentionLockUnlocker
}

func (t *lifecycleActionTree) AcquireDeploy(id v1.DeployID, path tree.Path) error {
	for i := 0; i <= path.Depth(); i++ {
		// acquire intention locks all the way until the leaf of the path, and
		// acquire an exclusive Lock on the leaf
		granularity := LockGranularityIntentionExclusive
		if i == path.Depth() {
			granularity = LockGranularityExclusive
		}

		prefix, _ := path.Prefix(i)

		// if there doesn't exist a node for this prefix yet, create it and Lock it at the
		// desired granularity
		n, ok := t.get(prefix)
		if !ok {
			n := newLifecycleActionTreeNode()
			unlocker, ok := n.Lock.TryLock(granularity)
			if !ok {
				// want to unlock all the locks we just created so we don't keep the tree locked now
				// that we know we can't lock our node
				t.releaseDeploy(id, path)
				return fmt.Errorf("unable to Lock freshly created lock for deploy %v at path %v", id, prefix.String())
			}

			n.Deploys[id] = unlocker
			t.insert(prefix, n)
			continue
		}

		// see if this deploy has already locked this path.
		// if it has at the same granularity, then there's no problem.
		// if it's trying to change its granularity it must release the lock first, so return an error
		unlocker, ok := n.Deploys[id]
		if ok {
			if unlocker.Granularity() != granularity {
				return fmt.Errorf(
					"attempting to lock for deploy %v at path %v with granularity %v when it is already locked for the deploy with granularity %v",
					id,
					prefix.String(),
					granularity,
					unlocker.Granularity(),
				)
			}

			continue
		}

		unlocker, ok = n.Lock.TryLock(granularity)
		if !ok {
			// want to unlock all the locks we just created so we don't keep the tree locked now
			// that we know we can't lock our node
			t.releaseDeploy(id, path)
			return newConflictingLifecycleActionError(n)
		}

		n.Deploys[id] = unlocker
	}

	t.deploys[id] = path
	return nil
}

func (t *lifecycleActionTree) ReleaseDeploy(id v1.DeployID) {
	path, ok := t.deploys[id]
	if !ok {
		// this shouldn't happen
		// TODO(kevindrosendahl): send warn event
		return
	}

	t.releaseDeploy(id, path)
	delete(t.deploys, id)
}

func (t *lifecycleActionTree) releaseDeploy(id v1.DeployID, path tree.Path) {
	for i := 0; i <= path.Depth(); i++ {
		prefix, _ := path.Prefix(i)

		n, ok := t.get(prefix)
		if !ok {
			// this shouldn't happen
			// TODO(kevindrosendahl): send warn event
			continue
		}

		unlocker, ok := n.Deploys[id]
		if !ok {
			// this shouldn't happen
			// TODO(kevindrosendahl): send warn event
			continue
		}

		unlocker.Unlock()
		delete(n.Deploys, id)
	}
}

func (t *lifecycleActionTree) AcquireTeardown(id v1.TeardownID) error {
	// if there doesn't exist a node for the root yet, create it and lock it exclusively
	n, ok := t.get(tree.RootPath())
	if !ok {
		n := newLifecycleActionTreeNode()
		unlocker, ok := n.Lock.TryLock(LockGranularityExclusive)
		if !ok {
			return fmt.Errorf("unable to Lock freshly created lock for teardown %v", id)
		}

		n.Teardown = &lifecycleActionTreeNodeTeardown{
			ID:       id,
			Unlocker: unlocker,
		}
		t.insert(tree.RootPath(), n)
		t.teardown = &id
		return nil
	}

	unlocker, ok := n.Lock.TryLock(LockGranularityExclusive)
	if !ok {
		return newConflictingLifecycleActionError(n)
	}

	n.Teardown = &lifecycleActionTreeNodeTeardown{
		ID:       id,
		Unlocker: unlocker,
	}
	t.teardown = &id
	return nil
}

func (t *lifecycleActionTree) ReleaseTeardown(id v1.TeardownID) {
	// bail out if it looks like this teardown doesn't actually own the systems
	n, ok := t.get(tree.RootPath())
	if !ok {
		return
	}
	if n.Teardown == nil || n.Teardown.ID != id {
		return
	}

	// when you tear down a system, you've removed the entire system tree
	// we can use that as an opportunity here to prune our lifecycleActionTree
	// TODO(kevindrosendahl): this assumes a teardown succeeded, need to reconsider the failure case
	t.inner.DeletePrefix(tree.RootPath())
	t.teardown = nil
}

func (t *lifecycleActionTree) insert(p tree.Path, n *lifecycleActionTreeNode) (*lifecycleActionTreeNode, bool) {
	i, ok := t.inner.Insert(p, n)
	if !ok {
		return nil, false
	}

	return i.(*lifecycleActionTreeNode), true
}

func (t *lifecycleActionTree) get(p tree.Path) (*lifecycleActionTreeNode, bool) {
	i, ok := t.inner.Get(p)
	if !ok {
		return nil, false
	}

	return i.(*lifecycleActionTreeNode), true
}

func newConflictingLifecycleActionError(n *lifecycleActionTreeNode) *ConflictingLifecycleActionError {
	var deploys []string
	for id := range n.Deploys {
		deploys = append(deploys, string(id))
	}

	var teardownID *v1.TeardownID
	if n.Teardown != nil {
		teardownID = &n.Teardown.ID
	}

	return &ConflictingLifecycleActionError{
		deploys:  deploys,
		teardown: teardownID,
	}
}

type ConflictingLifecycleActionError struct {
	deploys  []string
	teardown *v1.TeardownID
}

func (e *ConflictingLifecycleActionError) Error() string {
	// this shouldn't happen
	if len(e.deploys) > 0 && e.teardown != nil {
		return fmt.Sprintf("locked by deploys %v and teardown %v", strings.Join(e.deploys, ", "), *e.teardown)
	}

	if len(e.deploys) > 0 {
		return fmt.Sprintf("locked by deploys %v", strings.Join(e.deploys, ", "))
	}

	if e.teardown != nil {
		return fmt.Sprintf("locked by teardown %v", *e.teardown)
	}

	// this also shouldn't happen
	return fmt.Sprintf("locked by unknown entity")
}
