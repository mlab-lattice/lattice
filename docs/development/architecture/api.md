# API

The interface that a lattice provides is exposed via an API.

Relevant type definitions for the API live in [`pkg/api`](../../../pkg/api).

## Versioning 

The API is versioned, with `v1` being the only version currently provided.

The types of objects (including errors) returned by the `v1` API are defined in [`pkg/api/v1`](../../../pkg/api/v1)

## Client

The `client.Interface` defined in [`pkg/api/client/interface.go`](../../../pkg/api/client/interface.go) is a helpful way to get an overview of the capabilities supported by the API.

As can be seen, there is a `V1` interface exposed that shows the capabilities of the `v1` API, which can be found at [`pkg/api/client/v1/interface.go`](../../../pkg/api/client/v1/interface.go).

## Server

There are a few sets of capabilities required of the API's server (often referred to as the `api-server`):

- authentication
- authorization
- persisting and interacting with API objects
- transport

With the exception of the transport layer, these capabilities are represented by separate decoupled interfaces. 

### Authentication

Authentication at its core is about resolving a set of credentials to an identity.

The types of supported authentication schemes are defined as interfaces in [`pkg/api/server/authentication/authenticator`](../../../pkg/api/server/authentication/authenticator).

As mentioned before, these interfaces are decoupled from e.g. the transport layer. For example, the `authenticator.Token` interface authenticates a user based on a presented token. However, it does not care if the token was supplied in a bearer token HTTP header vs an RPC framework's metadata.

### Authorization

Authorization has not yet been fully fleshed out, but when it is it will follow a similar pattern of decoupling.

The interface(s) will likely end up living in `pkg/api/server/authorization/authorizer`.

A possible interface could look like:

```go
type Interface interface {
	Authorized(user authentication.UserInfo, action authorization.Action) (bool, error)
}
```

### Persisting API objects

The main point of the `api-server` is to be able to interact with the objects stored by the API.

Similar to authentication and authorization, this is also abstracted behind an interface.

This interface can be found in [`pkg/api/server/backend`](../../../pkg/api/server/backend).

Note that it is not necessarily the responsibility of the `backend.Interface` to act upon the objects (e.g. deploy a system when a `deploy` is created). There should be an orchestrator watching the API objects and acting upon them, potentially from a different process external to the `api-server`.

That said, currently the [`v1.Interface`](../../../pkg/api/server/backend/v1/interface.go) does require some knowledge of the underlying orchestrator, as it needs to be able to stream logs from running workloads.


### Transport layer

In some sense the transport layer is the implementation of the `api-server`. The transport layer is often the glue that ties together the above interfaces.

For example, an HTTP server may find a token in an `Authorization` header, use an `authenticator.Token` interface implementation to retrieve user info, and pass that user info to an `authenticator.Interface`.

If either of these fail, it will translate the failure into a proper HTTP status code (403) and return a response to the user.

A general outline of the flow of an `api-server` implementation is included below:

```
              │                      │                   │             
┌─────────┐        ┌─────────────┐        ┌──────────┐       ┌───────┐ 
│transport│   │    │authenticator│   │    │authorizer│   │   │backend│ 
└─────────┘        └─────────────┘        └──────────┘       └───────┘ 
─ ─ ─ ─ ─ ─ ─ ┼ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│─ ─ ─ ─ ─ ─ ─ ─ ─ ─│─ ─ ─ ─ ─ ─ ─
┌─────────┐       ┌──────────────┐       ┌───────────┐       ┌────────┐
│         │   │   │              │   │   │           │   │   │        │
│ request │──────▶│ authenticate │──ok──▶│ authorize │──ok──▶│ action │
│         │   │   │              │   │   │           │   │   │        │
└─────────┘       └──────────────┘       └───────────┘       └────────┘
              │           │          │         │         │   │        │
┌──────────┐◀─────────────┴───────failure──────┴─────────────┘        │
│          │  │                      │                   │            │
│ response │                                                          │
│          │  │                      │                   │            │
└──────────┘◀───────────────────────ok────────────────────────────────┘
              │                      │                   │             
                                                                       
```

## Implementations

Currently, there is one implementation of the `api-server`, a WIP `authenticator.Interface`, and two `backend.Interface` implementations.

### api-server

The only current implementation of the `api-server` provides a RESTful HTTP service. 

The implementation of this server can be found in [`pkg/api/server/rest`](../../../pkg/api/server/rest).

The structs used to define request bodies, as well as information about the different routes exposed by the RESTful `api-server` can be found in [`pkg/api/v1/rest`](../../../pkg/api/v1/rest). The response bodies returned by the `api-server` are JSON encodings of the structs defined in [`pkg/api/v1`](../../../pkg/api/v1).

There is also included an implementation of the `client.Interface` for interacting with the RESTful `api-server`. This can be found in [pkg/api/client/rest](../../../pkg/api/client/rest).

### backend.Interface

There are currently two implementations of `backend.Interface`, `kubernetes` and `mock`. However, as the `api-server` is simply consuming the `backend.Interface`, the same server implementation can be used with either backend.

Notably however, there are currently two different binaries produced: one for `kubernetes` ([`cmd/kubernetes/api-server/rest`](../../../cmd/kubernetes/api-server/rest)) and one for `mock` ([`cmd/mock/api-server/rest`](../../../cmd/mock/api-server/rest)). This was done for simplicity's sake, but as the RESTful `api-server` is written as a library only dependent on `backend.Interface`, one binary could have been used instead.

#### kubernetes

Documentation about how `backend.Interface` is implemented on Kubernetes can be found in the kubernetes [documentation](kubernetes).

#### mock

The `mock` backend implements a fairly simple in-memory simulation of a lattice. This should only be used for testing and prototyping.

The `mock` backend implementation can be found at [`pkg/backend/mock/api/server/backend`](../../../pkg/backend/mock/api/server/backend).

At a high level, the mock is implemented by two components:
- registry ([`pkg/backend/mock/api/server/backend/registry`](../../../pkg/backend/mock/api/server/backend/registry))
  - ata structures holding the objects being managed by the system
  - important to note that as these are all in-memory data structures that are not persisted to storage, the `mock` backend will not survive an `api-server` crash
  - also, it is hopefully obvious by the name `mock`, but this backend does not actually deploy systems
    - there are no containers running anywhere as a result of the mock, so tests that check connectivity to rolled out services should not use the `mock` backend
    - similarly, for simplicity's sake the `mock` backend for the most part does not simulate random failure. while there are delays in its actions to simulate reality, all operations succeed
- controller ([`pkg/backend/mock/api/server/backend/controller`](../../../pkg/backend/mock/api/server/backend/controller))
  - control loops that simulate real-world actions acting upon API objects, e.g. scaling a `service`
  