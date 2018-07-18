Creating a lattice named `production` in the us-east-1 region of AWS:

```
$ lattice lattices:create --provider AWS --region us-east-1 --name us-east-1
Lattice production is being created. To watch the status of this lattice run:

    lattice lattices status --lattice production -w
```

Creating a lattice named `production` in the us-east-1 region of AWS and watching the status of the lattice as it is provisioned:

```
$ lattice lattices:create --provider AWS --region us-east-1 --name us-east-1 -w

Lattice production is being created...

    Name    | Provider |  Region   | Address | Status
------------|----------|-----------|---------|---------
 production | AWS      | us-east-1 |         | pending

⠦ Lattice production is pending...
```

This will exit with exit code 0 if the lattice is created successfully. If there is an error, the error will be printed and it will exit with exit code of 1.
