package kubernetes

import (
	"fmt"
)

func GetFullNamespace(namespacePrefix, namespace string) string {
	return fmt.Sprintf("%v-%v", namespacePrefix, namespace)
}
