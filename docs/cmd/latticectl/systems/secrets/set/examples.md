Set the secret `MONGODB_URI` on the service `/auth/api` to the value `mongodb://user:pass@ds012345.mlab.com:12345/database`:

```
$ lattice systems:secrets:set --secret /auth/api:MONGODB_URI --value mongodb://user:pass@ds012345.mlab.com:12345/database
```
TODO: No output
