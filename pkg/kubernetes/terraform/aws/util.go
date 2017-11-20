package aws

import (
	"fmt"
)

func GetS3BackendStatePathRoot(systemId string) string {
	return fmt.Sprintf("lattice/terraform/state/%v", systemId)
}
