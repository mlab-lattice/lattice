package dnscontroller

import (
	"flag"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"

	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	fakelattice "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned/fake"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions"

	"github.com/mlab-lattice/system/pkg/definition/tree"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	core "k8s.io/client-go/testing"
)

const (
	serverConfigPath = "./server_config"
	hostConfigPath   = "./host_config"
	defaultNamespace = metav1.NamespaceDefault

	processQueueTimeout   = true
	processTimeoutSeconds = 2

	logToStderr         = true
	loggingLevelDefault = "10"
)

type hostEntry struct {
	host string
	ip   string
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
		newLine := v.ip + " " + v.host + "\n"
		str = str + newLine
	}

	return str
}

type cnameEntry struct {
	original string
	alias    string
}

// CnameFileOutput returns the expected file output that the server configuration file should contain.
func CnameFileOutput(nameservers []cnameEntry) string {
	/*
	   Expected format:
	       cname=original,alias
	*/
	str := ""

	for _, v := range nameservers {
		newLine := "cname=" + v.original + "," + v.alias + "\n"
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
			IP:               &ip,
			ExternalEndpoint: &endpoint,
			Path:             path,
		},
	}

	// Unset some fields instead of using the empty string
	if ip == "" {
		ec.Spec.IP = nil
	}

	if endpoint == "" {
		ec.Spec.ExternalEndpoint = nil
	}

	return ec
}

//AlterResourceVersion changes and exisintg endpoints resource string
func AlterResourceVersion(ep *latticev1.Endpoint, res_ver string) *latticev1.Endpoint {
	ep.ResourceVersion = res_ver

	return ep
}

//AlterEndpointState adjust and existing endpoints state
func AlterEndpointState(ep *latticev1.Endpoint, newState latticev1.EndpointState) *latticev1.Endpoint {
	ep.Status = latticev1.EndpointStatus{newState}

	return ep
}

//AddDeletionTimestamp adds a deletion timestamp to an existing endpoint
func AddDeletionTimestamp(ep *latticev1.Endpoint) *latticev1.Endpoint {
	now_time := v1.NewTime(time.Now())

	ep.SetDeletionTimestamp(&now_time)

	return ep
}

// MarkEndpointCreated alters the state of an Endpoint to 'Created'
func MarkEndpointCreated(endpoint *latticev1.Endpoint) *latticev1.Endpoint {
	e := endpoint
	e.Status = latticev1.EndpointStatus{
		State: latticev1.EndpointStateCreated,
	}

	return e
}

// MakeNdoePathPanic tries to create a NodePath from the url string, and panics is it is unable to.
func MakeNodePathPanic(pathString string) tree.NodePath {
	np, err := tree.NewNodePath(pathString)

	if err != nil {
		panic(err)
	}

	return np
}

type reaction struct {
	verb     string
	resource string
	reactor  func(t *testing.T) core.ReactionFunc
}

type test_case struct {
	ClientObjects []runtime.Object

	IsAsync    bool
	MaxRetries int

	// Reactor determines how the controller responds to certain verb actions with a resource.
	Reactors []reaction

	ExistingEndpoints *latticev1.EndpointList

	AddedEndpoints          *latticev1.EndpointList
	UpdatedEndpoint         *latticev1.Endpoint
	UpdatedEndpointPrevious *latticev1.Endpoint
	DeletedEndpoint         *latticev1.Endpoint

	ExpectedActions []core.Action
	ExpectedHosts   []hostEntry
	ExpectedCnames  []cnameEntry
}

