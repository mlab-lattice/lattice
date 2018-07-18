Creating a system named `my-system` with a system definition residing at `https://github.com/mlab-lattice/system__petflix.git`:

```
$ lattice systems:create --name my-system --definition https://github.com/mlab-lattice/system__petflix.git
System my-system created. To rollout a version of this system run:

    lattice systems:deploy --system my-system --version <tag>

```
