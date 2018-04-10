package aws

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

const (
	EC2RoleDockerRegistryAuth = "aws-ec2-role"
)

type ECRRegistryAuthProvider struct{}

func (erap *ECRRegistryAuthProvider) GetLoginCredentials(registryURI string) (string, string, error) {
	registry, err := parseECRRegistryURI(registryURI)
	if err != nil {
		return "", "", err
	}

	svc, err := getECR(registry.region)
	if err != nil {
		return "", "", err
	}

	input := &ecr.GetAuthorizationTokenInput{
		RegistryIds: []*string{&registry.accountID},
	}
	output, err := svc.GetAuthorizationToken(input)
	if err != nil {
		return "", "", err
	}

	if len(output.AuthorizationData) != 1 {
		return "", "", fmt.Errorf("expected 1 authorization token, got %v", len(output.AuthorizationData))
	}

	creds, err := base64.StdEncoding.DecodeString(*output.AuthorizationData[0].AuthorizationToken)
	splitCreds := strings.Split(string(creds), ":")
	if len(splitCreds) != 2 {
		return "", "", fmt.Errorf("unexpected format for authorization creds")
	}

	return splitCreds[0], splitCreds[1], nil
}

func getECR(region string) (*ecr.ECR, error) {
	sess, err := session.NewSession(&aws.Config{Region: &region})
	if err != nil {
		return nil, nil
	}

	return ecr.New(sess), nil
}

type ecrRegistry struct {
	accountID string
	region    string
}

func parseECRRegistryURI(registryURI string) (ecrRegistry, error) {
	// Example ECR registry: <account_id>.dkr.ecr.us-east-1.amazonaws.com
	registryParts := strings.Split(registryURI, ".")
	if len(registryParts) != 6 {
		return ecrRegistry{}, fmt.Errorf("invalid ECR registry URI: %v", registryURI)
	}

	registry := ecrRegistry{
		accountID: registryParts[0],
		region:    registryParts[3],
	}
	return registry, nil
}
