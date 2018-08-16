package job

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func (c *Controller) syncKubeJob(
	jobRun *latticev1.JobRun,
	nodePool *latticev1.NodePool,
) (*batchv1.Job, error) {
	kubeJob, err := c.kubeJob(jobRun)
	if err != nil {
		return nil, err
	}

	if kubeJob != nil {
		return kubeJob, nil
	}

	return c.createNewKubeJob(jobRun, nodePool)
}

func (c *Controller) kubeJob(jobRun *latticev1.JobRun) (*batchv1.Job, error) {
	// First check the cache for the deployment
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.JobRunIDLabelKey, selection.Equals, []string{jobRun.Name})
	if err != nil {
		return nil, err
	}
	selector = selector.Add(*requirement)

	cachedKubeJobs, err := c.kubeJobLister.Jobs(jobRun.Namespace).List(selector)
	if err != nil {
		return nil, err
	}

	if len(cachedKubeJobs) > 1 {
		// This may become valid when doing blue/green deploys
		return nil, fmt.Errorf("found multiple cached kube jobs for %v", jobRun.Description(c.namespacePrefix))
	}

	if len(cachedKubeJobs) == 1 {
		return cachedKubeJobs[0], nil
	}

	// Didn't find the deployment in the cache. This likely means it hasn't been created, but since
	// we can't orphan kubeJobs, we need to do a quorum read first to ensure that the deployment
	// doesn't exist
	kubeJobs, err := c.kubeClient.BatchV1().Jobs(jobRun.Namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	if len(kubeJobs.Items) > 1 {
		// This may become valid when doing blue/green deploys
		return nil, fmt.Errorf("found multiple kube jobs for %v", jobRun.Description(c.namespacePrefix))
	}

	if len(kubeJobs.Items) == 1 {
		return &kubeJobs.Items[0], nil
	}

	return nil, nil
}

func (c *Controller) createNewKubeJob(
	jobRun *latticev1.JobRun,
	nodePool *latticev1.NodePool,
) (*batchv1.Job, error) {
	kubeJob, err := c.newKubeJob(jobRun, nodePool)
	if err != nil {
		return nil, err
	}

	result, err := c.kubeClient.BatchV1().Jobs(jobRun.Namespace).Create(kubeJob)
	if err != nil {
		err := fmt.Errorf("error creating kube job for %v: %v", jobRun.Description(c.namespacePrefix), err)
		return nil, err
	}

	return result, nil
}

func (c *Controller) newKubeJob(jobRun *latticev1.JobRun, nodePool *latticev1.NodePool) (*batchv1.Job, error) {
	// Need a consistent view of our config while generating the kube job spec
	c.configLock.RLock()
	defer c.configLock.RUnlock()

	name := kubeJobName(jobRun)
	kubeJobLabels := map[string]string{
		latticev1.JobRunIDLabelKey: jobRun.Name,
	}

	spec, err := c.kubeJobSpec(jobRun, name, kubeJobLabels, nodePool)
	if err != nil {
		err := fmt.Errorf(
			"error generating desired deployment spec for %v (on %v): %v",
			jobRun.Description(c.namespacePrefix),
			nodePool.Description(c.namespacePrefix),
			err,
		)
		return nil, err
	}

	deployment := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Labels:          kubeJobLabels,
			OwnerReferences: []metav1.OwnerReference{*controllerRef(jobRun)},
		},
		Spec: spec,
	}
	return deployment, nil
}

func kubeJobName(jobRun *latticev1.JobRun) string {
	// TODO(kevinrosendahl): May change this to UUID when a Service can have multiple Deployments (e.g. Blue/Green & Canary)
	return fmt.Sprintf("lattice-job-run-%s", jobRun.Name)
}

