Deletes a system from a lattice. This command will not run if a version of the system is currently deployed. You must explicitly teardown the system before deleting. A teardown will deprovision the services in an system. Once there are no running services in a system, the system can then be deleted.

If you want to teardown a deploy, but intend to deploy a new version at a later date, just teardown the system without delting.
