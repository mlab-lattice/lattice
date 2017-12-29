package dnscontroller

import (
    "reflect"
    "testing"
    "time"
    "io/ioutil"
    "flag"
    "fmt"
    "hash/fnv"
    "strconv"

    "github.com/davecgh/go-spew/spew"
    "github.com/golang/glog"

    fakelattice "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned/fake"
    latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
  //  latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
    latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

    //"k8s.io/api/core/v1"
   // apierrors "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    //"k8s.io/apimachinery/pkg/runtime/schema"
    //"k8s.io/client-go/tools/cache"
  //  "k8s.io/client-go/informers"
    ///"k8s.io/client-go/kubernetes/fake"
    core "k8s.io/client-go/testing"
    //api "k8s.io/kubernetes/pkg/apis/core"
    //"k8s.io/kubernetes/pkg/controller"
    "github.com/mlab-lattice/system/pkg/definition/tree"
)

const(
    serverConfigPath = "./server_config"
    hostConfigPath = "./host_config"
)

type hostEntry struct {
    host string
    ip string
}

// HostFileOutput returns the expected file output that the host file should contain for the given hosts
func HostFileOutput(hosts []hostEntry) string {
    /*
        Expected format:
            name ip
            ...
     */
     str := ""

     for _, v := range hosts {
         newLine := v.host + " " + v.ip + "\n"
         str = str + newLine
     }

     return str
}

type cnameEntry struct {
    original string
    alias string
}

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

func EndpointList(endpoint ...latticev1.Endpoint) *latticev1.EndpointList {
    var el = latticev1.EndpointList{

    }

    for _, endp := range endpoint {
        el.Items = append(el.Items, endp)
    }

    return &el
}

func Endpoint(ip string, endpoint string, path tree.NodePath) *latticev1.Endpoint {
    return  &latticev1.Endpoint{
        ObjectMeta: metav1.ObjectMeta{
            Name:            "default",
            UID:             "12345",
            Namespace:       "default",
            ResourceVersion: "1",
        },
        Status: latticev1.EndpointStatus{
            State: latticev1.EndpointStateCreated,
        },
        Spec:latticev1.EndpointSpec{
            IP: &ip,
            ExternalEndpoint: &endpoint,
            Path: path,
        },
    }
}

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

