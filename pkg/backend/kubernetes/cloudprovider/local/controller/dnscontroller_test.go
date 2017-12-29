package dnscontroller

import (
   // "errors"
    "reflect"
    "testing"
    "time"

    "github.com/davecgh/go-spew/spew"
    "github.com/golang/glog"

    fakelattice "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned/fake"
    latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
  //  latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
    latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

    "k8s.io/api/core/v1"
   // apierrors "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/runtime/schema"
    utilrand "k8s.io/apimachinery/pkg/util/rand"
  //  "k8s.io/client-go/informers"
    ///"k8s.io/client-go/kubernetes/fake"
    core "k8s.io/client-go/testing"
    //api "k8s.io/kubernetes/pkg/apis/core"
    //"k8s.io/kubernetes/pkg/controller"
    "github.com/mlab-lattice/system/pkg/definition/tree"
)

const(
    // TODO :: A test that ensures writing to these succeeds
    serverConfigPath = "/tmp/config"
    hostConfigPath = "/tmp/config"
)

//type testGenerator struct {
//    GeneratedServiceAccounts []v1.ServiceAccount
//    GeneratedSecrets         []v1.Secret
//    Token                    string
//    Err                      error
//}
//
//
//func (t *testGenerator) GenerateToken(serviceAccount v1.ServiceAccount, secret v1.Secret) (string, error) {
//    t.GeneratedSecrets = append(t.GeneratedSecrets, secret)
//    t.GeneratedServiceAccounts = append(t.GeneratedServiceAccounts, serviceAccount)
//    return t.Token, t.Err
//}

// emptySecretReferences is used by a service account without any secrets
func emptySecretReferences() []v1.ObjectReference {
    return []v1.ObjectReference{}
}

// addTokenSecretReference adds a reference to the ServiceAccountToken that will be created
func addTokenSecretReference(refs []v1.ObjectReference) []v1.ObjectReference {
    return addNamedTokenSecretReference(refs, "default-token-xn8fg")
}

// addNamedTokenSecretReference adds a reference to the named ServiceAccountToken
func addNamedTokenSecretReference(refs []v1.ObjectReference, name string) []v1.ObjectReference {
    return append(refs, v1.ObjectReference{Name: name})
}

// serviceAccount returns a service account with the given secret refs
func serviceAccount(secretRefs []v1.ObjectReference) *v1.ServiceAccount {
    return &v1.ServiceAccount{
        ObjectMeta: metav1.ObjectMeta{
            Name:            "default",
            UID:             "12345",
            Namespace:       "default",
            ResourceVersion: "1",
        },
        Secrets: secretRefs,
    }
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

// updatedServiceAccount returns a service account with the resource version modified
func updatedServiceAccount(secretRefs []v1.ObjectReference) *v1.ServiceAccount {
    sa := serviceAccount(secretRefs)
    sa.ResourceVersion = "2"
    return sa
}

// opaqueSecret returns a persisted non-ServiceAccountToken secret named "regular-secret-1"
func opaqueSecret() *v1.Secret {
    return &v1.Secret{
        ObjectMeta: metav1.ObjectMeta{
            Name:            "regular-secret-1",
            Namespace:       "default",
            UID:             "23456",
            ResourceVersion: "1",
        },
        Type: "Opaque",
        Data: map[string][]byte{
            "mykey": []byte("mydata"),
        },
    }
}

// createdTokenSecret returns the ServiceAccountToken secret posted when creating a new token secret.
// Named "default-token-xn8fg", since that is the first generated name after rand.Seed(1)
func createdTokenSecret(overrideName ...string) *v1.Secret {
    return namedCreatedTokenSecret("default-token-xn8fg")
}

// namedTokenSecret returns the ServiceAccountToken secret posted when creating a new token secret with the given name.
func namedCreatedTokenSecret(name string) *v1.Secret {
    return &v1.Secret{
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: "default",
            Annotations: map[string]string{
                v1.ServiceAccountNameKey: "default",
                v1.ServiceAccountUIDKey:  "12345",
            },
        },
        Type: v1.SecretTypeServiceAccountToken,
        Data: map[string][]byte{
            "token":     []byte("ABC"),
            "ca.crt":    []byte("CA Data"),
            "namespace": []byte("default"),
        },
    }
}

