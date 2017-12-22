package util

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mlab-lattice/system/pkg/definition/tree"
)

func GetLocalClusterNameForComponentPort(serviceCluster string, svcPath tree.NodePath, componentName string, port int32) string {
	return fmt.Sprintf("local:%v", GetClusterNameForComponentPort(serviceCluster, svcPath, componentName, port))
}

func GetClusterNameForComponentPort(serviceCluster string, svcPath tree.NodePath, componentName string, port int32) string {
	return fmt.Sprintf("%v:%v:%v:%v", serviceCluster, svcPath.ToDomain(false), componentName, port)
}

func GetPartsFromClusterName(clusterName string) (string, tree.NodePath, string, int32, error) {
	parts := strings.Split(clusterName, ":")
	if len(parts) != 4 {
		return "", "", "", 0, fmt.Errorf("unexpected cluster name format: %v", clusterName)
	}

	serviceCluster := parts[0]
	serviceDomain := parts[1]
	componentName := parts[2]
	portNumber, err := strconv.ParseInt(parts[3], 10, 32)
	if err != nil {
		return "", "", "", 0, err
	}

	path, err := tree.NodePathFromDomain(serviceDomain)
	if err != nil {
		return "", "", "", 0, err
	}

	return serviceCluster, path, componentName, int32(portNumber), nil
}
