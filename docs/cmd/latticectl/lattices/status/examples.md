View the information of a lattice named production:

```
$ lattice lattices:status

    Name    | Provider |  Region   |                                 Address                                 |  Status
------------|----------|-----------|-------------------------------------------------------------------------|-----------
 production | AWS      | us-east-1 | http://lattice-ed9e558167-master-1366054109.us-east-1.elb.amazonaws.com | succeeded

```

Watch the information of a lattice:

```
$ lattice lattices:status

    Name    | Provider |  Region   |                                 Address                                 |  Status
------------|----------|-----------|-------------------------------------------------------------------------|-----------
 production | AWS      | us-east-1 | http://lattice-ed9e558167-master-1366054109.us-east-1.elb.amazonaws.com | succeeded

Lattice production is stable.

```

Get the status of a lattice using awk:

```
$ lattice lattices:status --lattice production | awk -F '|' 'FNR == 4 {print $5}' | sed 's/ //g'
succeeded
```

Piping to awk will not work with the `-w, --watch` flag.

Get the status of a lattice by outputting as JSON and piping to jq:

```
$ lattice lattices:status --lattice timl -o json | jq -r '.[0].state'
succeeded
```

Using the `-w, --watch` flag will stream the status every five seconds:

```
$ lattice lattices:status --lattice timl -o json -w | jq -r '.[0].state'
succeeded
succeeded
succeeded

```
