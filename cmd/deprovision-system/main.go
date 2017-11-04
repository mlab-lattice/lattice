package main

import (
	"flag"
	"fmt"

	"github.com/mlab-lattice/kubernetes-integration/pkg/util/minikube"
)

const (
	workingDir = "/tmp/lattice/provision"
)

var (
	systemName string
)

func init() {
	flag.StringVar(&systemName, "system-name", "", "name of the system to provision")
	flag.Parse()
}

func main() {
	mec, err := minikube.NewMinikubeExecContext(workingDir)
	if err != nil {
		panic(err)
	}

	pid, logFilename, waitFunc, err := mec.Delete(systemName)
	if err != nil {
		panic(err)
	}

	fmt.Printf("minikube delete\npid: %v\nlogFilename: %v\n\n", pid, logFilename)

	err = waitFunc()
	if err != nil {
		panic(err)
	}
}
