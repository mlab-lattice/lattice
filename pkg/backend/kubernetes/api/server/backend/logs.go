package backend

import (
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func toPodLogOptions(logOptions *v1.ContainerLogOptions) (*corev1.PodLogOptions, error) {

	podLogOptions := &corev1.PodLogOptions{
		Follow:       logOptions.Follow,
		TailLines:    logOptions.TailLines,
		Previous:     logOptions.Previous,
		Timestamps:   logOptions.Timestamps,
		SinceSeconds: logOptions.SinceSeconds,
	}

	if logOptions.SinceTime != "" {
		t, err := time.Parse(time.RFC3339, logOptions.SinceTime)
		if err != nil {
			return nil, err
		}
		sinceTime := metav1.NewTime(t)
		podLogOptions.SinceTime = &sinceTime
	}

	return podLogOptions, nil
}