// serviceAccountTokenSecret returns an existing ServiceAccountToken secret named "token-secret-1"
func serviceAccountTokenSecret() *v1.Secret {
    return &v1.Secret{
        ObjectMeta: metav1.ObjectMeta{
            Name:            "token-secret-1",
            Namespace:       "default",
            UID:             "23456",
            ResourceVersion: "1",
            Annotations: map[string]string{
                v1.ServiceAccountNameKey: "default",
                v1.ServiceAccountUIDKey:  "12345",
            },
        },
        Type: v1.SecretTypeServiceAccountToken,
        Data: map[string][]byte{
            "token":     []byte("ABC"),
            "ca.crt":    []byte("CA Data"),
            "namespace": []byte("default"),
        },
    }
}

// serviceAccountTokenSecretWithoutTokenData returns an existing ServiceAccountToken secret that lacks token data
func serviceAccountTokenSecretWithoutTokenData() *v1.Secret {
    secret := serviceAccountTokenSecret()
    delete(secret.Data, v1.ServiceAccountTokenKey)
    return secret
}

// serviceAccountTokenSecretWithoutCAData returns an existing ServiceAccountToken secret that lacks ca data
func serviceAccountTokenSecretWithoutCAData() *v1.Secret {
    secret := serviceAccountTokenSecret()
    delete(secret.Data, v1.ServiceAccountRootCAKey)
    return secret
}

// serviceAccountTokenSecretWithCAData returns an existing ServiceAccountToken secret with the specified ca data
func serviceAccountTokenSecretWithCAData(data []byte) *v1.Secret {
    secret := serviceAccountTokenSecret()
    secret.Data[v1.ServiceAccountRootCAKey] = data
    return secret
}

// serviceAccountTokenSecretWithoutNamespaceData returns an existing ServiceAccountToken secret that lacks namespace data
func serviceAccountTokenSecretWithoutNamespaceData() *v1.Secret {
    secret := serviceAccountTokenSecret()
    delete(secret.Data, v1.ServiceAccountNamespaceKey)
    return secret
}

// serviceAccountTokenSecretWithNamespaceData returns an existing ServiceAccountToken secret with the specified namespace data
func serviceAccountTokenSecretWithNamespaceData(data []byte) *v1.Secret {
    secret := serviceAccountTokenSecret()
    secret.Data[v1.ServiceAccountNamespaceKey] = data
    return secret
}

type reaction struct {
    verb     string
    resource string
    reactor  func(t *testing.T) core.ReactionFunc
}

