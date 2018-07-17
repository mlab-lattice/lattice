package v1

import (
	"fmt"
)

var (
	// SecretPath label is the key that should be used for the path of the secret.
	SecretPathLabelKey = fmt.Sprintf("secret.%v/path", GroupName)
)
