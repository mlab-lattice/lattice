Listing lattices in the default table format:

```sh
$ lattice lattices

    Name    | Provider |  Region   |                                 Address                                 |  Status
------------|----------|-----------|-------------------------------------------------------------------------|-----------
 production | AWS      | us-east-1 | http://lattice-ed9e558167-master-1366054109.us-east-1.elb.amazonaws.com | succeeded
 staging    | AWS      | us-east-1 | http://lattice-301657cb4c-master-1469449927.us-east-1.elb.amazonaws.com | succeeded
```

Listing lattices as JSON and piping to the [`jq`](https://stedolan.github.io/jq/) tool:

```
$ lattice lattices -o json | jq

[
  {
    "id": "ed9e558167a1ba8e39dadaaf85839320",
    "name": "production",
    "provider": "AWS",
    "region": "us-east-1",
    "state": "succeeded",
    "address": "http://lattice-ed9e558167-master-1366054109.us-east-1.elb.amazonaws.com",
    "accountId": "a3ecd193cea9156687d3df744be0e88c",
    "createdAt": "2018-04-11T22:20:25.142Z"
  },
  {
    "id": "301657cb4c6b32691ee1430c8b1174a6",
    "name": "staging",
    "provider": "AWS",
    "region": "us-east-1",
    "state": "succeeded",
    "address": "http://lattice-301657cb4c-master-1469449927.us-east-1.elb.amazonaws.com",
    "accountId": "a3ecd193cea9156687d3df744be0e88c",
    "createdAt": "2018-04-16T19:07:58.689Z"
  }
]
```
