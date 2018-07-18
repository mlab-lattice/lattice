List all the secrets in a system. A secret is made up of a path, a name, and a value. The path is the path of the service where the secret will be available. The name is the name of the environment variable that the secret will be set as on that service. And the value is the value which that environment variable will be set to.

So if you have a secret in lattice `/auth/api:MONGODB_URI` set to `mongodb://user:pass@ds012345.mlab.com:12345/database`, that means the service at `/auth/api` will have an environment variable `MONGODB_URI` set to the desired value.
