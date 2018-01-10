package dnscontroller

import (
	"flag"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"strconv"
	"testing"

	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	fakelattice "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned/fake"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	"github.com/mlab-lattice/system/pkg/definition/tree"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/kubernetes/fake"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
)

const (
	dnsmasqConfigPathPrefix = "./server_config"
	hostConfigPathPrefix    = "./host_config"

	// End of defaultNamespace should match defaultSystemID
	defaultNamespace = "namespace-filler-system-id"
	defaultSystemID  = "system-id"

	logToStderr         = true
	loggingLevelDefault = "10"

	clusterID = "cluster"
)

type hostEntry struct {
	host     string
	systemID string
	ip       string
}

type cnameEntry struct {
	original string
	systemID string
	alias    string
}

// hostFileOutput returns the expected file output that the host file should contain for the given hosts
func hostFileOutput(hosts []hostEntry) string {
	/*
	   Expected format:
	       ip name0 name1 ..
	       ...
	*/
	str := ""
	for _, v := range hosts {
		systemID := defaultSystemID

		if v.systemID != "" {
			// Allow overriding system id for certain tests.
			systemID = v.systemID
		}

		fullPath := v.host + ".local." + clusterID + "." + systemID + ".local"
		newLine := v.ip + " " + fullPath + "\n"
		str = str + newLine
	}
	return str
}

// dnsmasqConfigFileOutput returns the expected file output that the server configuration file should contain.
func dnsmasqConfigFileOutput(nameservers []cnameEntry) string {
	/*
	   Expected format:
	       cname=alias,target
	*/
	str := ""
	for _, v := range nameservers {
		systemID := defaultSystemID

		if v.systemID != "" {
			// Allow overriding system id for certain tests.
			systemID = v.systemID
		}

		fullPath := v.alias + ".local." + clusterID + "." + systemID + ".local"
		newLine := "cname=" + fullPath + "," + v.original + "\n"
		str = str + newLine
	}
	return str
}

// endpointList creates an endpointList schema from a list of endpoints.
func endpointList(endpoint ...latticev1.Endpoint) *latticev1.EndpointList {
	var el = latticev1.EndpointList{}

	for _, endp := range endpoint {
		el.Items = append(el.Items, endp)
	}

	return &el
}

// endpoint creates an endpoint schema with the specified parameters.
func endpoint(key string, ip string, endpoint string, path tree.NodePath) *latticev1.Endpoint {
	ec := &latticev1.Endpoint{
		ObjectMeta: metav1.ObjectMeta{
			// Our tests shouldn't be concerned about unique naming - let this be provided for us
			Name:            key,
			UID:             "12345",
			Namespace:       defaultNamespace,
			ResourceVersion: "1",
		},
		Status: latticev1.EndpointStatus{
			State: latticev1.EndpointStatePending,
		},
		Spec: latticev1.EndpointSpec{
			IP:           &ip,
			ExternalName: &endpoint,
			Path:         path,
		},
	}

	// Unset some fields instead of using the empty string
	if ip == "" {
		ec.Spec.IP = nil
	}

	if endpoint == "" {
		ec.Spec.ExternalName = nil
	}

	return ec
}

// alterNamespace adds a specified namespace to the Endpoints definition
func alterNamespace(namespace string, endpoint *latticev1.Endpoint) *latticev1.Endpoint {
	// First two hyphens in an endpoints namespace do not affect the output. Determined by the trailing string after the 2nd hyphen.
	endpoint.Namespace = "A-B-" + namespace

	return endpoint
}

// makeNdoePathPanic tries to create a NodePath from the url string, and panics is it is unable to.
func makeNodePathPanic(pathString string) tree.NodePath {
	np, err := tree.NewNodePath(pathString)

	if err != nil {
		panic(err)
	}

	return np
}

type test_case struct {
	ClientObjects []runtime.Object

	EndpointsBefore *latticev1.EndpointList
	EndpointsAfter  *latticev1.EndpointList

	ExpectedHosts  []hostEntry
	ExpectedCnames []cnameEntry
}

