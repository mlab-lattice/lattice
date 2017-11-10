package util

import (
	"fmt"
	"strconv"
	"strings"

	systemtree "github.com/mlab-lattice/core/pkg/system/tree"
)

func GetLocalClusterNameForComponentPort(svcPath systemtree.NodePath, componentName string, port int32) string {
	return fmt.Sprintf("local:%v", GetClusterNameForComponentPort(svcPath, componentName, port))
}

func GetClusterNameForComponentPort(svcPath systemtree.NodePath, componentName string, port int32) string {
	return fmt.Sprintf("%v:%v:%v", svcPath.ToDomain(false), componentName, port)
}

func GetPartsFromClusterName(clusterName string) (systemtree.NodePath, string, int32, error) {
	parts := strings.Split(clusterName, ":")
	if len(parts) != 3 {
		return "", "", 0, fmt.Errorf("unexpected cluster name format: %v", clusterName)
	}

	serviceDomain := parts[0]
	componentName := parts[1]
	portNumber, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil {
		return "", "", 0, err
	}

	path, err := systemtree.NodePathFromDomain(serviceDomain)
	if err != nil {
		return "", "", 0, err
	}

	return path, componentName, int32(portNumber), nil
}
