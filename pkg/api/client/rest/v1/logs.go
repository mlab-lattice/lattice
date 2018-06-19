package v1

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

func logOptionsToQueryString(logOptions *v1.ContainerLogOptions) string {
	qs := fmt.Sprintf(
		"follow=%v&timestamps=%v&previous=%v&since=%v&sinceTime=%v",
		logOptions.Follow,
		logOptions.Timestamps,
		logOptions.Previous,
		logOptions.Since,
		logOptions.SinceTime,
	)

	if logOptions.Tail != nil {
		qs += fmt.Sprintf("&tail=%v", *logOptions.Tail)
	}
	return qs
}