// TestEndpointCreation tests the DNS and cname file contents of the dns controller
func TestEndpointCreation(t *testing.T) {
	if logToStderr {
		flag.Set("alsologtostderr", fmt.Sprintf("%t", true))
	}

	var logLevel string
	flag.StringVar(&logLevel, "logLevel", loggingLevelDefault, "test")
	flag.Lookup("v").Value.Set(logLevel)

	testcases := map[string]test_case{
		"new endpoint with ip is written to host file": {
			EndpointsAfter: endpointList(
				*endpoint("key", "1", "", makeNodePathPanic("/nodepath"))),
			ExpectedHosts: []hostEntry{
				{
					ip:   "1",
					host: "nodepath",
				},
			},
		},
		"new endpoint with name is written as cname file": {
			EndpointsAfter: endpointList(
				*endpoint("key", "", "my_cname", makeNodePathPanic("/nodepath"))),
			ExpectedCnames: []cnameEntry{
				{
					original: "my_cname",
					alias:    "nodepath",
				},
			},
		},
		"endpoints write url correctly": {
			EndpointsAfter: endpointList(
				*endpoint("key", "", "my_cname", makeNodePathPanic("/root/nested/nested_some_more")),
				*endpoint("key2", "1", "", makeNodePathPanic("/root/nested/nested_again_but_different")),
			),
			ExpectedCnames: []cnameEntry{
				{
					original: "my_cname",
					alias:    "nested_some_more.nested.root",
				},
			},
			ExpectedHosts: []hostEntry{
				{
					ip:   "1",
					host: "nested_again_but_different.nested.root",
				},
			},
		},
		"normal endpoint update changes the underlying endpoint": {
			EndpointsBefore: endpointList(
				*endpoint("key", "", "my_cname", makeNodePathPanic("/nodepath")),
			),
			EndpointsAfter: endpointList(
				*endpoint("key", "", "my_cname_2", makeNodePathPanic("/root/nested/nested_some_more")),
			),
			ExpectedCnames: []cnameEntry{
				{
					original: "my_cname_2",
					alias:    "nested_some_more.nested.root",
				},
			},
			ExpectedHosts: []hostEntry{},
		},
		"endpoint update can change between cname and IP address type endpoint": {
			EndpointsBefore: endpointList(
				*endpoint("key", "", "my_cname", makeNodePathPanic("/nodepath")),
			),
			EndpointsAfter: endpointList(
				*endpoint("key", "5.5.5.5", "", makeNodePathPanic("/root/nested/nested_again_but_different")),
			),
			ExpectedCnames: []cnameEntry{},
			ExpectedHosts: []hostEntry{
				{
					ip:   "5.5.5.5",
					host: "nested_again_but_different.nested.root",
				},
			},
		},
		"rewriting the cache with no endpoints clears it": {
			EndpointsBefore: endpointList(
				*endpoint("key", "", "my_cname", makeNodePathPanic("/nodepath")),
			),
			EndpointsAfter: endpointList(),
			ExpectedCnames: []cnameEntry{},
			ExpectedHosts:  []hostEntry{},
		},
		"changing endpoint to a new namespace changes the output": {
			EndpointsBefore: endpointList(
				*endpoint("key", "", "my_cname", makeNodePathPanic("/nodepath")),
			),
			EndpointsAfter: endpointList(
				*alterNamespace("new-namespace", endpoint("key", "", "my_cname", makeNodePathPanic("/nodepath"))),
			),
			ExpectedCnames: []cnameEntry{
				{
					systemID: "new-namespace",
					alias:    "nodepath",
					original: "my_cname",
				},
			},
			ExpectedHosts: []hostEntry{},
		},
	}

	for testName, testCase := range testcases {
		glog.Infof(testName)

		// Write to different files on each iteration by using a hash of the test string
		hash := fnv.New32a()
		hash.Write([]byte(testName))
		pathSuffix := strconv.Itoa(int(hash.Sum32()))

		dnsmasqConfigPath := dnsmasqConfigPathPrefix + "_" + pathSuffix
		hostsFilePath := hostConfigPathPrefix + "_" + pathSuffix

		latticeClient := fakelattice.NewSimpleClientset(testCase.ClientObjects...)
		client := fake.NewSimpleClientset()

		informers := latticeinformers.NewSharedInformerFactory(latticeClient, 0)
		endpointInformer := informers.Lattice().V1().Endpoints()
		endpoints := informers.Lattice().V1().Endpoints().Informer().GetStore()

		controller := NewController(dnsmasqConfigPath, hostsFilePath, clusterID, latticeClient, client, endpointInformer)

		if testCase.EndpointsBefore != nil {
			for _, e := range testCase.EndpointsBefore.Items {
				s := e.DeepCopy()
				err := endpoints.Add(s)

				if err != nil {
					t.Fatal(err)
				}
			}
		}

		controller.updateConfigs()

		if testCase.EndpointsBefore != nil {
			for _, e := range testCase.EndpointsBefore.Items {
				s := e.DeepCopy()
				err := endpoints.Delete(s)

				if err != nil {
					t.Fatal(err)
				}
			}
		}

		if testCase.EndpointsAfter != nil {
			for _, v := range testCase.EndpointsAfter.Items {
				s := v.DeepCopy()
				err := endpoints.Add(s)

				if err != nil {
					t.Fatal(err)
				}
			}

			t.Logf("Updating cache and writing to: %v", controller.hostFilePath)
			controller.updateConfigs()
		}

		if controller.queue.Len() > 0 {
			t.Errorf("%s: unexpected items in endpoint queue: %d", testName, controller.queue.Len())
		}

		if testCase.ExpectedCnames != nil {
			dnsmasqConfig, err := ioutil.ReadFile(controller.dnsmasqConfigPath)
			if err != nil {
				t.Errorf("Error reading cname file: %v", err)
			}

			dnsmasqConfigStr := string(dnsmasqConfig)
			expectedDnsmasqConfigStr := dnsmasqConfigFileOutput(testCase.ExpectedCnames)

			if dnsmasqConfigStr != expectedDnsmasqConfigStr {
				t.Errorf("%s:\nExpected:\n%s\ngot:\n%s", testName, spew.Sdump(expectedDnsmasqConfigStr), spew.Sdump(dnsmasqConfigStr))
			}
		}

		if testCase.ExpectedHosts != nil {
			hostFile, err := ioutil.ReadFile(controller.hostFilePath)
			if err != nil {
				t.Errorf("Error reading host file: %v", err)
			}

			hostStr := string(hostFile)
			hostExpectedStr := hostFileOutput(testCase.ExpectedHosts)

			if hostStr != hostExpectedStr {
				t.Errorf("%s:\nExpected:\n%s\ngot:\n%s", testName, spew.Sdump(hostExpectedStr), spew.Sdump(hostStr))
			}
		}
	}
}
