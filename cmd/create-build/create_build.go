package main

import (
	"flag"
	"fmt"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	sdb "github.com/mlab-lattice/core/pkg/system/definition/block"
	crdclient "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource"
	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "Path to a kube config.")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	crClient, _, err := crdclient.NewClient(config)
	if err != nil {
		panic(err)
	}

	configList := crv1.ConfigList{}
	err = crClient.Get().Resource(crv1.ConfigResourcePlural).Do().Into(&configList)
	if err != nil {
		panic(err)
	}
	fmt.Printf("CONFIG LIST: %#v\n", configList)

	commit := "16d0ad5a7ef969b34174c39f12a588a38f4ff076"
	command := "npm install"
	language := "node:boron"
	build := &crv1.ComponentBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name: "example1",
		},
		Spec: crv1.ComponentBuildSpec{
			BuildDefinitionBlock: sdb.ComponentBuild{
				GitRepository: &sdb.GitRepository{
					Url:    "https://github.com/kevindrosendahl/example__hello-world-service-chaining",
					Commit: &commit,
				},
				Command:  &command,
				Language: &language,
			},
		},
		Status: crv1.ComponentBuildStatus{
			State:   crv1.ComponentBuildStatePending,
			Message: "Created, not processed yet",
		},
	}
	var buildResult crv1.ComponentBuild

	err = crClient.Post().
		Namespace("default").
		Resource(crv1.ComponentBuildResourcePlural).
		Body(build).
		Do().Into(&buildResult)
	if err == nil {
		fmt.Printf("CREATED BUILD: %#v\n", buildResult)
	} else if apierrors.IsAlreadyExists(err) {
		fmt.Printf("BUILD ALREADY EXISTS: %#v\n", buildResult)
	} else {
		panic(err)
	}

	buildList := crv1.ComponentBuildList{}
	err = crClient.Get().Resource(crv1.ComponentBuildResourcePlural).Do().Into(&buildList)
	if err != nil {
		panic(err)
	}
	fmt.Printf("BUILD LIST: %#v\n", buildList)
}
