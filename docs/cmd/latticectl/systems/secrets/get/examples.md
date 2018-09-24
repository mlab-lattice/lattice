View the secret `MONGODB_URI` on the service `/auth/api`:

```
$ lattice systems secrets get --name /petfix/api:MONGODB_URI

    Path     |    Name     |                        Value
-------------|-------------|------------------------------------------------------
 /petfix/api | MONGODB_URI | mongodb://user:pass@ds012345.mlab.com:12345/database

```
