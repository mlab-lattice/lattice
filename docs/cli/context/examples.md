Without a context you must pass the `--lattice` and `--system` flags to the `lattice systems deploy` command:

```
$ lattice context get
lattice:
system:

$ lattice systems deploy --version v1.0.0 --lattice my-lattice --system my-system

Deploying version v1.0.0 for system my-system. Deploy ID: babdea1e-fd8e-47e1-a80e-4cd2b65a8822

To watch deploy, run:

    lattice system:deploys:status -w --deploy babdea1e-fd8e-47e1-a80e-4cd2b65a8822
```

Instead, you can set the context and the `--lattice` and `--system` flags are automatically filled in by the context.

```
$ lattice context set --lattice my-lattice --system my-system
lattice: my-lattice
system: my-system

$ lattice systems deploy --version v2.0.0

Deploying version v1.0.0 for system my-system. Deploy ID: faf615f5-8a8a-4efa-89e5-ee55210dcf84

To watch deploy, run:

    lattice system:deploys:status -w --deploy faf615f5-8a8a-4efa-89e5-ee55210dcf84
```

!!! caution
    Make sure to check the current context by running `lattice context get` before running commands. Otherwise you may inadvertently make a change to a lattice or system you did not want to change.
