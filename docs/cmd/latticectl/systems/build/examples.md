Build version `1.0.0` of a system:

```
$ lattice systems:build --version 1.0.0
Building version 1.0.0, Build ID: 30745090-73e4-46de-b180-c76959422fda

To view the status of the build, run:

    latticectl system:builds:status --build 30745090-73e4-46de-b180-c76959422fda [--watch]

```

Watch the build process:

```
lattice systems:build --version 1.0.0 -w

Build ID: f43dcf2b-e164-4863-a6cc-8ffff846cb58

    Component     |   State   |         Info
------------------|-----------|----------------------
 /petflix/api:api | succeeded | pushing docker image
 /petflix/www:www | succeeded | pushing docker image

✓ 1.0.0 built successfully! You can now deploy this build using:

    lattice systems:deploy f43dcf2b-e164-4863-a6cc-8ffff846cb58
```
