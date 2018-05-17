package system

import (
	"fmt"
	"reflect"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/deckarep/golang-set"
	"github.com/satori/go.uuid"
)

func (c *Controller) syncSystemNodePools(
	system *latticev1.System,
) (map[v1.NodePoolPath]latticev1.SystemStatusNodePool, error) {
	// N.B.: as it currently is, this controller does not allow for a "move" i.e.
	// renaming a node pool (changing its path). (see nodePool.go for more information)
	nodePools := make(map[v1.NodePoolPath]latticev1.SystemStatusNodePool)
	systemNamespace := system.ResourceNamespace(c.namespacePrefix)
	nodePoolNames := mapset.NewSet()

	// Loop through the nodePools defined in the system's Spec, and create/update any that need it
	for path, spec := range system.Spec.NodePools {
		var nodePool *latticev1.NodePool

		if path.Name == nil {
			// FIXME: all system level node pools should have a path, send a warn if it doesn't
			continue
		}

		nodePoolStatus, ok := system.Status.NodePools[path]
		if !ok {
			// If a status for this node pool's path hasn't been set, then either we haven't created the node pool yet,
			// or we were unable to update the system's Status after creating the node pool

			// First check our cache to see if the node pool exists.
			var err error
			nodePool, err = c.getNodePoolFromCache(systemNamespace, path.Path, *path.Name)
			if err != nil {
				return nil, err
			}

			if nodePool == nil {
				// The nodePool wasn't in the cache, so do a quorum read to see if it was created.
				// N.B.: could first loop through and check to see if we need to do a quorum read
				// on any of the nodePools, then just do one list.
				nodePool, err = c.getNodePoolFromAPI(systemNamespace, path.Path, *path.Name)
				if err != nil {
					return nil, err
				}

				if nodePool == nil {
					// The nodePool actually doesn't exist yet. Create it with a new UUID as the name.
					nodePool, err = c.createNewNodePool(system, path.Path, *path.Name, spec)
					if err != nil {
						return nil, err
					}

					// Successfully created the nodePool. No need to check if it needs to be updated.
					nodePools[path] = latticev1.SystemStatusNodePool{
						Name:           nodePool.Name,
						Generation:     nodePool.Generation,
						NodePoolStatus: nodePool.Status,
					}
					nodePoolNames.Add(nodePool.Name)
					continue
				}
			}
			// We were able to find an existing nodePool for this path. We'll check below if it
			// needs to be updated.
		} else {
			// There is supposedly already a nodePool for this path.
			nodePoolName := nodePoolStatus.Name
			var err error

			nodePool, err = c.nodePoolLister.NodePools(systemNamespace).Get(nodePoolName)
			if err != nil {
				if !errors.IsNotFound(err) {
					return nil, fmt.Errorf("error trying to get cached node pool %v for %v", nodePoolName, system.Description())
				}

				// The nodePool wasn't in the cache. Perhaps it was recently created. Do a quorum read.
				nodePool, err = c.latticeClient.LatticeV1().NodePools(systemNamespace).Get(nodePoolName, metav1.GetOptions{})
				if err != nil {
					if !errors.IsNotFound(err) {
						return nil, fmt.Errorf("error trying to get node pool %v for %v", nodePoolName, system.Description())
					}

					// FIXME: should we just recreate the nodePool here?
					// what happens when a deploy doesnt fully succeed and there's a leftover terminating nodePool with
					// the same path as a new nodePool?
					return nil, fmt.Errorf("%v has reference to non existant nodePool %v", system.Description(), nodePoolName)
				}
			}
		}

		// We found an existing nodePool, update it if needed
		nodePool, err := c.updateNodePool(nodePool, spec, path.Path, *path.Name)
		if err != nil {
			return nil, err
		}

		nodePoolNames.Add(nodePool.Name)
		nodePools[path] = latticev1.SystemStatusNodePool{
			Name:           nodePool.Name,
			Generation:     nodePool.Generation,
			NodePoolStatus: nodePool.Status,
		}
	}

	// Loop through all of the node pools that exist in the systems's namespace, and delete any
	// that are no longer a part of the system's Spec
	// TODO(kevinrosendahl): should we wait until all other node pools are successfully rolled out before deleting these?
	allNodePools, err := c.nodePoolLister.NodePools(systemNamespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, nodePool := range allNodePools {
		if !nodePoolNames.Contains(nodePool.Name) {
			// if the node pool is not a system shared node pool, it's not our responsibility
			// to clean it up
			if _, ok, err := nodePool.SystemSharedPathLabel(); err == nil || !ok {
				continue
			}

			if nodePool.DeletionTimestamp == nil {
				err := c.deleteNodePool(nodePool)
				if err != nil {
					return nil, err
				}
			}

			path, ok, err := nodePool.SystemSharedPathLabel()
			if err != nil || !ok {
				// FIXME: warn
				continue
			}

			// copy so the shared cache isn't mutated
			status := nodePool.Status.DeepCopy()
			status.State = latticev1.NodePoolStatePending

			nodePools[path] = latticev1.SystemStatusNodePool{
				Name:           nodePool.Name,
				Generation:     nodePool.Generation,
				NodePoolStatus: *status,
			}
		}
	}

	return nodePools, nil
}

func (c *Controller) createNewNodePool(
	system *latticev1.System,
	path tree.NodePath,
	name string,
	spec latticev1.NodePoolSpec,
) (*latticev1.NodePool, error) {
	nodePool := c.newNodePool(system, path, name, spec)

	result, err := c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).Create(nodePool)
	if err != nil {
		return nil, fmt.Errorf("error creating new node pool for %v in %v: %v", path.String(), system.Description(), err)
	}

	return result, nil
}

