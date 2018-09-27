Deploying a system and watching the deploy:

```
$ lattice systems:deploy --version 1.0.0 -w

    Component     |   State   |         Info
------------------|-----------|----------------------
 /petflix/api:api | succeeded | pushing docker image
 /petflix/www:www | succeeded | pushing docker image

✓ 1.0.0 built successfully! Now deploying...

   Service    | State  | Updated | Stale |                                        Addresses                                         | Info
--------------|--------|---------|-------|------------------------------------------------------------------------------------------|------
 /petflix/api | stable |       1 |     0 |                                                                                          |
 /petflix/www | stable |       1 |     0 | 8080: http://tf-lb-20180424220446719300000003-421405457.us-east-2.elb.amazonaws.com:8080 |

✓ Rollout for system petflix has succeeded.
```
