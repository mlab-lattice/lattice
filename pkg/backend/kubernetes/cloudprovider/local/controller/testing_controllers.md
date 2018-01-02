# Testing Controllers

This resource provides an overview of testing Kubernetes controllers.

Client mock
lattice fake client
actions
reactors
CRUD

Test logic -
    Update loop - single threaded.

Reactor - arbritrary function call on each object update.

Action - chronological list of CRUD events occuring on the client
    note these might be made to the cache rather than the client, so expected actions might not appear.