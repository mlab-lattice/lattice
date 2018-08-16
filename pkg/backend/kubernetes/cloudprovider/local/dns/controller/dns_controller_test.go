package controller

// FIXME <GEB>: refactor so that test services are more easily referenced by their associated
//              addresses. currently using list indices to cross-reference. this was not an issue
//              before because we were not managing leases dependent on protocol before and did
//              not need access to the service definition to determine the protocol.
// TODO <GEB>: add tests with TCP based service addresses

import (
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"testing"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	fakelattice "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned/fake"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/client-go/kubernetes/fake"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/satori/go.uuid"
)

const (
	dnsmasqConfigPathPrefix = "./server_config"
	hostConfigPathPrefix    = "./host_config"

	namespacePrefix   = "lattice"
	latticeID         = v1.LatticeID("lattice-local")
	internalDNSDomain = "lattice.local"
)

type hostEntry struct {
	systemID v1.SystemID
	name     string
	value    string
}

type cnameEntry struct {
	systemID v1.SystemID
	name     string
	value    string
}

// hostFileOutput returns the expected file output that the name file should contain for the given hosts
func hostFileOutput(entries []hostEntry) string {
	/*
	   Expected format:
	       value name0 name1 ..
	       ...
	*/
	str := ""
	for _, v := range entries {
		fullPath := fmt.Sprintf("%v.local.%v.%v.%v", v.name, v.systemID, latticeID, internalDNSDomain)
		newLine := v.value + " " + fullPath + "\n"
		str = str + newLine
	}
	return str
}

// dnsmasqConfigFileOutput returns the expected file output that the server configuration file should contain.
func dnsmasqConfigFileOutput(entries []cnameEntry) string {
	/*
	   Expected format:
	       cname=value,target
	*/
	str := ""
	for _, v := range entries {
		fullPath := fmt.Sprintf("%v.local.%v.%v.%v", v.name, v.systemID, latticeID, internalDNSDomain)
		newLine := "cname=" + fullPath + "," + v.value + "\n"
		str = str + newLine
	}
	return str
}

// address creates an address schema with the specified parameters.
func address(name string, addressPath tree.Path, servicePath *tree.Path, endpoint *string, systemID v1.SystemID) latticev1.Address {
	namespace := kubeutil.SystemNamespace(namespacePrefix, systemID)
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%v/%v", namespace, name)))

	return latticev1.Address{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			UID:             types.UID(base64.StdEncoding.EncodeToString(hash.Sum(nil))),
			ResourceVersion: "1",
			Labels: map[string]string{
				latticev1.AddressPathLabelKey: addressPath.ToDomain(),
			},
			Annotations: map[string]string{},
		},
		Spec: latticev1.AddressSpec{
			Service:      servicePath,
			ExternalName: endpoint,
		},
		Status: latticev1.AddressStatus{
			State: latticev1.AddressStateStable,
		},
	}
}

func service(path tree.Path, systemID v1.SystemID) latticev1.Service {
	return latticev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            uuid.NewV4().String(),
			Namespace:       kubeutil.SystemNamespace(namespacePrefix, systemID),
			UID:             types.UID(uuid.NewV4().String()),
			ResourceVersion: "1",
			Labels: map[string]string{
				latticev1.ServicePathLabelKey: path.ToDomain(),
			},
		},
		Spec: latticev1.ServiceSpec{
			Definition: &definitionv1.Service{
				Container: definitionv1.Container{
					Ports: map[int32]definitionv1.ContainerPort{
						8080: definitionv1.ContainerPort{
							Protocol: "HTTP",
						},
					},
				},
			},
		},
	}
}

func nodePathOrDie(path string) tree.Path {
	np, err := tree.NewPath(path)
	if err != nil {
		panic(err)
	}

	return np
}

type testCase struct {
	ClientObjects []runtime.Object

	Services []latticev1.Service

	AddressesBefore []latticev1.Address
	AddressesAfter  []latticev1.Address

	ExpectedHosts  []hostEntry
	ExpectedCnames []cnameEntry
}

