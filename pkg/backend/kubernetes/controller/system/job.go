package system

import (
	"fmt"
	"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/deckarep/golang-set"
	"github.com/satori/go.uuid"
)

func (c *Controller) syncSystemJobs(system *latticev1.System) error {
	systemNamespace := system.ResourceNamespace(c.namespacePrefix)
	jobNames := mapset.NewSet()

	// Loop through the jobs defined in the system's Spec, and create/update any that need it
	if system.Spec.Definition != nil {
		var err error
		system.Spec.Definition.V1().Jobs(func(path tree.Path, definition *definitionv1.Job, info *resolver.ResolutionInfo) tree.WalkContinuation {
			artifacts, ok := system.Spec.WorkloadBuildArtifacts.Get(path)
			if !ok {
				err = fmt.Errorf(
					"%v spec has job %v but does not have build information about it",
					system.Description(),
					path.String(),
				)
				return tree.HaltWalk
			}

			var job *latticev1.Job

			// First check our cache to see if the job exists.
			job, err = c.getJobFromCache(systemNamespace, path)
			if err != nil {
				return tree.HaltWalk
			}

			if job == nil {
				// The job wasn't in the cache, so do a quorum read to see if it was created.
				// N.B.: could first loop through and check to see if we need to do a quorum read
				// on any of the services, then just do one list.
				job, err = c.getJobFromAPI(systemNamespace, path)
				if err != nil {
					return tree.HaltWalk
				}

				if job == nil {
					// The job actually doesn't exist yet. Create it with a new UUID as the name.
					job, err = c.createNewJob(system, path, definition, &artifacts)
					if err != nil {
						return tree.HaltWalk
					}

					// Successfully created the job. No need to check if it needs to be updated.
					jobNames.Add(job.Name)
					return tree.ContinueWalk
				}
			}

			// We found an existing job. Calculate what its Spec should look like,
			// and update the job if its current Spec is different.
			spec := jobSpec(definition, &artifacts)
			job, err = c.updateJob(job, spec, path)
			if err != nil {
				return tree.HaltWalk
			}

			jobNames.Add(job.Name)
			return tree.ContinueWalk
		})
		if err != nil {
			return err
		}
	}

	// Loop through all of the jobs that exist in the System's namespace, and delete any
	// that are no longer a part of the system's Spec
	allJobs, err := c.jobLister.Jobs(systemNamespace).List(labels.Everything())
	if err != nil {
		return err
	}

	for _, job := range allJobs {
		if !jobNames.Contains(job.Name) {
			if job.DeletionTimestamp == nil {
				err := c.deleteJob(job)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (c *Controller) createNewJob(
	system *latticev1.System,
	path tree.Path,
	definition *definitionv1.Job,
	artifacts *latticev1.WorkloadContainerBuildArtifacts,
) (*latticev1.Job, error) {
	job, err := c.newJob(system, path, definition, artifacts)
	if err != nil {
		return nil, fmt.Errorf("error getting new job for %v in %v: %v", path.String(), system.Description(), err)
	}

	result, err := c.latticeClient.LatticeV1().Jobs(job.Namespace).Create(job)
	if err != nil {
		return nil, fmt.Errorf("error creating new job for %v in %v: %v", path.String(), system.Description(), err)
	}

	return result, nil
}

func (c *Controller) newJob(
	system *latticev1.System,
	path tree.Path,
	definition *definitionv1.Job,
	artifacts *latticev1.WorkloadContainerBuildArtifacts,
) (*latticev1.Job, error) {
	systemNamespace := system.ResourceNamespace(c.namespacePrefix)
	job := &latticev1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            uuid.NewV4().String(),
			Namespace:       systemNamespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(system, latticev1.SystemKind)},
			Labels: map[string]string{
				latticev1.JobPathLabelKey: path.ToDomain(),
			},
		},
		Spec: jobSpec(definition, artifacts),
	}

	return job, nil
}

func jobSpec(
	definition *definitionv1.Job,
	artifacts *latticev1.WorkloadContainerBuildArtifacts,
) latticev1.JobSpec {
	return latticev1.JobSpec{
		Definition:              *definition,
		ContainerBuildArtifacts: *artifacts,
	}
}

func (c *Controller) deleteJob(job *latticev1.Job) error {
	// As of right now the job controller does not act upon Job objects, just JobRun
	// objects, so there is no cleaning up that it needs to do.
	// N.B.: When adding scheduled jobs this may need to change
	backgroundDelete := metav1.DeletePropagationForeground
	deleteOptions := &metav1.DeleteOptions{
		PropagationPolicy: &backgroundDelete,
	}

	err := c.latticeClient.LatticeV1().Jobs(job.Namespace).Delete(job.Name, deleteOptions)
	if err != nil {
		return fmt.Errorf("error deleting %v: %v", job.Description(c.namespacePrefix), err)
	}

	return nil
}

func (c *Controller) updateJob(job *latticev1.Job, spec latticev1.JobSpec, path tree.Path) (*latticev1.Job, error) {
	if !c.jobNeedsUpdate(job, spec, path) {
		return job, nil
	}

	// Copy so the cache isn't mutated
	job = job.DeepCopy()
	job.Spec = spec

	if job.Labels == nil {
		job.Labels = make(map[string]string)
	}
	job.Labels[latticev1.JobPathLabelKey] = path.ToDomain()

	result, err := c.latticeClient.LatticeV1().Jobs(job.Namespace).Update(job)
	if err != nil {
		return nil, fmt.Errorf("error updating %v: %v", job.Description(c.namespacePrefix), err)
	}

	return result, err
}

func (c *Controller) jobNeedsUpdate(job *latticev1.Job, spec latticev1.JobSpec, path tree.Path) bool {
	if !reflect.DeepEqual(job.Spec, spec) {
		return true
	}

	currentPath, err := job.PathLabel()
	if err != nil {
		return true
	}

	return currentPath != path
}

func (c *Controller) getJobFromCache(namespace string, path tree.Path) (*latticev1.Job, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.JobPathLabelKey, selection.Equals, []string{path.ToDomain()})
	if err != nil {
		return nil, fmt.Errorf("error getting selector for cached job %v in namespace %v", path.String(), namespace)
	}
	selector = selector.Add(*requirement)

	jobs, err := c.jobLister.Jobs(namespace).List(selector)
	if err != nil {
		return nil, fmt.Errorf("error getting cached jobs in namespace %v", namespace)
	}

	if len(jobs) == 0 {
		return nil, nil
	}

	if len(jobs) > 1 {
		return nil, fmt.Errorf("found multiple cached jobs with path %v in namespace %v", path.String(), namespace)
	}

	return jobs[0], nil
}

func (c *Controller) getJobFromAPI(namespace string, path tree.Path) (*latticev1.Job, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.JobPathLabelKey, selection.Equals, []string{path.ToDomain()})
	if err != nil {
		return nil, fmt.Errorf("error getting selector for job %v in namespace %v", path.String(), namespace)
	}
	selector = selector.Add(*requirement)

	jobs, err := c.latticeClient.LatticeV1().Jobs(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, fmt.Errorf("error getting jobs in namespace %v", namespace)
	}

	if len(jobs.Items) == 0 {
		return nil, nil
	}

	if len(jobs.Items) > 1 {
		return nil, fmt.Errorf("found multiple jobs with path %v in namespace %v", path.String(), namespace)
	}

	return &jobs.Items[0], nil
}
