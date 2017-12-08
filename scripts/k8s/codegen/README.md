# Codegen

We use a code generator to create the clientset, informers, and listers for the for our custom resources. The code generated will be the same structure as [client-go](https://godoc.org/k8s.io/client-go).

Eventually (waiting on [1315](https://github.com/kubernetes/community/pull/1315)) we'll want to run this through Bazel, similar to how we run gazelle.

For now however, we have to run it by hand:

```
$ go get -d k8s.io/code-generator
$ cd ${GOPATH}/src/k8s.io/code-generator
$ git checkout <current-k8s-components-version>
$ ./generate-groups.sh all \
                       github.com/mlab-lattice/system/pkg/kubernetes/customresource/generated \
                       github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis \
                       lattice:v1 \
                       --go-header-file ~/go/src/github.com/mlab-lattice/system/scripts/k8s/codegen/go-header.txt
```

This will create the above mentioned libraries by parsing `pkg/kubernetes/customresource/apis/lattice/v1` and put the output into `pkg/kubernetes/customresource/generated`.

It'll also use `go-header.txt` as the header instead of the default license it puts in.
