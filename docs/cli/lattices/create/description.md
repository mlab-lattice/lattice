This command enqueues the creation of a new lattice in the desired cloud and region. A new lattice consists of a [_Lattice Control Plane_](/user-guide/#lattices). It does not yet have any service nodes, node pools, or systems. The currently supported clouds and regions are:

- `AWS`
    - `us-east-1`

You must provide the cloud provider and region in which you want the lattice to be provisioned. You must also give the lattice a name. Lattice names must be unique within accounts.

Provisioning a new lattice can take several minutes as a VPC, set of VMs, a load balancer, and more must be provisioned to run the lattice. By default, this command will enqueue the creation of the lattice then exit if the creation was successfully enqueued. In order to make the command wait for the creation to complete, pass the `-w, --watch` flag. This will update you with the status of the lattice every five seconds and will exit when the creation has completed (or if there has been an error).

If you enqueue the creation of a lattice and decide after that you want to watch the creation, you can run `lattice lattices:status --lattice LATTICE_NAME -w` to watch the status of the creation. The `lattices:status -w` command will not exit until you send the `SIGTERM` signal.

Once you have provisioned a lattice, you will then be able to run multiple of systems managed by one lattice. In general you will only need to create a small number of lattices to manage all your systems. You may choose to have production, staging, and development lattices... or perhaps a lattice for each team in your organization.
