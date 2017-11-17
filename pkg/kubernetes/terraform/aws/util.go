package aws

import (
	"fmt"
)

func GetS3BackendStatePathRoot(systemId string) string {
	return fmt.Sprintf("/terraform/state/%v", systemId)
}
