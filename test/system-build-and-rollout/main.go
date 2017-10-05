package main

import (
	"flag"
	"fmt"

	"github.com/mlab-lattice/core/pkg/constants"
	systemdefinition "github.com/mlab-lattice/core/pkg/system/definition"
	systemdefinitionblock "github.com/mlab-lattice/core/pkg/system/definition/block"

	"github.com/mlab-lattice/kubernetes-integration/pkg/system-environment-manager/backend"
)

var kubeconfig string

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	flag.Parse()

}

func main() {
	kb, err := backend.NewKubernetesBackend(kubeconfig)
	if err != nil {
		panic(err)
	}

	commit := "16d0ad5a7ef969b34174c39f12a588a38f4ff076"
	language := "node:boron"
	command := "npm install"
	instanceType := "t2.micro"
	sysDefinition := &systemdefinition.System{
		Meta: systemdefinitionblock.Metadata{
			Name: "my-system",
			Type: systemdefinition.SystemType,
		},
		Subsystems: []systemdefinition.Interface{
			systemdefinition.Interface(&systemdefinition.Service{
				Meta: systemdefinitionblock.Metadata{
					Name: "private-service",
					Type: systemdefinition.ServiceType,
				},
				Components: []*systemdefinitionblock.Component{
					{
						Name: "http",
						Ports: []*systemdefinitionblock.Port{
							{
								Name:     "http",
								Port:     9999,
								Protocol: systemdefinitionblock.HttpProtocol,
							},
						},
						Build: systemdefinitionblock.ComponentBuild{
							GitRepository: &systemdefinitionblock.GitRepository{
								Url:    "https://github.com/kevindrosendahl/example__hello-world-service-chaining",
								Commit: &commit,
							},
							Language: &language,
							Command:  &command,
						},
						Exec: systemdefinitionblock.Exec{
							Command: []string{
								"node",
								"lib/PrivateHelloService.js",
								"-p",
								"9999",
							},
						},
						HealthCheck: &systemdefinitionblock.HealthCheck{
							Http: &systemdefinitionblock.HttpHealthCheck{
								Path: "/status",
								Port: "http",
							},
						},
					},
				},
				Resources: systemdefinitionblock.Resources{
					MinInstances: 1,
					MaxInstances: 1,
					InstanceType: &instanceType,
				},
			}),
			systemdefinition.Interface(&systemdefinition.Service{
				Meta: systemdefinitionblock.Metadata{
					Name: "public-service",
					Type: systemdefinition.ServiceType,
				},
				Components: []*systemdefinitionblock.Component{
					{
						Name: "http",
						Ports: []*systemdefinitionblock.Port{
							{
								Name:     "http",
								Port:     8888,
								Protocol: systemdefinitionblock.HttpProtocol,
								ExternalAccess: &systemdefinitionblock.ExternalAccess{
									Public: true,
								},
							},
						},
						Build: systemdefinitionblock.ComponentBuild{
							GitRepository: &systemdefinitionblock.GitRepository{
								Url:    "https://github.com/kevindrosendahl/example__hello-world-service-chaining",
								Commit: &commit,
							},
							Language: &language,
							Command:  &command,
						},
						Exec: systemdefinitionblock.Exec{
							Command: []string{
								"node",
								"lib/PublicHelloService.js",
								"-p",
								"8888",
							},
							Environment: map[string]string{
								"PRIVATE_HELLO_SERVICE_URL": "http://private-service.my-system:9999",
							},
						},
						HealthCheck: &systemdefinitionblock.HealthCheck{
							Http: &systemdefinitionblock.HttpHealthCheck{
								Path: "/status",
								Port: "http",
							},
						},
					},
				},
				Resources: systemdefinitionblock.Resources{
					MinInstances: 1,
					MaxInstances: 1,
					InstanceType: &instanceType,
				},
			}),
		},
	}
	rid, err := kb.RollOutSystem(constants.UserSystemNamespace, sysDefinition, "v1.0.0")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created SystemRollout %v\n", rid)
}
