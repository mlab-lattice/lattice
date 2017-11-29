package componentbuild

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"

	"github.com/mlab-lattice/system/pkg/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	workingDirectoryVolumeHostPathPrefixLocal = "/data/component-builder"
	workingDirectoryVolumeHostPathPrefixCloud = "/var/lib/component-builder"

	jobWorkingDirectory           = "/var/run/builder"
	jobWorkingDirectoryVolumeName = "workdir"

	jobDockerSocketVolumePath = "/var/run/docker.sock"
	jobDockerSocketPath       = "/var/run/docker.sock"
	jobDockerSocketVolumeName = "dockersock"

	jobDockerFqnAnnotationKey = "docker-image-fqn"
)

// getJobForBuild uses ControllerRefManager to retrieve the Job for a ComponentBuild
func (cbc *ComponentBuildController) getJobForBuild(cb *crv1.ComponentBuild) (*batchv1.Job, error) {
	// List all Jobs to find in the ComponentBuild's namespace to find the Job the ComponentBuild manages.
	jList, err := cbc.jobLister.Jobs(cb.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	matchingJobs := []*batchv1.Job{}
	cbControllerRef := metav1.NewControllerRef(cb, controllerKind)

	for _, j := range jList {
		jControllerRef := metav1.GetControllerOf(j)

		if reflect.DeepEqual(cbControllerRef, jControllerRef) {
			matchingJobs = append(matchingJobs, j)
		}
	}

	if len(matchingJobs) == 0 {
		return nil, nil
	}

	if len(matchingJobs) > 1 {
		return nil, fmt.Errorf("ComponentBuild %v has multiple Jobs", cb.Name)
	}

	return matchingJobs[0], nil
}

func (cbc *ComponentBuildController) getBuildJob(cb *crv1.ComponentBuild) (*batchv1.Job, error) {
	// Need a consistent view of our config while generating the Job
	cbc.configLock.RLock()
	defer cbc.configLock.RUnlock()

	name := getBuildJobName(cb)

	// FIXME: get job spec for build.DockerImage as well
	jSpec, dockerImageFqn, err := cbc.getGitRepositoryBuildJobSpec(cb)
	if err != nil {
		return nil, err
	}

	jLabels := map[string]string{
		crv1.ComponentBuildJobLabelKey: "true",
	}
	jAnnotations := map[string]string{
		jobDockerFqnAnnotationKey: dockerImageFqn,
	}

	j := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Annotations:     jAnnotations,
			Labels:          jLabels,
			Name:            name,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(cb, controllerKind)},
		},
		Spec: jSpec,
	}
	return j, nil
}

func getBuildJobName(cb *crv1.ComponentBuild) string {
	return fmt.Sprintf("lattice-build-%s", cb.Name)
}

func (cbc *ComponentBuildController) getGitRepositoryBuildJobSpec(cb *crv1.ComponentBuild) (batchv1.JobSpec, string, error) {
	buildContainer, dockerImageFQN, err := cbc.getBuildContainer(cb)
	if err != nil {
		return batchv1.JobSpec{}, "", err
	}

	name := getBuildJobName(cb)

	provider, err := crv1.GetProviderFromConfigSpec(cbc.config)
	if err != nil {
		return batchv1.JobSpec{}, "", err
	}

	var volumeHostPathPrefix string
	switch provider {
	case coreconstants.ProviderLocal:
		volumeHostPathPrefix = workingDirectoryVolumeHostPathPrefixLocal
	case coreconstants.ProviderAWS:
		volumeHostPathPrefix = workingDirectoryVolumeHostPathPrefixCloud
	default:
		panic(fmt.Sprintf("unsupported provider: %s", provider))
	}

	workingDirectoryVolumeSource := corev1.VolumeSource{
		HostPath: &corev1.HostPathVolumeSource{
			Path: volumeHostPathPrefix + "/" + name,
		},
	}

	// FIXME: add build node affinity for cloud case
	var zero int32 = 0
	jobSpec := batchv1.JobSpec{
		BackoffLimit: &zero,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
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
			},
		},
	}

	return jobSpec, dockerImageFQN, nil
}

func (cbc *ComponentBuildController) getBuildContainer(cb *crv1.ComponentBuild) (*corev1.Container, string, error) {
	componentBuildJson, err := json.Marshal(&cb.Spec.BuildDefinitionBlock)
	if err != nil {
		return nil, "", err
	}

	repo := cbc.config.ComponentBuild.DockerConfig.Repository
	tag := cb.Annotations[crv1.AnnotationKeyComponentBuildDefinitionHash]
	if cbc.config.ComponentBuild.DockerConfig.RepositoryPerImage {
		repo = cb.Annotations[crv1.AnnotationKeyComponentBuildDefinitionHash]
		tag = fmt.Sprint(time.Now().Unix())
	}

	args := []string{
		"--component-build-id", cb.Name,
		"--component-build-definition", string(componentBuildJson),
		"--docker-registry", cbc.config.ComponentBuild.DockerConfig.Registry,
		"--docker-repository", repo,
		"--docker-tag", tag,
		"--work-directory", jobWorkingDirectory,
	}

	if cbc.config.ComponentBuild.DockerConfig.Push {
		args = append(args, "--docker-push")
	}

	buildContainer := &corev1.Container{
		Name:  "build",
		Image: cbc.config.ComponentBuild.BuildImage,
		Args:  args,
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

	dockerImageFQN := fmt.Sprintf(
		"%v/%v:%v",
		cbc.config.ComponentBuild.DockerConfig.Registry,
		repo,
		tag,
	)

	return buildContainer, dockerImageFQN, nil
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
