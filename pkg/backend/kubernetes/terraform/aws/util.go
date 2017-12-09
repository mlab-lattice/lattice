package aws

import (
	"fmt"
)

func GetS3BackendStatePathRoot(systemID string) string {
	return fmt.Sprintf("lattice/terraform/state/%v", systemID)
}
