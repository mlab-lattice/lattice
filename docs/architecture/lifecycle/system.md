# System Lifecycle

This document covers some of the different states a `system` can be in, and some of the constraints on actions effecting the `system`'s state.

## State Diagram

The state diagram representing a `system`'s lifecycle is presented below.

Note that all of the states (besides `does-not-exist`) can be found in [`pkg/api/v1/system.go`](../../../pkg/api/v1/system.go).

```
does-not-exist -> pending -> failed --------------------------------
                     |                                              | 
                     |                                              | 
                     |                                              | 
                     |     ------------------------                 |
                     |    |                        |                | 
                     |    |      ----------        |                | 
                     |    |     |          |       |                | 
                     |    |     v          v       |                | 
                      ------> stable -> updating   |                | 
                          |    ^ ^        ^  ^     |                | 
                          |    |  \      /   |     |                | 
                          |    |   \    /    |     |                | 
                          |    |    \  /     |     |                | 
                          |    |     \/      |     |                v
                          |    |     /\      |     | ----------> deleting -> does-not-exist
                          |    |    /  \     |     |
                          |    |   /    \    |     |
                          |    v  v      v   v     |
                          |  scaling -> degraded   |
                          |     ^          ^       |
                          |     |          |       |
                          |      ----------        |
                          |                        |
                           ------------------------
```

## System Initialization

When a `system` is first created, it gets put into the `pending` state. The `backend` then will attempt validate the `system` and create any necessary resources

If the `backend` is unable to do so, the `system` will be moved to the `failed` state.

If the `backend` was able to successfully initialize the `system` it will be moved to the `stable` state.

## System Deletion

At any point, a `system` may be deleted. When it is deleted, it is moved to the `deleting` state, and the `backend` should do whatever it needs to to gracefully tear down the `system `and clean up any `system` specific resources. Once this is complete the system will no longer exist.

## Events Effecting the System State

As mentioned above, the lifecycle of a `system` is bookended by creation/initialization and deletion respectively.

However, as can be seen from the diagram above, between being created and deleted, a `system` can freely transition between any of the states (`stable`, `updating`, `scaling`, `degraded`).

As of right now, the state of a `system` that is neither `deleting` or `pending` is the result of state of all of its `services` (currently `jobs` do not impact the state of a `system`).

The state of a `system` is determined by the following logic represented in psuedocode:

```
if (any service is degraded):
  return degraded
  
if (any service is updating):
  return updating
  
if (any service is scaling):
  return scaling

return stable
```

For information on how `service`s get into different states, please refer to the [service lifecycle documentation](service).

Note that even when a `system`'s desired definition has not been changed, its state can change due to transient changes in the state of its `service`s.

## Lifecycle Action Locking