func TestEndpointCreation(t *testing.T) {
    flag.Set("alsologtostderr", fmt.Sprintf("%t", true))
    var logLevel string
    flag.StringVar(&logLevel, "logLevel", "10", "test")
    flag.Lookup("v").Value.Set(logLevel)

    // Reduce DNS flush timer to more appropriate time
    updateWaitBeforeFlushTimer = 2

    testcases := map[string]struct {
        ClientObjects []runtime.Object

        IsAsync    bool
        MaxRetries int

        // Reactor determines how the controller responds to certain verb actions with a resource.
        Reactors []reaction

        ExistingEndpoints *latticev1.EndpointList

        AddedEndpoints   *latticev1.EndpointList
        UpdatedEndpoint *latticev1.Endpoint
        UpdatedEndpointPrevious * latticev1.Endpoint
        DeletedEndpoint *latticev1.Endpoint

        ExpectedActions []core.Action
        ExpectedHosts []hostEntry
        ExpectedCnames []cnameEntry
    }{
        "new endpoint created triggers DNS flush": {
            ClientObjects: []runtime.Object{},
            AddedEndpoints: EndpointList(
                *Endpoint("1", "", MakeNodePathPanic("/nodepath"))),
            ExpectedHosts: []hostEntry{
                {
                    ip: "1",
                    host: "nodepath",
                },
            },
        },
        //"new serviceaccount with no secrets encountering create error": {
        //    ClientObjects: []runtime.Object{serviceAccount(emptySecretReferences())},
        //    MaxRetries:    10,
        //    IsAsync:       true,
        //    Reactors: []reaction{{
        //        verb:     "create",
        //        resource: "secrets",
        //        reactor: func(t *testing.T) core.ReactionFunc {
        //            i := 0
        //            return func(core.Action) (bool, runtime.Object, error) {
        //                i++
        //                if i < 3 {
        //                    return true, nil, apierrors.NewForbidden(api.Resource("secrets"), "foo", errors.New("No can do"))
        //                }
        //                return false, nil, nil
        //            }
        //        },
        //    }},
        //    AddedServiceAccount: serviceAccount(emptySecretReferences()),
        //    ExpectedActions: []core.Action{
        //        // Attempt 1
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, "default"),
        //        core.NewCreateAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, createdTokenSecret()),
        //
        //        // Attempt 2
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, "default"),
        //        core.NewCreateAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, namedCreatedTokenSecret("default-token-txhzt")),
        //
        //        // Attempt 3
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, "default"),
        //        core.NewCreateAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, namedCreatedTokenSecret("default-token-vnmz7")),
        //        core.NewUpdateAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, serviceAccount(addNamedTokenSecretReference(emptySecretReferences(), "default-token-vnmz7"))),
        //    },
        //},
        //"new serviceaccount with no secrets encountering unending create error": {
        //    ClientObjects: []runtime.Object{serviceAccount(emptySecretReferences()), createdTokenSecret()},
        //    MaxRetries:    2,
        //    IsAsync:       true,
        //    Reactors: []reaction{{
        //        verb:     "create",
        //        resource: "secrets",
        //        reactor: func(t *testing.T) core.ReactionFunc {
        //            return func(core.Action) (bool, runtime.Object, error) {
        //                return true, nil, apierrors.NewForbidden(api.Resource("secrets"), "foo", errors.New("No can do"))
        //            }
        //        },
        //    }},

    }

    for k, tc := range testcases {

        glog.Infof(k)

        // Write to different files on each iteration
        hash := fnv.New32a()
        hash.Write([]byte(k))
        pathSuffix := strconv.Itoa(int(hash.Sum32()))

        controllerServerConfigPath := serverConfigPath + "_" + pathSuffix
        controllerHostConfigPath := hostConfigPath + "_" + pathSuffix

        client := fakelattice.NewSimpleClientset(tc.ClientObjects...)

        for _, reactor := range tc.Reactors {
            client.Fake.PrependReactor(reactor.verb, reactor.resource, reactor.reactor(t))
        }

        // Using controller noresyncfunc creates an include error
        informers := latticeinformers.NewSharedInformerFactory(client, time.Hour)
        endpointInformer := informers.Lattice().V1().Endpoints()
        endpoints := informers.Lattice().V1().Endpoints().Informer().GetStore()

        controller := NewController(controllerServerConfigPath, controllerHostConfigPath, client, endpointInformer)

        if tc.ExistingEndpoints != nil {
            for _, e := range tc.ExistingEndpoints.Items {
                endpoints.Add(&e)
            }
        }

        if tc.AddedEndpoints != nil {
            for _, v := range tc.AddedEndpoints.Items {
                endpoints.Add(&v)
                controller.addEndpoint(&v)
            }
        }
        if tc.UpdatedEndpoint != nil {
            endpoints.Update(tc.UpdatedEndpoint)
            controller.updateEndpoint(tc.UpdatedEndpointPrevious, tc.UpdatedEndpoint)
        }
        if tc.DeletedEndpoint != nil {
            endpoints.Delete(tc.UpdatedEndpoint)
            controller.deleteEndpoint(tc.DeletedEndpoint)
        }

        // This is the longest we'll wait for async tests
        timeout := time.Now().Add(30 * time.Second)
        waitedForAdditionalActions := false

        for {
            if controller.queue.Len() > 0 {
                key, done := controller.queue.Get()
                if done {
                    break
                }

                controller.syncEndpointUpdate(key.(string))
            }

            // The queues still have things to work on
            if controller.queue.Len() > 0 {
                continue
            }

            // If we expect this test to work asynchronously...
            if tc.IsAsync {
                // if we're still missing expected actions within our test timeout
                if len(client.Actions()) < len(tc.ExpectedActions) && time.Now().Before(timeout) {
                    // wait for the expected actions (without hotlooping)
                    time.Sleep(time.Millisecond)
                    continue
                }

                // if we exactly match our expected actions, wait a bit to make sure no other additional actions show up
                if len(client.Actions()) == len(tc.ExpectedActions) && !waitedForAdditionalActions {
                    time.Sleep(time.Second)
                    waitedForAdditionalActions = true
                    continue
                }
            }

            break
        }

        if controller.queue.Len() > 0 {
            t.Errorf("%s: unexpected items in endpoint queue: %d", k, controller.queue.Len())
        }

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
        }

        if tc.ExpectedCnames != nil || tc.ExpectedHosts != nil {

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
        }
    }
}
