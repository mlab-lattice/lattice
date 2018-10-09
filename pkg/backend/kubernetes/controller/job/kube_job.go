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
	job *latticev1.Job,
	nodePool *latticev1.NodePool,
) (*batchv1.Job, error) {
	kubeJob, err := c.kubeJob(job)
	if err != nil {
		return nil, err
	}

	if kubeJob != nil {
		return kubeJob, nil
	}

	return c.createNewKubeJob(job, nodePool)
}

func (c *Controller) kubeJob(job *latticev1.Job) (*batchv1.Job, error) {
	// First check the cache for the deployment
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.JobIDLabelKey, selection.Equals, []string{job.Name})
	if err != nil {
		return nil, err
	}
	selector = selector.Add(*requirement)

	cachedKubeJobs, err := c.kubeJobLister.Jobs(job.Namespace).List(selector)
	if err != nil {
		return nil, err
	}

	if len(cachedKubeJobs) > 1 {
		// This may become valid when doing blue/green deploys
		return nil, fmt.Errorf("found multiple cached kube jobs for %v", job.Description(c.namespacePrefix))
	}

	if len(cachedKubeJobs) == 1 {
		return cachedKubeJobs[0], nil
	}

	// Didn't find the deployment in the cache. This likely means it hasn't been created, but since
	// we can't orphan kubeJobs, we need to do a quorum read first to ensure that the deployment
	// doesn't exist
	kubeJobs, err := c.kubeClient.BatchV1().Jobs(job.Namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	if len(kubeJobs.Items) > 1 {
		// This may become valid when doing blue/green deploys
		return nil, fmt.Errorf("found multiple kube jobs for %v", job.Description(c.namespacePrefix))
	}

	if len(kubeJobs.Items) == 1 {
		return &kubeJobs.Items[0], nil
	}

	return nil, nil
}

func (c *Controller) createNewKubeJob(
	job *latticev1.Job,
	nodePool *latticev1.NodePool,
) (*batchv1.Job, error) {
	kubeJob, err := c.newKubeJob(job, nodePool)
	if err != nil {
		return nil, err
	}

	result, err := c.kubeClient.BatchV1().Jobs(job.Namespace).Create(kubeJob)
	if err != nil {
		err := fmt.Errorf("error creating kube job for %v: %v", job.Description(c.namespacePrefix), err)
		return nil, err
	}

	return result, nil
}

func (c *Controller) newKubeJob(job *latticev1.Job, nodePool *latticev1.NodePool) (*batchv1.Job, error) {
	// Need a consistent view of our config while generating the kube job spec
	c.configLock.RLock()
	defer c.configLock.RUnlock()

	name := kubeJobName(job)
	kubeJobLabels := map[string]string{
		latticev1.JobIDLabelKey: job.Name,
	}

	spec, err := c.kubeJobSpec(job, name, kubeJobLabels, nodePool)
	if err != nil {
		err := fmt.Errorf(
			"error generating desired deployment spec for %v (on %v): %v",
			job.Description(c.namespacePrefix),
			nodePool.Description(c.namespacePrefix),
			err,
		)
		return nil, err
	}

	deployment := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Labels:          kubeJobLabels,
			OwnerReferences: []metav1.OwnerReference{*controllerRef(job)},
		},
		Spec: spec,
	}
	return deployment, nil
}

func kubeJobName(job *latticev1.Job) string {
	// TODO(kevindrosendahl): May change this to UUID when a Service can have multiple Deployments (e.g. Blue/Green & Canary)
	return fmt.Sprintf("lattice-job-%s", job.Name)
}

func (c *Controller) kubeJobSpec(
	job *latticev1.Job,
	name string,
	deploymentLabels map[string]string,
	nodePool *latticev1.NodePool,
) (batchv1.JobSpec, error) {
	podTemplateSpec, err := c.untransformedPodTemplateSpec(job, name, deploymentLabels, nodePool)
	if err != nil {
		return batchv1.JobSpec{}, err
	}

	one := int32(1)
	numRetries := int32(0)
	if job.Spec.NumRetries != nil {
		numRetries = *job.Spec.NumRetries
	}

	//return kubeJobSpec, nil
	// IMPORTANT: the order of these TransformServicePodTemplateSpec and the order of the IsDeploymentSpecUpdated calls in
	// isDeploymentSpecUpdated _must_ be inverses.
	// That is, if we call cloudProvider then serviceMesh here, we _must_ call serviceMesh then cloudProvider
	// in isDeploymentSpecUpdated.
	//podTemplateSpec, err = c.serviceMesh.TransformServicePodTemplateSpec(job, podTemplateSpec)
	//if err != nil {
	//	return batchv1.DeploymentSpec{}, err
	//}
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
	job *latticev1.Job,
	name string,
	jobLabels map[string]string,
	nodePool *latticev1.NodePool,
) (*corev1.PodTemplateSpec, error) {
	path, err := job.PathLabel()
	if err != nil {
		err := fmt.Errorf("error getting path label for %v: %v", job.Description(c.namespacePrefix), err)
		return nil, err
	}

	podAffinityTerm := corev1.PodAffinityTerm{
		LabelSelector: &metav1.LabelSelector{
			MatchLabels: jobLabels,
		},
		Namespaces: []string{job.Namespace},

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

	// copy so we don't mutate the cache
	job = job.DeepCopy()
	if job.Spec.Definition.Exec == nil {
		job.Spec.Definition.Exec = &definitionv1.ContainerExec{
			Environment: make(definitionv1.ContainerExecEnvironment),
		}
	}

	// if a command was passed in, use it instead of the definition's command
	if job.Spec.Command != nil {
		job.Spec.Definition.Exec.Command = job.Spec.Command
	}

	// if environment variables were passed in, set them
	for k, v := range job.Spec.Environment {
		job.Spec.Definition.Exec.Environment[k] = v
	}

	return latticev1.PodTemplateSpecForV1Workload(
		&job.Spec.Definition,
		path,
		c.latticeID,
		c.internalDNSDomain,
		c.namespacePrefix,
		job.Namespace,
		job.Name,
		jobLabels,
		job.Spec.ContainerBuildArtifacts,
		corev1.RestartPolicyNever,
		affinity,
		tolerations,
	)
}

func controllerRef(job *latticev1.Job) *metav1.OwnerReference {
	return metav1.NewControllerRef(job, latticev1.JobKind)
}
