Enqueues building a version of a system. Specify a git tag of the system definition repo. Lattice will pull the definition from that tag and will build a system adhering to that definition. This does not deploy the system. It builds the docker images necessary to deploy the system.

Use this command to build systems in advance of deploying so that when you deploy it takes much quicker.

If you want to build and deploy at the same time, you can just use the `lattice systems:deploy` command to do both at the same time. The `lattice systems:deploy` command will be quicker if you have previously built all the images for that version.