func (c *Controller) newNodePool(
	system *latticev1.System,
	path tree.NodePath,
	name string,
	spec latticev1.NodePoolSpec,
) *latticev1.NodePool {
	systemNamespace := system.ResourceNamespace(c.namespacePrefix)

	return &latticev1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:            uuid.NewV4().String(),
			Namespace:       systemNamespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(system, latticev1.SystemKind)},
			Labels: map[string]string{
				latticev1.NodePoolSystemSharedPathLabelKey: latticev1.NodePoolSystemSharedPathLabelValue(path, name),
			},
		},
		Spec: spec,
	}
}

func (c *Controller) deleteNodePool(nodePool *latticev1.NodePool) error {
	// background delete will add deletionTimestamp to the service, but will not
	// try to act upon any of the dependents since the service has a finalizer
	// this allows us to clean up the service in a controlled way
	backgroundDelete := metav1.DeletePropagationBackground
	deleteOptions := &metav1.DeleteOptions{
		PropagationPolicy: &backgroundDelete,
	}

	err := c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).Delete(nodePool.Name, deleteOptions)
	if err != nil {
		return fmt.Errorf("error deleting %v: %v", nodePool.Description(c.namespacePrefix), err)
	}

	return nil
}

func (c *Controller) updateNodePool(
	nodePool *latticev1.NodePool,
	spec latticev1.NodePoolSpec,
	path tree.NodePath,
	name string,
) (*latticev1.NodePool, error) {
	if !c.nodePoolNeedsUpdate(nodePool, spec, path, name) {
		return nodePool, nil
	}

	// Copy so the cache isn't mutated
	nodePool = nodePool.DeepCopy()
	nodePool.Spec = spec

	if nodePool.Labels == nil {
		nodePool.Labels = make(map[string]string)
	}
	nodePool.Labels[latticev1.NodePoolSystemSharedPathLabelKey] = latticev1.NodePoolSystemSharedPathLabelValue(path, name)

	result, err := c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).Update(nodePool)
	if err != nil {
		return nil, fmt.Errorf("error updating %v: %v", nodePool.Description(c.namespacePrefix), err)
	}

	return result, err
}

func (c *Controller) nodePoolNeedsUpdate(
	nodePool *latticev1.NodePool,
	spec latticev1.NodePoolSpec,
	path tree.NodePath,
	name string,
) bool {
	if !reflect.DeepEqual(nodePool.Spec, spec) {
		return true
	}

	currentPath, ok, err := nodePool.SystemSharedPathLabel()
	if err != nil || !ok {
		return true
	}

	np := v1.NewSystemSharedNodePoolPath(path, name)
	return currentPath != np
}

func (c *Controller) getNodePoolFromCache(namespace string, path tree.NodePath, name string) (*latticev1.NodePool, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(
		latticev1.NodePoolSystemSharedPathLabelKey,
		selection.Equals,
		[]string{latticev1.NodePoolSystemSharedPathLabelValue(path, name)},
	)
	if err != nil {
		return nil, fmt.Errorf("error getting selector for cached node pool %v in namespace %v", path.String(), namespace)
	}
	selector = selector.Add(*requirement)

	nodePools, err := c.nodePoolLister.NodePools(namespace).List(selector)
	if err != nil {
		return nil, fmt.Errorf("error getting cached node pools in namespace %v", namespace)
	}

	if len(nodePools) == 0 {
		return nil, nil
	}

	if len(nodePools) > 1 {
		return nil, fmt.Errorf("found multiple cached node pools with path %v in namespace %v", path.String(), namespace)
	}

	return nodePools[0], nil
}

func (c *Controller) getNodePoolFromAPI(namespace string, path tree.NodePath, name string) (*latticev1.NodePool, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(
		latticev1.ServicePathLabelKey,
		selection.Equals,
		[]string{latticev1.NodePoolSystemSharedPathLabelValue(path, name)},
	)
	if err != nil {
		return nil, fmt.Errorf("error getting selector for node pool %v in namespace %v", path.String(), namespace)
	}
	selector = selector.Add(*requirement)

	nodePools, err := c.latticeClient.LatticeV1().NodePools(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, fmt.Errorf("error getting node pools in namespace %v", namespace)
	}

	if len(nodePools.Items) == 0 {
		return nil, nil
	}

	if len(nodePools.Items) > 1 {
		return nil, fmt.Errorf("found multiple node pools with path %v in namespace %v", path.String(), namespace)
	}

	return &nodePools.Items[0], nil
}
