Enqueues a teardown of a system. A teardown will unschedule services from nodes and will deprovision the nodes so that the system is left with nothing deployed. A teardown is necessary before deleting a system.

By default, this command will enqueue the teardown of the system then exit if the teardown was successfully enqueued. In order to make the command wait for the teardown to complete, pass the `-w, --watch` flag. This will update you with the status of the system every five seconds and will exit when the teardown has completed (or if there has been an error).

If you enqueue the teardown of a system and decide after that you want to watch the teardown, you can run `lattice system:teardowns:status -w --teardown TEARDOWN_ID` to watch the status of the teardown. The `system:teardowns:status -w` command will exit with status code 1 when the teardown is complete.
