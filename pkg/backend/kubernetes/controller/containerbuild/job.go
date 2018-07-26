package containerbuild

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/util/docker"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/mlab-lattice/lattice/pkg/util/sha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

const (
	jobWorkingDirectory           = "/var/run/builder"
	jobWorkingDirectoryVolumeName = "workdir"

	jobDockerSocketVolumePath = "/var/run/docker.sock"
	jobDockerSocketPath       = "/var/run/docker.sock"
	jobDockerSocketVolumeName = "dockersock"
)

func (c *Controller) getJobForBuild(build *latticev1.ContainerBuild) (*batchv1.Job, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ContainerBuildIDLabelKey, selection.Equals, []string{build.Name})
	if err != nil {
		return nil, fmt.Errorf("error creating requirement for %v cache job lookup: %v", build.Description(c.namespacePrefix), err)
	}

	selector = selector.Add(*requirement)
	jobs, err := c.jobLister.Jobs(build.Namespace).List(selector)
	if err != nil {
		return nil, err
	}

	if len(jobs) == 1 {
		return jobs[0], nil
	}

	if len(jobs) > 1 {
		return nil, fmt.Errorf("multiple cached jobs found for %v", build.Description(c.namespacePrefix))
	}

	// Didn't find anything in the cache. Will do a full API query to see if one exists.
	jobList, err := c.kubeClient.BatchV1().Jobs(build.Namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, fmt.Errorf("error creating requirement for %v quorum job lookup: %v", build.Description(c.namespacePrefix), err)
	}

	if len(jobList.Items) == 0 {
		return nil, nil
	}

	if len(jobList.Items) > 1 {
		return nil, fmt.Errorf("multiple jobs found for %v", build.Description(c.namespacePrefix))
	}

	return &jobList.Items[0], nil
}

func (c *Controller) createNewJob(build *latticev1.ContainerBuild) (*batchv1.Job, error) {
	job, err := c.newJob(build)
	if err != nil {
		return nil, fmt.Errorf("error getting new job for %v: %v", build.Description(c.namespacePrefix), err)
	}

	result, err := c.kubeClient.BatchV1().Jobs(build.Namespace).Create(job)
	if err != nil {
		return nil, fmt.Errorf("error creating new job for %v: %v", build.Description(c.namespacePrefix), err)
	}

	return result, nil
}

func (c *Controller) newJob(build *latticev1.ContainerBuild) (*batchv1.Job, error) {
	// Need a consistent view of our config while generating the Job
	c.configLock.RLock()
	defer c.configLock.RUnlock()

	name := jobName(build)

	spec, dockerImageFQN, err := c.jobSpec(build)
	if err != nil {
		return nil, err
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				latticev1.ContainerBuildJobDockerImageFQNAnnotationKey: dockerImageFQN,
			},
			Labels: map[string]string{
				latticev1.ContainerBuildIDLabelKey: build.Name,
			},
			Name:            name,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(build, latticev1.ContainerBuildKind)},
		},
		Spec: spec,
	}
	return job, nil
}

func jobName(build *latticev1.ContainerBuild) string {
	return fmt.Sprintf("lattice-container-build-%s", build.Name)
}

func (c *Controller) jobSpec(build *latticev1.ContainerBuild) (batchv1.JobSpec, string, error) {
	buildContainer, dockerImageFQN, err := c.getBuildContainer(build)
	if err != nil {
		return batchv1.JobSpec{}, "", err
	}

	name := jobName(build)
	workingDirectoryVolumeSource := c.cloudProvider.ComponentBuildWorkDirectoryVolumeSource(name)

	var zero int32
	spec := &batchv1.JobSpec{
		// Only †ry to run the build once
		BackoffLimit: &zero,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					latticev1.ContainerBuildIDLabelKey: build.Name,
				},
			},
			Spec: corev1.PodSpec{
				Tolerations: []corev1.Toleration{
					// Can tolerate build node taint even in local case
					constants.TolerationBuildNode,
				},
				Volumes: []corev1.Volume{
					{
						Name:         jobWorkingDirectoryVolumeName,
						VolumeSource: workingDirectoryVolumeSource,
					},
					{
						Name: jobDockerSocketVolumeName,
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: jobDockerSocketVolumePath,
							},
						},
					},
				},
				Containers:         []corev1.Container{*buildContainer},
				ServiceAccountName: constants.ServiceAccountComponentBuilder,
				RestartPolicy:      corev1.RestartPolicyNever,
				DNSPolicy:          corev1.DNSDefault,
				Affinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &constants.NodeSelectorBuildNode,
					},
				},
			},
		},
	}

	spec = c.cloudProvider.TransformComponentBuildJobSpec(spec)

	return *spec, dockerImageFQN, nil
}

