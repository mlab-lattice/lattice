# API

The interface that a lattice provides is exposed via an API.

Relevant type definitions for the API live in [`pkg/api`](../../pkg/api).

## Versioning 

The API is versioned, with `v1` being the only version currently provided.

The types of objects (including errors) returned by the `v1` API are defined in [`pkg/api/v1`](../../pkg/api/v1)

## Client

The `client.Interface` defined in [`pkg/api/client/interface.go`](../../pkg/api/client/interface.go) is a helpful way to get an overview of the capabilities supported by the API.

As can be seen, there is a `V1` interface exposed that shows the capabilities of the `v1` API, which can be found at [`pkg/api/client/v1/interface.go`](../../pkg/api/client/v1/interface.go).

## Server

There are two sets of capabilities required of the API's server (often referred to as the `api-server`):

- process health and transport layer
- implementing the API's interface

The `client.Interface` has a `Health` method, which should return the health of the `api-server` that is connected to. What it means for the `api-server` to be healthy depends on the implementation of the server. Additionally, `client.Interface` does not tie itself to a specific transport layer used to talk to an `api-server`. For that matter in theory it doesn't even tie itself to the idea an `api-server` existing at all, save for the `Health` method.

This is all a long winded way to say that the `api-server` simply servers up an interface that is not specific to its implementation, and the implementation of this interface should be decoupled from the transport layer and process concerns of a given `api-server`.

We use the `backend.Interface` defined in [`pkg/api/server/backend/interface.go`](../../pkg/api/server/backend/interface.go) to accomplish this decoupling. The `backend.Interface` should be implemented by different persistence and control backends. 

When implementing an `api-server`, you should simply accept a `backend.Interface`. Doing so allows you to use the same transport layer implementation with different backend implementations (see immediately below).

This can mostly be summed up with the following: the responsibility of an `api-server` implementation is simply to shepard requests between a client and the supplied `backend.Interface`.

Note that authentication and authorization are still in flux. It seems likely that authorization will be a responsibility of the `backend.Interface`, as it can persist role information etc. It's possible that we add a third interface, something like `authentication.Interface` that could be plugged in to be used with any combination of `api-server` and `backend.Interface`.

## Implementations

Currently, there is one implementation of the `api-server`, and two `backend.Interface` implementations.

### api-server

The only current implementation of the `api-server` provides a RESTful HTTP service. 

The implementation of this server can be found in [`pkg/api/server/rest`](../../pkg/api/server/rest).

The structs used to define `POST` request bodies, as well as information about the different routes exposed by the RESTful `api-server` can be found in [`pkg/api/v1/rest`](../../pkg/api/v1/rest). The response bodies returned by the `api-server` are the structs defined in [`pkg/api/v1`](../../pkg/api/v1).

There is also included an implementation of the `client.Interface` for interacting with the RESTful `api-server`. This can be found in [pkg/api/client/rest](../../pkg/api/client/rest).

### backend.Interface

There are currently two implementations of `backend.Interface`, `kubernetes` and `mock`. However, as the `api-server` is simply consuming the `backend.Interface`, the same server implementation can be used with either backend.

Notably however, there are currently two different binaries produced: one for `kubernetes` ([`cmd/kubernetes/api-server/rest`](../../cmd/kubernetes/api-server/rest)) and one for `mock` ([`cmd/mock/api-server/rest`](../../cmd/mock/api-server/rest)). This was done for simplicity's sake, but as the RESTful `api-server` is written as a library only dependent on `backend.Interface`, one binary could have been used instead.

#### kubernetes

Documentation about how `backend.Interface` is implemented on Kubernetes can be found in the [kubernetes folder](kubernetes).

#### mock

The `mock` backend implements a fairly simple in-memory simulation of a lattice. This should only be used for testing and prototyping.

The `mock` backend implementation can be found at [`pkg/backend/mock/api/server/backend`](../../pkg/backend/mock/api/server/backend).

At a high level, the mock is implemented by two components:
- registry ([`pkg/backend/mock/api/server/backend/registry`](../../pkg/backend/mock/api/server/backend/registry))
  - data structures holding the objects being managed by the system
  - important to note that as these are all in-memory data structures that are not persisted to storage, the `mock` backend will not survive an `api-server` crash
  - also, it is hopefully obvious by the name `mock`, but this backend does not actually deploy systems. there are no containers running anywhere as a result of the mock, so tests that check connectivity to rolled out services should not use the `mock` backend
- controller ([`pkg/backend/mock/api/server/backend/controller`](../../pkg/backend/mock/api/server/backend/controller))
  - control loops that simulate real-world actions acting upon API objects, e.g. scaling a `service`
  