func (c *Controller) kubeJobSpec(
	jobRun *latticev1.JobRun,
	name string,
	deploymentLabels map[string]string,
	nodePool *latticev1.NodePool,
) (batchv1.JobSpec, error) {
	podTemplateSpec, err := c.untransformedPodTemplateSpec(jobRun, name, deploymentLabels, nodePool)
	if err != nil {
		return batchv1.JobSpec{}, err
	}

	one := int32(1)
	numRetries := int32(1)
	if jobRun.Spec.NumRetries != nil {
		numRetries = *jobRun.Spec.NumRetries
	}

	//return kubeJobSpec, nil
	// IMPORTANT: the order of these TransformWorkloadPodTemplateSpec and the order of the IsDeploymentSpecUpdated calls in
	// isDeploymentSpecUpdated _must_ be inverses.
	// That is, if we call cloudProvider then serviceMesh here, we _must_ call serviceMesh then cloudProvider
	// in isDeploymentSpecUpdated.
	podTemplateSpec, err = c.serviceMesh.TransformWorkloadPodTemplateSpec(jobRun, podTemplateSpec)
	if err != nil {
		return batchv1.JobSpec{}, err
	}

	podTemplateSpec = c.cloudProvider.TransformPodTemplateSpec(podTemplateSpec)

	kubeJobSpec := batchv1.JobSpec{
		Parallelism:  &one,
		Completions:  &one,
		BackoffLimit: &numRetries,
		Template:     *podTemplateSpec,
	}
	return kubeJobSpec, nil
}

func (c *Controller) untransformedPodTemplateSpec(
	jobRun *latticev1.JobRun,
	name string,
	jobRunLabels map[string]string,
	nodePool *latticev1.NodePool,
) (*corev1.PodTemplateSpec, error) {
	path, err := jobRun.PathLabel()
	if err != nil {
		err := fmt.Errorf("error getting path label for %v: %v", jobRun.Description(c.namespacePrefix), err)
		return nil, err
	}

	podAffinityTerm := corev1.PodAffinityTerm{
		LabelSelector: &metav1.LabelSelector{
			MatchLabels: jobRunLabels,
		},
		Namespaces: []string{jobRun.Namespace},

		// This basically tells the pod anti-affinity to only be applied to nodes who all
		// have the same value for that label.
		// Since we also add a RequiredDuringScheduling NodeAffinity for our NodePool,
		// this NodePool's nodes are the only nodes that these pods could be scheduled on,
		// so this TopologyKey doesn't really matter (besides being required).
		TopologyKey: latticev1.NodePoolIDLabelKey,
	}

	nodePoolEpoch, ok := nodePool.Status.Epochs.CurrentEpoch()
	if !ok {
		return nil, fmt.Errorf("unable to get current epoch for %v: %v", nodePool.Description(c.namespacePrefix), err)
	}

	affinity := &corev1.Affinity{
		NodeAffinity: nodePool.Affinity(nodePoolEpoch),
		PodAntiAffinity: &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight:          50,
					PodAffinityTerm: podAffinityTerm,
				},
			},
		},
	}

	tolerations := []corev1.Toleration{nodePool.Toleration(nodePoolEpoch)}

	// use the supplied command and environment as the job container's exec
	// if they are specified
	// copy so we don't mutate the cache
	jobRun = jobRun.DeepCopy()
	if jobRun.Spec.Definition.Exec == nil {
		jobRun.Spec.Definition.Exec = &definitionv1.ContainerExec{
			Command:     jobRun.Spec.Command,
			Environment: jobRun.Spec.Environment,
		}
	}

	if jobRun.Spec.Command != nil {
		jobRun.Spec.Definition.Exec.Command = jobRun.Spec.Command
	}

	if jobRun.Spec.Environment != nil {
		jobRun.Spec.Definition.Exec.Environment = jobRun.Spec.Environment
	}

	return latticev1.PodTemplateSpecForComponent(
		jobRun.Spec.Definition,
		path,
		c.latticeID,
		c.internalDNSDomain,
		c.namespacePrefix,
		jobRun.Namespace,
		jobRun.Name,
		jobRunLabels,
		jobRun.Spec.ContainerBuildArtifacts,
		corev1.RestartPolicyOnFailure,
		affinity,
		tolerations,
	)
}

func controllerRef(jobRun *latticev1.JobRun) *metav1.OwnerReference {
	return metav1.NewControllerRef(jobRun, latticev1.JobRunKind)
}