func TestEndpointCreation(t *testing.T) {
    testcases := map[string]struct {
        ClientObjects []runtime.Object

        IsAsync    bool
        MaxRetries int

        Reactors []reaction

        ExistingEndpoints *latticev1.EndpointList

        AddedEndpoint   *latticev1.Endpoint
        UpdatedEndpoint *latticev1.Endpoint
        DeletedEndpoint *latticev1.Endpoint

        ExpectedActions []core.Action
    }{
        "new endpoint created triggers DNS flush": {
            ClientObjects: []runtime.Object{},

            AddedEndpoint: Endpoint("1", "1", MakeNodePathPanic("/nodepath")),
            ExpectedActions: []core.Action{
                core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, "default"),
                core.NewCreateAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, createdTokenSecret()),
                core.NewUpdateAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, serviceAccount(addTokenSecretReference(emptySecretReferences()))),
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
        //
        //    AddedServiceAccount: serviceAccount(emptySecretReferences()),
        //    ExpectedActions: []core.Action{
        //        // Attempt
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, "default"),
        //        core.NewCreateAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, createdTokenSecret()),
        //        // Retry 1
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, "default"),
        //        core.NewCreateAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, namedCreatedTokenSecret("default-token-txhzt")),
        //        // Retry 2
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, "default"),
        //        core.NewCreateAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, namedCreatedTokenSecret("default-token-vnmz7")),
        //    },
        //},
        //"new serviceaccount with missing secrets": {
        //    ClientObjects: []runtime.Object{serviceAccount(missingSecretReferences())},
        //
        //    AddedServiceAccount: serviceAccount(missingSecretReferences()),
        //    ExpectedActions: []core.Action{
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, "default"),
        //        core.NewCreateAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, createdTokenSecret()),
        //        core.NewUpdateAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, serviceAccount(addTokenSecretReference(missingSecretReferences()))),
        //    },
        //},
        //"new serviceaccount with missing secrets and a local secret in the cache": {
        //    ClientObjects: []runtime.Object{serviceAccount(missingSecretReferences())},
        //
        //    AddedServiceAccount: serviceAccount(tokenSecretReferences()),
        //    AddedSecretLocal:    serviceAccountTokenSecret(),
        //    ExpectedActions:     []core.Action{},
        //},
        //"new serviceaccount with non-token secrets": {
        //    ClientObjects: []runtime.Object{serviceAccount(regularSecretReferences()), opaqueSecret()},
        //
        //    AddedServiceAccount: serviceAccount(regularSecretReferences()),
        //    ExpectedActions: []core.Action{
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, "default"),
        //        core.NewCreateAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, createdTokenSecret()),
        //        core.NewUpdateAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, serviceAccount(addTokenSecretReference(regularSecretReferences()))),
        //    },
        //},
        //"new serviceaccount with token secrets": {
        //    ClientObjects:   []runtime.Object{serviceAccount(tokenSecretReferences()), serviceAccountTokenSecret()},
        //    ExistingSecrets: []*v1.Secret{serviceAccountTokenSecret()},
        //
        //    AddedServiceAccount: serviceAccount(tokenSecretReferences()),
        //    ExpectedActions:     []core.Action{},
        //},
        //"new serviceaccount with no secrets with resource conflict": {
        //    ClientObjects: []runtime.Object{updatedServiceAccount(emptySecretReferences()), createdTokenSecret()},
        //    IsAsync:       true,
        //    MaxRetries:    1,
        //
        //    AddedServiceAccount: serviceAccount(emptySecretReferences()),
        //    ExpectedActions: []core.Action{
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, "default"),
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, "default"),
        //    },
        //},
        //"updated serviceaccount with no secrets": {
        //    ClientObjects: []runtime.Object{serviceAccount(emptySecretReferences())},
        //
        //    UpdatedServiceAccount: serviceAccount(emptySecretReferences()),
        //    ExpectedActions: []core.Action{
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, "default"),
        //        core.NewCreateAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, createdTokenSecret()),
        //        core.NewUpdateAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, serviceAccount(addTokenSecretReference(emptySecretReferences()))),
        //    },
        //},
        //"updated serviceaccount with missing secrets": {
        //    ClientObjects: []runtime.Object{serviceAccount(missingSecretReferences())},
        //
        //    UpdatedServiceAccount: serviceAccount(missingSecretReferences()),
        //    ExpectedActions: []core.Action{
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, "default"),
        //        core.NewCreateAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, createdTokenSecret()),
        //        core.NewUpdateAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, serviceAccount(addTokenSecretReference(missingSecretReferences()))),
        //    },
        //},
        //"updated serviceaccount with non-token secrets": {
        //    ClientObjects: []runtime.Object{serviceAccount(regularSecretReferences()), opaqueSecret()},
        //
        //    UpdatedServiceAccount: serviceAccount(regularSecretReferences()),
        //    ExpectedActions: []core.Action{
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, "default"),
        //        core.NewCreateAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, createdTokenSecret()),
        //        core.NewUpdateAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, serviceAccount(addTokenSecretReference(regularSecretReferences()))),
        //    },
        //},
        //"updated serviceaccount with token secrets": {
        //    ExistingSecrets: []*v1.Secret{serviceAccountTokenSecret()},
        //
        //    UpdatedServiceAccount: serviceAccount(tokenSecretReferences()),
        //    ExpectedActions:       []core.Action{},
        //},
        //"updated serviceaccount with no secrets with resource conflict": {
        //    ClientObjects: []runtime.Object{updatedServiceAccount(emptySecretReferences())},
        //    IsAsync:       true,
        //    MaxRetries:    1,
        //
        //    UpdatedServiceAccount: serviceAccount(emptySecretReferences()),
        //    ExpectedActions: []core.Action{
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, "default"),
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, "default"),
        //    },
        //},
        //
        //"deleted serviceaccount with no secrets": {
        //    DeletedServiceAccount: serviceAccount(emptySecretReferences()),
        //    ExpectedActions:       []core.Action{},
        //},
        //"deleted serviceaccount with missing secrets": {
        //    DeletedServiceAccount: serviceAccount(missingSecretReferences()),
        //    ExpectedActions:       []core.Action{},
        //},
        //"deleted serviceaccount with non-token secrets": {
        //    ClientObjects: []runtime.Object{opaqueSecret()},
        //
        //    DeletedServiceAccount: serviceAccount(regularSecretReferences()),
        //    ExpectedActions:       []core.Action{},
        //},
        //"deleted serviceaccount with token secrets": {
        //    ClientObjects:   []runtime.Object{serviceAccountTokenSecret()},
        //    ExistingSecrets: []*v1.Secret{serviceAccountTokenSecret()},
        //
        //    DeletedServiceAccount: serviceAccount(tokenSecretReferences()),
        //    ExpectedActions: []core.Action{
        //        core.NewDeleteAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, "token-secret-1"),
        //    },
        //},
        //
        //"added secret without serviceaccount": {
        //    ClientObjects: []runtime.Object{serviceAccountTokenSecret()},
        //
        //    AddedSecret: serviceAccountTokenSecret(),
        //    ExpectedActions: []core.Action{
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, "default"),
        //        core.NewDeleteAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, "token-secret-1"),
        //    },
        //},
        //"added secret with serviceaccount": {
        //    ExistingServiceAccount: serviceAccount(tokenSecretReferences()),
        //
        //    AddedSecret:     serviceAccountTokenSecret(),
        //    ExpectedActions: []core.Action{},
        //},
        //"added token secret without token data": {
        //    ClientObjects:          []runtime.Object{serviceAccountTokenSecretWithoutTokenData()},
        //    ExistingServiceAccount: serviceAccount(tokenSecretReferences()),
        //
        //    AddedSecret: serviceAccountTokenSecretWithoutTokenData(),
        //    ExpectedActions: []core.Action{
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, "token-secret-1"),
        //        core.NewUpdateAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, serviceAccountTokenSecret()),
        //    },
        //},
        //"added token secret without ca data": {
        //    ClientObjects:          []runtime.Object{serviceAccountTokenSecretWithoutCAData()},
        //    ExistingServiceAccount: serviceAccount(tokenSecretReferences()),
        //
        //    AddedSecret: serviceAccountTokenSecretWithoutCAData(),
        //    ExpectedActions: []core.Action{
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, "token-secret-1"),
        //        core.NewUpdateAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, serviceAccountTokenSecret()),
        //    },
        //},
        //"added token secret with mismatched ca data": {
        //    ClientObjects:          []runtime.Object{serviceAccountTokenSecretWithCAData([]byte("mismatched"))},
        //    ExistingServiceAccount: serviceAccount(tokenSecretReferences()),
        //
        //    AddedSecret: serviceAccountTokenSecretWithCAData([]byte("mismatched")),
        //    ExpectedActions: []core.Action{
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, "token-secret-1"),
        //        core.NewUpdateAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, serviceAccountTokenSecret()),
        //    },
        //},
        //"added token secret without namespace data": {
        //    ClientObjects:          []runtime.Object{serviceAccountTokenSecretWithoutNamespaceData()},
        //    ExistingServiceAccount: serviceAccount(tokenSecretReferences()),
        //
        //    AddedSecret: serviceAccountTokenSecretWithoutNamespaceData(),
        //    ExpectedActions: []core.Action{
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, "token-secret-1"),
        //        core.NewUpdateAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, serviceAccountTokenSecret()),
        //    },
        //},
        //"added token secret with custom namespace data": {
        //    ClientObjects:          []runtime.Object{serviceAccountTokenSecretWithNamespaceData([]byte("custom"))},
        //    ExistingServiceAccount: serviceAccount(tokenSecretReferences()),
        //
        //    AddedSecret:     serviceAccountTokenSecretWithNamespaceData([]byte("custom")),
        //    ExpectedActions: []core.Action{
        //        // no update is performed... the custom namespace is preserved
        //    },
        //},
        //
        //"updated secret without serviceaccount": {
        //    ClientObjects: []runtime.Object{serviceAccountTokenSecret()},
        //
        //    UpdatedSecret: serviceAccountTokenSecret(),
        //    ExpectedActions: []core.Action{
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, "default"),
        //        core.NewDeleteAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, "token-secret-1"),
        //    },
        //},
        //"updated secret with serviceaccount": {
        //    ExistingServiceAccount: serviceAccount(tokenSecretReferences()),
        //
        //    UpdatedSecret:   serviceAccountTokenSecret(),
        //    ExpectedActions: []core.Action{},
        //},
        //"updated token secret without token data": {
        //    ClientObjects:          []runtime.Object{serviceAccountTokenSecretWithoutTokenData()},
        //    ExistingServiceAccount: serviceAccount(tokenSecretReferences()),
        //
        //    UpdatedSecret: serviceAccountTokenSecretWithoutTokenData(),
        //    ExpectedActions: []core.Action{
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, "token-secret-1"),
        //        core.NewUpdateAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, serviceAccountTokenSecret()),
        //    },
        //},
        //"updated token secret without ca data": {
        //    ClientObjects:          []runtime.Object{serviceAccountTokenSecretWithoutCAData()},
        //    ExistingServiceAccount: serviceAccount(tokenSecretReferences()),
        //
        //    UpdatedSecret: serviceAccountTokenSecretWithoutCAData(),
        //    ExpectedActions: []core.Action{
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, "token-secret-1"),
        //        core.NewUpdateAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, serviceAccountTokenSecret()),
        //    },
        //},
        //"updated token secret with mismatched ca data": {
        //    ClientObjects:          []runtime.Object{serviceAccountTokenSecretWithCAData([]byte("mismatched"))},
        //    ExistingServiceAccount: serviceAccount(tokenSecretReferences()),
        //
        //    UpdatedSecret: serviceAccountTokenSecretWithCAData([]byte("mismatched")),
        //    ExpectedActions: []core.Action{
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, "token-secret-1"),
        //        core.NewUpdateAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, serviceAccountTokenSecret()),
        //    },
        //},
        //"updated token secret without namespace data": {
        //    ClientObjects:          []runtime.Object{serviceAccountTokenSecretWithoutNamespaceData()},
        //    ExistingServiceAccount: serviceAccount(tokenSecretReferences()),
        //
        //    UpdatedSecret: serviceAccountTokenSecretWithoutNamespaceData(),
        //    ExpectedActions: []core.Action{
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, "token-secret-1"),
        //        core.NewUpdateAction(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, metav1.NamespaceDefault, serviceAccountTokenSecret()),
        //    },
        //},
        //"updated token secret with custom namespace data": {
        //    ClientObjects:          []runtime.Object{serviceAccountTokenSecretWithNamespaceData([]byte("custom"))},
        //    ExistingServiceAccount: serviceAccount(tokenSecretReferences()),
        //
        //    UpdatedSecret:   serviceAccountTokenSecretWithNamespaceData([]byte("custom")),
        //    ExpectedActions: []core.Action{
        //        // no update is performed... the custom namespace is preserved
        //    },
        //},
        //
        //"deleted secret without serviceaccount": {
        //    DeletedSecret:   serviceAccountTokenSecret(),
        //    ExpectedActions: []core.Action{},
        //},
        //"deleted secret with serviceaccount with reference": {
        //    ClientObjects:          []runtime.Object{serviceAccount(tokenSecretReferences())},
        //    ExistingServiceAccount: serviceAccount(tokenSecretReferences()),
        //
        //    DeletedSecret: serviceAccountTokenSecret(),
        //    ExpectedActions: []core.Action{
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, "default"),
        //        core.NewUpdateAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, serviceAccount(emptySecretReferences())),
        //    },
        //},
        //"deleted secret with serviceaccount without reference": {
        //    ExistingServiceAccount: serviceAccount(emptySecretReferences()),
        //
        //    DeletedSecret: serviceAccountTokenSecret(),
        //    ExpectedActions: []core.Action{
        //        core.NewGetAction(schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, metav1.NamespaceDefault, "default"),
        //    },
        //},
    }

    for k, tc := range testcases {
        glog.Infof(k)

        // Re-seed to reset name generation
        utilrand.Seed(1)

        //generator := &testGenerator{Token: "ABC"}

        // I thought it was this but we actually generate fake stuff
        // client, _ := newLatticeFakeClient(nil, tc.ClientObjects)

        client := fakelattice.NewSimpleClientset(tc.ClientObjects...)

        for _, reactor := range tc.Reactors {
            client.Fake.PrependReactor(reactor.verb, reactor.resource, reactor.reactor(t))
        }

        // Using controller noresyncfunc creates an include error
        informers := latticeinformers.NewSharedInformerFactory(client, time.Hour)
        endpointInformer := informers.Lattice().V1().Endpoints()
        endpoints := informers.Lattice().V1().Endpoints().Informer().GetStore()

        controller := NewController(serverConfigPath, hostConfigPath, client, endpointInformer)

        if tc.ExistingEndpoints != nil {
            for _, e := range tc.ExistingEndpoints.Items {
                endpoints.Add(e)
            }
        }

        if tc.AddedEndpoint != nil {
            endpoints.Add(tc.AddedEndpoint)
            controller.addEndpoint(tc.AddedEndpoint)
        }
        if tc.UpdatedEndpoint != nil {
            endpoints.Update(tc.UpdatedEndpoint)
            //TODO :: Shouldnt always be nil, logic depends on old
            controller.updateEndpoint(nil, tc.UpdatedEndpoint)
        }
        if tc.DeletedEndpoint != nil {
            endpoints.Delete(tc.UpdatedEndpoint)
            controller.deleteEndpoint(tc.DeletedEndpoint)
        }

        // This is the longest we'll wait for async tests
        timeout := time.Now().Add(30 * time.Second)
        waitedForAdditionalActions := false

        for {
            break
            
            if controller.queue.Len() > 0 {
                // need to supply key
                // controller.syncServiceAccount()
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
    }
}
