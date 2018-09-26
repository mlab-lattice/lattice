Get the status of

```
lattice systems:builds:status --build 91696a41-0d60-4e2a-ae37-dd54d2b8114c

    Component     |   State   |         Info
------------------|-----------|----------------------
 /petflix/api:api | succeeded | pushing docker image
 /petflix/www:www | succeeded | pushing docker image
```

To watch the status of a build after it has been enqueued, use the `-w, --watch` flag:

```
lattice systems:builds:status --build 91696a41-0d60-4e2a-ae37-dd54d2b8114c -w

    Component     |   State   |         Info
------------------|-----------|----------------------
 /petflix/api:api | succeeded | pushing docker image
 /petflix/www:www | succeeded | pushing docker image

✓ 1.0.0 built successfully! You can now deploy this build using:

    lattice systems:deploy 91696a41-0d60-4e2a-ae37-dd54d2b8114c
```
