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

	"k8s.io/client-go/kubernetes/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
)

const (
	serverConfigPath = "./server_config"
	hostConfigPath   = "./host_config"

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

// HostFileOutput returns the expected file output that the host file should contain for the given hosts
func HostFileOutput(hosts []hostEntry) string {
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

type cnameEntry struct {
	original string
	systemID string
	alias    string
}

// CnameFileOutput returns the expected file output that the server configuration file should contain.
func CnameFileOutput(nameservers []cnameEntry) string {
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

// EndpointList creates an EndpointList schema from a list of endpoints.
func EndpointList(endpoint ...latticev1.Endpoint) *latticev1.EndpointList {
	var el = latticev1.EndpointList{}

	for _, endp := range endpoint {
		el.Items = append(el.Items, endp)
	}

	return &el
}

// Endpoint creates an Endpoint schema with the specified parameters.
func Endpoint(key string, ip string, endpoint string, path tree.NodePath) *latticev1.Endpoint {
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

// AlterNamespace adds a specified namespace to the Endpoints definition
func AlterNamespace(namespace string, endpoint *latticev1.Endpoint) *latticev1.Endpoint {
	endpoint.Namespace = "A-B-" + namespace

	return endpoint
}

// MakeNdoePathPanic tries to create a NodePath from the url string, and panics is it is unable to.
func MakeNodePathPanic(pathString string) tree.NodePath {
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

// TestEndpointCreation tests the DNS and cname file contents of the dnscontroller
func TestEndpointCreation(t *testing.T) {

	if logToStderr {
		flag.Set("alsologtostderr", fmt.Sprintf("%t", true))
	}

	var logLevel string
	flag.StringVar(&logLevel, "logLevel", loggingLevelDefault, "test")
	flag.Lookup("v").Value.Set(logLevel)

	// Reduce DNS flush timer to more appropriate time
	updateWaitBeforeFlushTimerSeconds = 2

	testcases := map[string]test_case{
		"new endpoint with ip is written to host file": {
			EndpointsAfter: EndpointList(
				*Endpoint("key", "1", "", MakeNodePathPanic("/nodepath"))),
			ExpectedHosts: []hostEntry{
				{
					ip:   "1",
					host: "nodepath",
				},
			},
		},
		"new endpoint with name is written as cname file": {
			EndpointsAfter: EndpointList(
				*Endpoint("key", "", "my_cname", MakeNodePathPanic("/nodepath"))),
			ExpectedCnames: []cnameEntry{
				{
					original: "my_cname",
					alias:    "nodepath",
				},
			},
		},
		"endpoints write url correctly": {
			EndpointsAfter: EndpointList(
				*Endpoint("key", "", "my_cname", MakeNodePathPanic("/root/nested/nested_some_more")),
				*Endpoint("key2", "1", "", MakeNodePathPanic("/root/nested/nested_again_but_different")),
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
			EndpointsBefore: EndpointList(
				*Endpoint("key", "", "my_cname", MakeNodePathPanic("/nodepath")),
			),
			EndpointsAfter: EndpointList(
				*Endpoint("key", "", "my_cname_2", MakeNodePathPanic("/root/nested/nested_some_more")),
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
			EndpointsBefore: EndpointList(
				*Endpoint("key", "", "my_cname", MakeNodePathPanic("/nodepath")),
			),
			EndpointsAfter: EndpointList(
				*Endpoint("key", "5.5.5.5", "", MakeNodePathPanic("/root/nested/nested_again_but_different")),
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
			EndpointsBefore: EndpointList(
				*Endpoint("key", "", "my_cname", MakeNodePathPanic("/nodepath")),
			),
			EndpointsAfter: EndpointList(),
			ExpectedCnames: []cnameEntry{},
			ExpectedHosts:  []hostEntry{},
		},
		"changing endpoint to a new namespace changes the output": {
			EndpointsBefore: EndpointList(
				*Endpoint("key", "", "my_cname", MakeNodePathPanic("/nodepath")),
			),
			EndpointsAfter: EndpointList(
				*AlterNamespace("new-namespace", Endpoint("key", "", "my_cname", MakeNodePathPanic("/nodepath"))),
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

	for k, tc := range testcases {

		glog.Infof(k)

		// Write to different files on each iteration by using a hash of the test string
		hash := fnv.New32a()
		hash.Write([]byte(k))
		pathSuffix := strconv.Itoa(int(hash.Sum32()))

		controllerServerConfigPath := serverConfigPath + "_" + pathSuffix
		controllerHostConfigPath := hostConfigPath + "_" + pathSuffix

		latticeClient := fakelattice.NewSimpleClientset(tc.ClientObjects...)
		client := fake.NewSimpleClientset()

		informers := latticeinformers.NewSharedInformerFactory(latticeClient, 0)
		endpointInformer := informers.Lattice().V1().Endpoints()
		endpoints := informers.Lattice().V1().Endpoints().Informer().GetStore()

		controller := NewController(controllerServerConfigPath, controllerHostConfigPath, clusterID, latticeClient, client, endpointInformer)

		if tc.EndpointsBefore != nil {
			for _, e := range tc.EndpointsBefore.Items {
				s := e.DeepCopy()
				err := endpoints.Add(s)

				if err != nil {
					t.Fatal(err)
				}
			}
		}

		controller.calculateCache()

		if tc.EndpointsBefore != nil {
			for _, e := range tc.EndpointsBefore.Items {
				s := e.DeepCopy()
				err := endpoints.Delete(s)

				if err != nil {
					t.Fatal(err)
				}
			}
		}

		if tc.EndpointsAfter != nil {
			for _, v := range tc.EndpointsAfter.Items {
				s := v.DeepCopy()
				err := endpoints.Add(s)

				if err != nil {
					t.Fatal(err)
				}
			}

			t.Logf("Updating cache and writing to: %v", controller.hostConfigPath)
			controller.calculateCache()
		}

		if controller.queue.Len() > 0 {
			t.Errorf("%s: unexpected items in endpoint queue: %d", k, controller.queue.Len())
		}

		if tc.ExpectedCnames == nil && tc.ExpectedHosts == nil {
			return
		}

		if tc.ExpectedCnames != nil {
			cnameFile, err := ioutil.ReadFile(controller.serverConfigPath)

			if err != nil {
				t.Errorf("Error reading cname file: %v", err)
			}

			cnameStr := string(cnameFile)
			cnameExpectedStr := CnameFileOutput(tc.ExpectedCnames)

			if cnameStr != cnameExpectedStr {
				t.Errorf("%s:\nExpected:\n%s\ngot:\n%s", k, spew.Sdump(cnameExpectedStr), spew.Sdump(cnameStr))
			}
		}

		if tc.ExpectedHosts != nil {
			hostFile, err := ioutil.ReadFile(controller.hostConfigPath)

			if err != nil {
				t.Errorf("Error reading host file: %v", err)
			}

			hostStr := string(hostFile)
			hostExpectedStr := HostFileOutput(tc.ExpectedHosts)

			if hostStr != hostExpectedStr {
				t.Errorf("%s:\nExpected:\n%s\ngot:\n%s", k, spew.Sdump(hostExpectedStr), spew.Sdump(hostStr))
			}
		}
	}
}
