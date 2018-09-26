Unset the entire context:

```
$ lattice context get
lattice: my-lattice
system: my-system

$ lattice context unset
lattice:
system:
```

Unset just the system:

```
$ lattice context get
lattice: my-lattice
system: my-system

$ lattice context unset --system
lattice: my-lattice
system:
```