// TestEndpointCreation tests the DNS and cname file contents of the dns controller
func TestAddressCreation(t *testing.T) {
	flag.Set("alsologtostderr", fmt.Sprintf("%t", true))

	var logLevel string
	flag.StringVar(&logLevel, "logLevel", "6", "test")
	flag.Lookup("v").Value.Set(logLevel)

	expectedIP, redirectCIDRBlock, _ := net.ParseCIDR("172.16.0.0/16")

	path1 := nodePathOrDie("/a")
	path2 := nodePathOrDie("/a/b/c")
	path3 := nodePathOrDie("/a/d/e/f")
	externalName1 := "mlab.com"
	externalName2 := "lattice.mlab.com"

	systemID1 := v1.SystemID(uuid.NewV4().String())
	systemID2 := v1.SystemID(uuid.NewV4().String())

	addressName1 := uuid.NewV4().String()
	addressName2 := uuid.NewV4().String()

	tests := map[string]testCase{
		"new address address is written to name file": {
			Services:       []latticev1.Service{service(path1, systemID1)},
			AddressesAfter: []latticev1.Address{address(addressName1, path1, &path1, nil, systemID1)},
			ExpectedHosts: []hostEntry{
				{
					systemID: systemID1,
					name:     path1.ToDomain(),
					value:    expectedIP.String(),
				},
			},
		},
		"new external name address is written as cname file": {
			AddressesAfter: []latticev1.Address{address(addressName1, path1, nil, &externalName1, systemID1)},
			ExpectedCnames: []cnameEntry{
				{
					systemID: systemID1,
					name:     path1.ToDomain(),
					value:    externalName1,
				},
			},
		},
		"addresses write url correctly": {
			Services: []latticev1.Service{service(path3, systemID1)},
			AddressesAfter: []latticev1.Address{
				// NOTE: the address order here matters (i.e., the service address's index must be the same as its service)
				address(addressName2, path3, &path3, nil, systemID1),
				address(addressName1, path2, nil, &externalName1, systemID1),
			},
			ExpectedCnames: []cnameEntry{
				{
					systemID: systemID1,
					name:     path2.ToDomain(),
					value:    externalName1,
				},
			},
			ExpectedHosts: []hostEntry{
				{
					systemID: systemID1,
					name:     path3.ToDomain(),
					value:    expectedIP.String(),
				},
			},
		},
		"address update changes the underlying address": {
			AddressesBefore: []latticev1.Address{address(addressName1, path2, nil, &externalName1, systemID1)},
			AddressesAfter:  []latticev1.Address{address(addressName1, path2, nil, &externalName2, systemID1)},
			ExpectedCnames: []cnameEntry{
				{
					systemID: systemID1,
					name:     path2.ToDomain(),
					value:    externalName2,
				},
			},
			ExpectedHosts: []hostEntry{},
		},
		"address update can change between external name and service address": {
			Services:        []latticev1.Service{service(path3, systemID1)},
			AddressesBefore: []latticev1.Address{address(addressName1, path2, nil, &externalName2, systemID1)},
			AddressesAfter:  []latticev1.Address{address(addressName1, path2, &path3, nil, systemID1)},
			ExpectedCnames:  []cnameEntry{},
			ExpectedHosts: []hostEntry{
				{
					systemID: systemID1,
					name:     path2.ToDomain(),
					value:    expectedIP.String(),
				},
			},
		},
		"rewriting the cache with no endpoints clears it": {
			AddressesBefore: []latticev1.Address{address(addressName1, path2, nil, &externalName1, systemID1)},
			AddressesAfter:  []latticev1.Address{},
			ExpectedCnames:  []cnameEntry{},
			ExpectedHosts:   []hostEntry{},
		},
		"changing address to a new namespace changes the output": {
			AddressesBefore: []latticev1.Address{address(addressName1, path2, nil, &externalName1, systemID1)},
			AddressesAfter:  []latticev1.Address{address(addressName1, path2, nil, &externalName1, systemID2)},
			ExpectedCnames: []cnameEntry{
				{
					systemID: systemID2,
					name:     path2.ToDomain(),
					value:    externalName1,
				},
			},
			ExpectedHosts: []hostEntry{},
		},
	}

	for description, test := range tests {
		glog.Infof(description)

		// Write to different files on each iteration by using a hash of the test string
		hash := fnv.New32a()
		hash.Write([]byte(description))
		pathSuffix := strconv.Itoa(int(hash.Sum32()))

		dnsmasqConfigPath := dnsmasqConfigPathPrefix + "_" + pathSuffix
		hostsFilePath := hostConfigPathPrefix + "_" + pathSuffix

		latticeClient := fakelattice.NewSimpleClientset(test.ClientObjects...)
		kubeClient := fake.NewSimpleClientset()

		informers := latticeinformers.NewSharedInformerFactory(latticeClient, 0)
		configInformer := informers.Lattice().V1().Configs()
		addressInformer := informers.Lattice().V1().Addresses()
		serviceInformer := informers.Lattice().V1().Services()

		serviceMeshOptions := &servicemesh.Options{
			Envoy: &envoy.Options{
				RedirectCIDRBlock: *redirectCIDRBlock,
			},
		}

		serviceMesh, err := servicemesh.NewServiceMesh(serviceMeshOptions)
		if err != nil {
			panic(err.Error())
		}
		serviceMeshUpdateIP := func(service *latticev1.Service, address *latticev1.Address) (*latticev1.Address, string) {
			var err error
			var ip string
			var annotations map[string]string

			address_ := address.DeepCopy()

			if address_.Spec.Service != nil {
				ip, err = serviceMesh.HasWorkloadIP(address_)
				if err == nil && ip == "" {
					ip, annotations, err = serviceMesh.WorkloadIP(service, address_)
					for k, v := range annotations {
						address_.Annotations[k] = v
					}
				}
			}

			if err != nil {
				t.Fatal(err)
			}

			return address_, ip
		}

		controller := NewController(
			latticeID,
			namespacePrefix,
			internalDNSDomain,
			dnsmasqConfigPath,
			hostsFilePath,
			serviceMeshOptions,
			latticeClient,
			kubeClient,
			configInformer,
			addressInformer,
			serviceInformer,
		)

		config := &latticev1.Config{
			Spec: latticev1.ConfigSpec{
				ServiceMesh: latticev1.ConfigServiceMesh{
					Envoy: &latticev1.ConfigServiceMeshEnvoy{},
				},
			},
		}
		configs := informers.Lattice().V1().Configs().Informer().GetStore()
		configs.Add(config)

		services := informers.Lattice().V1().Services().Informer().GetStore()
		for _, service := range test.Services {
			if err := services.Add(service.DeepCopy()); err != nil {
				t.Fatal(err)
			}
		}

		addresses := informers.Lattice().V1().Addresses().Informer().GetStore()
		for i, address := range test.AddressesBefore {
			address_ := address.DeepCopy()
			if len(test.Services) > i {
				address_, _ = serviceMeshUpdateIP(&test.Services[i], &address)
			}
			if err := addresses.Add(address_); err != nil {
				t.Fatal(err)
			}
		}

		controller.handleConfigAdd(config)
		controller.syncAddresses()

		for _, address := range test.AddressesBefore {
			if err := addresses.Delete(address.DeepCopy()); err != nil {
				t.Fatal(err)
			}
		}

		if test.AddressesAfter != nil {
			for i, address := range test.AddressesAfter {
				address_ := address.DeepCopy()
				if len(test.Services) > i {
					address_, _ = serviceMeshUpdateIP(&test.Services[i], &address)
				}
				if err := addresses.Add(address_); err != nil {
					t.Fatal(err)
				}
			}

			t.Logf("Updating cache and writing to: %v", hostsFilePath)
			controller.syncAddresses()
		}

		if test.ExpectedCnames != nil {
			dnsmasqConfig, err := ioutil.ReadFile(dnsmasqConfigPath)
			if err != nil {
				t.Errorf("Error reading cname file: %v", err)
				break
			}

			dnsmasqConfigStr := string(dnsmasqConfig)
			expectedDnsmasqConfigStr := dnsmasqConfigFileOutput(test.ExpectedCnames)

			if dnsmasqConfigStr != expectedDnsmasqConfigStr {
				t.Errorf("%s:\nExpected:\n%s\ngot:\n%s", description, spew.Sdump(expectedDnsmasqConfigStr), spew.Sdump(dnsmasqConfigStr))
			}

			err = os.Remove(dnsmasqConfigPath)
			if err != nil {
				runtimeutil.HandleError(err)
			}
		}

		if test.ExpectedHosts != nil {
			hostFile, err := ioutil.ReadFile(hostsFilePath)
			if err != nil {
				t.Errorf("Error reading name file: %v", err)
				break
			}

			hostStr := string(hostFile)
			hostExpectedStr := hostFileOutput(test.ExpectedHosts)

			if hostStr != hostExpectedStr {
				t.Errorf("%s:\nExpected:\n%s\ngot:\n%s", description, spew.Sdump(hostExpectedStr), spew.Sdump(hostStr))
			}

			err = os.Remove(hostsFilePath)
			if err != nil {
				runtimeutil.HandleError(err)
			}
		}
	}
}
