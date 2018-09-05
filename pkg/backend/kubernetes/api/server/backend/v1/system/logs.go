package system

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var sinceRegex = regexp.MustCompile(`^([0-9]+)([smh])$`)

func toPodLogOptions(logOptions *v1.ContainerLogOptions) (*corev1.PodLogOptions, error) {

	podLogOptions := &corev1.PodLogOptions{
		Follow:       logOptions.Follow,
		TailLines:    logOptions.Tail,
		Previous:     logOptions.Previous,
		Timestamps:   logOptions.Timestamps,
		SinceSeconds: nil,
	}
	// sinceTime
	if logOptions.SinceTime != "" {
		t, err := time.Parse(time.RFC3339, logOptions.SinceTime)
		if err != nil {
			return nil, err
		}
		sinceTime := metav1.NewTime(t)
		podLogOptions.SinceTime = &sinceTime
	}

	// since
	if logOptions.Since != "" {
		sinceSeconds, err := parseSinceSeconds(logOptions.Since)
		if err != nil {
			return nil, err
		}
		podLogOptions.SinceSeconds = &sinceSeconds
	}

	return podLogOptions, nil
}

func parseSinceSeconds(since string) (int64, error) {
	parts := sinceRegex.FindAllStringSubmatch(since, -1)
	if len(parts) < 1 {
		return int64(0), fmt.Errorf("bad since expression: '%s'", since)
	}

	num, _ := strconv.ParseInt(parts[0][1], 10, 64)
	unit := parts[0][2]

	switch unit {
	case "s":
		return num, nil
	case "m":
		return num * 60, nil
	case "h":
		return num * 60 * 60, nil
	default:
		return int64(0), fmt.Errorf("unexcpected error while parsing dince '%s'", since)
	}

}