// TestEndpointCreation tests the default resource CRUD and output of Endpoint controller operations, including the DNS and cname file contents
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
			AddedEndpoints: EndpointList(
				*Endpoint("key", "1", "", MakeNodePathPanic("/nodepath"))),
			ExpectedHosts: []hostEntry{
				{
					ip:   "1",
					host: "nodepath",
				},
			},
		},
		"new endpoint with name is written as cname file": {
			AddedEndpoints: EndpointList(
				*Endpoint("key", "", "my_cname", MakeNodePathPanic("/nodepath"))),
			ExpectedCnames: []cnameEntry{
				{
					alias:    "my_cname",
					original: "nodepath",
				},
			},
		},
		// TODO :: This could probably be moved to a nodepath test but there arent any right now
		"endpoints write url correctly": {
			AddedEndpoints: EndpointList(
				*Endpoint("key", "", "my_cname", MakeNodePathPanic("/root/nested/nested_some_more")),
				*Endpoint("key2", "1", "", MakeNodePathPanic("/root/nested/nested_again_but_different")),
			),
			ExpectedCnames: []cnameEntry{
				{
					alias:    "my_cname",
					original: "nested_some_more.nested.root",
				},
			},
			ExpectedHosts: []hostEntry{
				{
					ip:   "1",
					host: "nested_again_but_different.nested.root",
				},
			},
		},
		"new endpoint with existing deletion timestamp immediately added as tombstone": {
			AddedEndpoints: EndpointList(
				*AddDeletionTimestamp(Endpoint("key", "", "my_cname", MakeNodePathPanic("/nodepath")))),
			ExpectedActions: []core.Action{
			//Endpoint should be removed from the store, noa actions expected.
			},
		},
		"normal endpoint update changes the underlying endpoint": {
			AddedEndpoints: EndpointList(
				*Endpoint("key", "", "my_cname", MakeNodePathPanic("/nodepath")),
			),
			UpdatedEndpointPrevious: Endpoint("key", "", "my_cname", MakeNodePathPanic("/nodepath")),
			UpdatedEndpoint:         AlterResourceVersion(Endpoint("key", "", "my_new_cname", MakeNodePathPanic("/nodepath_ver2")), "2"),
			ExpectedActions: []core.Action{
				core.NewUpdateAction(latticev1.GroupVersionResource("endpoints"), metav1.NamespaceDefault,
					AlterEndpointState(
						AlterResourceVersion(
							Endpoint("key", "", "my_new_cname", MakeNodePathPanic("/nodepath_ver2")), "2"), latticev1.EndpointStateCreated)),
			},
		},
		"endpoint update can change between cname and IP address type endpoint": {
			AddedEndpoints: EndpointList(
				*Endpoint("key", "", "first_an_alias", MakeNodePathPanic("/nodepath")),
			),
			UpdatedEndpointPrevious: Endpoint("key", "", "first_an_alias", MakeNodePathPanic("/nodepath")),
			UpdatedEndpoint:         AlterResourceVersion(Endpoint("key", "5.5.5.5", "", MakeNodePathPanic("/nodepath")), "2"),
			ExpectedCnames:          []cnameEntry{
			//Expect empty cname
			},
			ExpectedHosts: []hostEntry{
				{
					ip:   "5.5.5.5",
					host: "nodepath",
				},
			},
		},
		"changing the underlying endpoint writes only the updated entry": {
			AddedEndpoints: EndpointList(
				*Endpoint("key", "", "my_cname", MakeNodePathPanic("/nodepath")),
			),
			UpdatedEndpointPrevious: Endpoint("key", "", "my_cname", MakeNodePathPanic("/nodepath")),
			UpdatedEndpoint:         AlterResourceVersion(Endpoint("key", "", "my_new_cname", MakeNodePathPanic("/new/nodepath_ver2")), "2"),
			ExpectedCnames: []cnameEntry{
				{
					alias:    "my_new_cname",
					original: "nodepath_ver2.new",
				},
			},
		},
		"adding an endpoint in the created state performs no action": {
			AddedEndpoints: EndpointList(
				*AlterEndpointState(Endpoint("key", "", "my_cname", MakeNodePathPanic("/nodepath")), latticev1.EndpointStateCreated),
			),
			ExpectedActions: []core.Action{
			// No actions expected
			},
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

		client := fakelattice.NewSimpleClientset(tc.ClientObjects...)

		for _, reactor := range tc.Reactors {
			client.Fake.PrependReactor(reactor.verb, reactor.resource, reactor.reactor(t))
		}

		informers := latticeinformers.NewSharedInformerFactory(client, 0)
		endpointInformer := informers.Lattice().V1().Endpoints()
		endpoints := informers.Lattice().V1().Endpoints().Informer().GetStore()

		controller := NewController(controllerServerConfigPath, controllerHostConfigPath, client, endpointInformer)

		if tc.ExistingEndpoints != nil {
			for _, e := range tc.ExistingEndpoints.Items {
				s := e.DeepCopy()
				err := endpoints.Add(s)

				if err != nil {
					t.Fatal(err)
				}
			}
		}

		if tc.AddedEndpoints != nil {
			for _, v := range tc.AddedEndpoints.Items {
				s := v.DeepCopy()
				err := endpoints.Add(s)
				// event handlers need to be called manually in the test cases
				controller.addEndpoint(s)

				if err != nil {
					t.Fatal(err)
				}
			}
		}

		// FIXME :: Processing logic is separated from the storing logic.
		// any calls i.e. endpoints.Add() might not be processed in syncEndpointUpdate, because the store might be modified before
		// ProcessControllerQueue is called.
		// If the desired logic of the program is to have all updates immediately reflected within the process queue, the logic would need to be changed
		// away from teh Existing / Added / Updated test structure.
		// The current setup of manually calling ProcessControllerQueue at set intervals works for small sized unit tests, so this may be okay.

		// Processes AddedEndpoints.
		ProcessControllerQueue(t, k, tc, client, controller)

		if tc.UpdatedEndpoint != nil {
			endpoints.Update(tc.UpdatedEndpoint)
			controller.updateEndpoint(tc.UpdatedEndpointPrevious, tc.UpdatedEndpoint)
		}
		if tc.DeletedEndpoint != nil {
			endpoints.Delete(tc.UpdatedEndpoint)
			controller.deleteEndpoint(tc.DeletedEndpoint)
		}

		// Process the updates
		ProcessControllerQueue(t, k, tc, client, controller)

		if controller.queue.Len() > 0 {
			t.Errorf("%s: unexpected items in endpoint queue: %d", k, controller.queue.Len())
		}

		if tc.ExpectedCnames != nil && tc.ExpectedHosts != nil && tc.ExpectedActions != nil {
			return
		}

		err := controller.RewriteDnsmasqConfig()

		t.Logf("Writing to: %v", controller.hostConfigPath)

		if err != nil {
			t.Errorf("Error rewriting DNSConfig: %v", err)
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

		// Test actions after flushing the dns file. This is because updates are synchronised to the client after the DNS writes.
		if tc.ExpectedActions == nil {
			return
		}

		ProcessControllerQueue(t, k, tc, client, controller)

		glog.V(5).Infof("Got %v actions, expected %v", len(client.Actions()), len(tc.ExpectedActions))

		actions := client.Actions()
		for i, action := range actions {
			if len(tc.ExpectedActions) < i+1 {
				t.Errorf("%s: %d unexpected actions: %+v", k, len(actions)-len(tc.ExpectedActions), actions[i:])
				break
			}

			expectedAction := tc.ExpectedActions[i]
			if !reflect.DeepEqual(expectedAction, action) {
				t.Errorf("%s:\nExpected:\n%s\ngot:\n%s", k, spew.Sdump(expectedAction), spew.Sdump(action))
				continue
			}
		}

		if len(tc.ExpectedActions) > len(actions) {
			t.Errorf("%s: %d additional expected actions", k, len(tc.ExpectedActions)-len(actions))
			for _, a := range tc.ExpectedActions[len(actions):] {
				t.Logf("    %+v", a)
			}
		} else if len(actions) > len(tc.ExpectedActions) {
			t.Errorf("%s: %d additional unexpected actions", k, len(actions)-len(tc.ExpectedActions))
			for _, a := range actions[len(tc.ExpectedActions):] {
				t.Logf("    %s", spew.Sdump(a))
			}
		}
	}
}

// ProcessControllerQueue runs an update loop on the queue until either the controller signals that the work is done, or the queue is empty.
func ProcessControllerQueue(t *testing.T, test_name string, tc test_case, client *fakelattice.Clientset, controller *Controller) {

	stopChannel := make(chan int)

	if processQueueTimeout {
		closeSecond := func() {
			stopChannel <- 0
		}

		// Must handle timeout in the main goroutine, so any fail should be handled in the main for loop below.
		go time.AfterFunc(time.Second*processTimeoutSeconds, closeSecond)
	}

	t.Logf("After flush, %v items in queue:", controller.queue.Len())

	for {
		if controller.queue.Len() > 0 {
			select {
			case _ = <-stopChannel:
				t.Fatalf("%s: reached ProcessControllerQueue timeout", test_name)
			default:
				if !controller.processNextWorkItem() {
					break
				}

				continue
			}

		}

		break
	}
}
