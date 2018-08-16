package kubernetes

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"strings"
)

const (
	hackJobSidecarTerminationVolume          = "lattice-job-sidecar-termination"
	hackJobSidecarTerminationVolumeMountPath = "/tmp/lattice/job"
)

var (
	hackJobSidecarTerminationFilePath = fmt.Sprintf("%v/main-terminated", hackJobSidecarTerminationVolumeMountPath)
)

func HackJobSidecarTermination(spec *corev1.PodTemplateSpec) *corev1.PodTemplateSpec {
	// see https://github.com/kubernetes/kubernetes/issues/25908#issuecomment-308569672
	spec = spec.DeepCopy()

	var containers []corev1.Container
	for _, c := range spec.Spec.Containers {
		tmp := c.DeepCopy()
		tmp.Command = []string{"/bin/bash", "-c"}

		if tmp.VolumeMounts == nil {
			tmp.VolumeMounts = make([]corev1.VolumeMount, 0)
		}
		tmp.VolumeMounts = append(tmp.VolumeMounts, corev1.VolumeMount{
			MountPath: hackJobSidecarTerminationVolumeMountPath,
			Name:      hackJobSidecarTerminationVolume,
		})

		args := strings.Join(c.Command, " ")
		args = args + strings.Join(c.Args, " ")

		if c.Name == UserMainContainerName {
			tmp.Args = []string{fmt.Sprintf(`
trap "touch %v" EXIT
%v`, hackJobSidecarTerminationFilePath, args)}
		} else {
			tmp.VolumeMounts[len(tmp.VolumeMounts)-1].ReadOnly = true
			tmp.Args = []string{fmt.Sprintf(`
%v &
CHILD_PID=$!
(while true; do if [[ -f "%v" ]]; then kill $CHILD_PID; fi; sleep 1; done) &
wait $CHILD_PID
if [[ -f "%v" ]]; then exit 0; fi`, args, hackJobSidecarTerminationFilePath, hackJobSidecarTerminationFilePath)}
		}

		containers = append(containers, *tmp)
	}

	spec.Spec.Containers = containers

	if spec.Spec.Volumes == nil {
		spec.Spec.Volumes = make([]corev1.Volume, 0)
	}
	spec.Spec.Volumes = append(spec.Spec.Volumes, corev1.Volume{
		Name:         hackJobSidecarTerminationVolume,
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	})

	return spec
}