func (c *Controller) getBuildContainer(build *latticev1.ContainerBuild) (*corev1.Container, string, error) {
	buildJSON, err := json.Marshal(&build.Spec.Definition)
	if err != nil {
		return nil, "", err
	}

	repo := c.config.ContainerBuild.DockerArtifact.Repository
	tag, ok := build.DefinitionHashLabel()
	if !ok {
		err := fmt.Errorf(
			"%v does not have %v label",
			build.Description(c.namespacePrefix),
			latticev1.ContainerBuildDefinitionHashLabelKey,
		)
		return nil, "", err
	}

	if c.config.ContainerBuild.DockerArtifact.RepositoryPerImage {
		repo = tag
		tag = fmt.Sprint(time.Now().Unix())
	}

	systemID, err := kubeutil.SystemID(c.namespacePrefix, build.Namespace)
	if err != nil {
		return nil, "", err
	}

	args := []string{
		"--container-build-id", build.Name,
		"--namespace-prefix", c.namespacePrefix,
		"--system-id", string(systemID),
		"--container-build-definition", string(buildJSON),
		"--docker-registry", c.config.ContainerBuild.DockerArtifact.Registry,
		"--docker-repository", repo,
		"--docker-tag", tag,
		"--work-directory", jobWorkingDirectory,
	}

	if c.config.ContainerBuild.DockerArtifact.Push {
		args = append(args, "--docker-push")
	}

	if c.config.ContainerBuild.DockerArtifact.RegistryAuthType != nil {
		args = append(args, "--docker-registry-auth-type", *c.config.ContainerBuild.DockerArtifact.RegistryAuthType)
	}

	buildContainer := &corev1.Container{
		Name:  "build",
		Image: c.config.ContainerBuild.Builder.Image,
		Args:  args,
		Env: []corev1.EnvVar{
			{
				Name:  docker.APIVersionEnvironmentVariable,
				Value: c.config.ContainerBuild.Builder.DockerAPIVersion,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      jobWorkingDirectoryVolumeName,
				MountPath: jobWorkingDirectory,
			},
			{
				Name:      jobDockerSocketVolumeName,
				MountPath: jobDockerSocketPath,
			},
		},
	}

	if err := maybeSetSSSHKey(build, buildContainer); err != nil {
		return nil, "", err
	}

	dockerImageFQN := fmt.Sprintf(
		"%v/%v:%v",
		c.config.ContainerBuild.DockerArtifact.Registry,
		repo,
		tag,
	)

	return buildContainer, dockerImageFQN, nil
}

func maybeSetSSSHKey(build *latticev1.ContainerBuild, container *corev1.Container) error {
	def := build.Spec.Definition
	if def.CommandBuild == nil {
		return nil
	}

	cb := def.CommandBuild
	if cb.Source == nil {
		return nil
	}

	s := cb.Source
	if s.GitRepository == nil {
		return nil
	}

	sshKeySecret := s.GitRepository.SSHKey
	if sshKeySecret == nil {
		return nil
	}

	secretName, err := sha1.EncodeToHexString([]byte(sshKeySecret.NodePath().String()))
	if err != nil {
		return err
	}

	container.Env = append(container.Env, corev1.EnvVar{
		Name: "GIT_REPO_SSH_KEY",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secretName,
				},
				Key: sshKeySecret.Subcomponent(),
			},
		},
	})

	return nil
}

func jobStatus(j *batchv1.Job) (finished bool, succeeded bool) {
	for _, c := range j.Status.Conditions {
		if c.Type == batchv1.JobComplete && c.Status == corev1.ConditionTrue {
			finished = true
			succeeded = true
			return
		}
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			finished = true
			return
		}
	}
	return
}
