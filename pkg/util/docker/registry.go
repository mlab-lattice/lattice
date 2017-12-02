package docker

import (
	"net/url"
	"strings"
)

const (
	DockerRegistryAuthAWSEC2Role = "aws-ec2-role"

	RegistryTypeECR     = "ecr"
	RegistryTypeUnknown = "unknown"

	ecrDomain = "amazonaws"
	ecrTLD    = ".com"
)

func RegistryType(registry string) string {
	urlInfo, err := url.Parse(registry)
	if err != nil {
		return RegistryTypeUnknown
	}

	hostParts := strings.Split(urlInfo.Hostname(), ".")

	// Example ECR registry: <account_id>.dkr.ecr.us-east-1.amazonaws.com
	if len(hostParts) == 6 && hostParts[len(hostParts)-1] == ecrTLD && hostParts[len(hostParts)-2] == ecrDomain {
		return RegistryTypeECR
	}

	return RegistryTypeUnknown
}

type RegistryLoginProvider interface {
	GetLoginCredentials(registry string) (username, password string, err error)
}
