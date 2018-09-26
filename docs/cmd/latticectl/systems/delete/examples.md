Teardown and then delete a system:

```
$ lattice systems:teardown --system petflix

Tearing down system petflix. Teardown ID: 15c77128-e7a0-4dbe-838b-60ebfab0a358

To watch teardown, run:

    lattice system:teardowns:status -w --teardown 15c77128-e7a0-4dbe-838b-60ebfab0a358

$ lattice systems:delete --system petflix
System petflix deleted.
```

Attempting to delete a system without first tearing down the system:

```
TODO: Doesn't work
```
