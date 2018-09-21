# System Lifecycle

This document covers some of the different states a `system` can be in, and some of the constraints on actions effecting the `system`'s state.

## State Diagram

The state diagram representing a `system`'s lifecycle is presented below.

Note that all of the states (besides `does-not-exist`) can be found in [`pkg/api/v1/system.go`](../../../pkg/api/v1/system.go).

```
        ┌───────────────────────────────────────────────────────────────────────┐       
        │                                                                       │       
        ▼                                                                       │       
┌──────────────┐         ┌─────────────┐         ┌────────────┐         ┌──────────────┐
│              │         │             │         │            │         │              │
│does not exist│────────▶│   pending   │────────▶│   failed   │────────▶│   deleting   │
│              │         │             │         │            │         │              │
└──────────────┘         └─────────────┘         └────────────┘         └──────────────┘
                                │                                               ▲       
                     ┌ ─ ─ ─ ─ ─│─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┐            │       
                                ▼                                               │       
                     │   ┌────────────┐         ┌──────────────┐   │            │       
                         │            │         │              │                │       
                     │   │   stable   │◀───────▶│   updating   │   │            │       
                         │            │         │              │                │       
                     │   └────────────┘         └──────────────┘   │            │       
                                ▲                       ▲                       │       
                     │          │                       │          │            │       
                                ├───────────────────────┤           ────────────┘       
                     │          │                       │          │                    
                                ▼                       ▼                               
                     │   ┌─────────────┐        ┌──────────────┐   │                    
                         │             │        │              │                        
                     │   │   scaling   │◀──────▶│   degraded   │   │                    
                         │             │        │              │                        
                     │   └─────────────┘        └──────────────┘   │                    
                                                                                        
                     └ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┘                    
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

The state of a `system` is determined by the following logic represented in pseudocode:

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

Only two actions can change a `system`'s desired definition: a `deploy` and a `teardown`. The rest of this document assumes familiarity with `deploy`s and `teardown`s, and will discuss their implementation details.

Lattice should allow maximal safe concurrency between deploys and teardowns. To accomplish this, we use [multiple granularity locking](https://en.wikipedia.org/wiki/Multiple_granularity_locking).

### `IntentionLock`
The actual multi-granular lock is implemented in [`pkg/util/sync/intention_lock.go`](../../../pkg/util/sync/intention_lock.go). 

More information can be found in the [documentation](https://godoc.org/github.com/mlab-lattice/lattice/pkg/util/sync#IntentionLock), but the gist is that the lock is a non-blocking lock supporting exclusive and intention-exclusive lock modes.

In english, a user can attempt to lock the lock with either `sync.LockGranularityExclusive` or `sync.LockGranularityIntentionExclusive`. If attempting to lock with `exclusive`, the request will succeed if the lock is not already locked in either `exclusive` or `intention-exclusive` mode. If attempting to lock with `exclusive`, the request will succeed if the lock is not already locked in `exclusive`, but it is okay for the lock to already be locked with `intention-exclusive`.

The `IntentionLock` will not block waiting to acquire the lock if it cannot, instead returning immediately and indicating it was unable to acquire the lock.

If it was able to acquire the lock, it will return an `IntentionLockUnlocker`. Where you would usually use 

```go
defer mutex.Unlock()
```

you should use

```go
defer unlocker.Unlock()
```

The unlocker remembers the granularity you acquired the lock with, and will handle the proper bookkeeping.

### `LifecycleActionManager`

The [`LifecycleActionManager`](../../../pkg/util/sync/lifecycle_action.go) uses a tree of `IntentionLock`s per `system` to manage concurrent `deploy`s and `teardown`s.

The `LifecycleActionManager` will for a given `deploy` (which has an associated `tree.Path`), attempt to acquire an `intention-exclusive` lock at each internal node in the path, and an `exclusive` lock at the leaf node.

So if the `deploy`'s `path` was `/a/b/c/d`, the `LifecycleActionManager` would attempt to lock `/`, `/a`, `/a/b`, and `/a/b/c` at `intention-exclusive`, and `/a/b/c/d` at `exclusive`.

A teardown effects the whole system, so the `LifecycleActionManager` will attempt to acquire an `exclusive` lock at `/`.

All this results in the following semantics:
- a `teardown` can be the only action being run on a system
  - if a single `teardown` is running, all attempts to `deploy` or `teardown` will fail while the `teardown` is active
- a `deploy` can run as long as
  - there is not a `teardown` running
  - there is no `deploy` running whose `path` is an ancestor of the desired `deploy`'s `path`

