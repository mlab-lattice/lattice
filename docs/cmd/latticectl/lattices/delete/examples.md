Deleting a lattice named staging:

```
$ lattice lattices

    Name    | Provider |  Region   |                                 Address                                 |  Status
------------|----------|-----------|-------------------------------------------------------------------------|-----------
 production | AWS      | us-east-1 | http://lattice-ed9e558167-master-1366054109.us-east-1.elb.amazonaws.com | succeeded
 staging    | AWS      | us-east-1 | http://lattice-301657cb4c-master-1469449927.us-east-1.elb.amazonaws.com | succeeded

$ lattice delete --lattice staging

Lattice staging is being deleted. To watch the status of this lattice run:

    lattice lattices status --lattice staging -w
```

Deleting a lattice named staging:

```
$ lattice lattices

    Name    | Provider |  Region   |                                 Address                                 |  Status
------------|----------|-----------|-------------------------------------------------------------------------|-----------
 production | AWS      | us-east-1 | http://lattice-ed9e558167-master-1366054109.us-east-1.elb.amazonaws.com | succeeded
 staging    | AWS      | us-east-1 | http://lattice-301657cb4c-master-1469449927.us-east-1.elb.amazonaws.com | succeeded

$ lattice delete --lattice production

Lattice production is being deleted. To watch the status of this lattice run:

    lattice lattices status --lattice production -w
```
