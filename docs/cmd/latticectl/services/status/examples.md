Show the information for a service named `/petflix/www` in a system named petflix:

```
$ lattice services --system petflix --service /petflix/www

   Service    | State  | Updated | Stale |                                        Addresses                                         | Info
--------------|--------|---------|-------|------------------------------------------------------------------------------------------|------
 /petflix/www | stable |       1 |     0 | 8080: http://tf-lb-20180420221715517300000004-996524393.us-east-2.elb.amazonaws.com:8080 |

```
