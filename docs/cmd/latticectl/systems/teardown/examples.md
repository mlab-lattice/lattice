Enqueue a teardown of a system:

```
$ lattice systems:teardown

Tearing down system petflix. Teardown ID: 477718ec-d5ab-4a6b-809f-a36fc168b29b

To watch teardown, run:

    lattice system:teardowns:status -w --teardown 477718ec-d5ab-4a6b-809f-a36fc168b29b
```

Enqueue a teardown of a system and watch the teardown:

```
lattice systems:teardown -w --system my-system

Tearing down system my-system. Teardown ID: 9850fd71-7d01-41ea-9693-c5d1c701f846


 Service | State | Updated | Stale | Addresses | Info
---------|-------|---------|-------|-----------|------

System my-system is stable.
```

TODO: Needs newline character at the end, also should this be watching services? This has to do with the system lifecycle - there should be a torn down state.
