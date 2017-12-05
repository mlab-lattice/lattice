# Building

## Bazel

Lattice uses Bazel heavily. As such, it's recommended that you familiarize yourself with bazel via its [docs](https://docs.bazel.build), at the very least [Concepts and Terminology](https://docs.bazel.build/versions/master/build-ref.html#concepts-and-terminology).

Lattice currently requires Bazel version 0.7.0 (exactly 0.7.0, not >= 0.7.0).

To get you started though, we'll give a brief overview of Bazel and the capabilities being used by Lattice.

### Overview

There's a decent amount to take in here. I'd recommend reading [Commands](#commands) but pushing through any questions you may have, then reading [Go](#go) and following along with the example, referencing back to the information in [Commands](#commands) as needed.

#### Commands

Bazel has two main commands to be aware of:

```
bazel build <target>
```

and

```
bazel run <target>
```

`bazel build <target>` will build all of the target's dependencies and the target.

`bazel run <target>` will build then run the target if it is executable.

#### Workspace

The root of the project contains a `WORKSPACE` file.

This file contains a list of all of the external dependencies needed to build targets.

#### Targets

Bazel defines targets for the previously mentioned commands through `BUILD` files containing rules. These rules are defined in a language called [Skylark](https://github.com/google/skylark), which is almost a subset of Python.

A `BUILD` file may then look like this:

```
my_rule(
    name = "foo",
    *args,
    **kwargs,
)
```

This defines a rule `foo`. The target produced by by this rule depends on the build file's location relative to the `WORKSPACE` file.

For example, say the project looked like this, with the each `BUILD` file looking like above:

```
$ tree
.
├── BUILD
├── WORKSPACE
├── a
│   ├── BUILD
│   └── b
│       └── BUILD
└── c
    └── BUILD
```

A target's `label` is `//` then the path from the workspace root to the `BUILD` file containing the desired target, a `:`, then the name of the rule being targeted.

So if we're trying to target the rule in the file `/a/b/BUILD`, the `label` would be `//a/b:foo`.

The rule in `/BUILD` would be targeted with the label `//:foo`.

To build `//a/b:foo` simply run `bazel build //a/b:foo`. Finally, if a rule has the same name as the directory it's in, you can leave off the `:rule_name`. So if we had `//a/b:b` we could reference this as `//a/b`.

There are a number of built-in rules that you can always use in `BUILD` files, the [Bazel docs](https://docs.bazel.build) contain information about the built-in rules.

You can load custom (not built in) rules via the built in `load(label, rules...)` function.

So you could include another git repository that contains a `.bzl` file containing rule definitions and use those rule definitions by including this in your `WORKSPACE` file (using the [git_repository WORKSPACE rule](https://docs.bazel.build/versions/master/be/workspace.html#git_repository)):

```
git_repository(
    name = "my_other_repo",
    remote = "https://github.com/acme/other_rules.git",
    tag = "0.0.2",
)
```

If this repository contained the following structure:

```
$ tree
.
├── BUILD
├── WORKSPACE
└── bar
    ├── BUILD
    └── custom_rules.bzl
```

and `custom_rules.bzl` contained the definition for the `my_rule` example above, in your `BUILD` file, you could load and use this rule:

```
load("@my_other_repo//bar/custom_rules.bzl", "my_rule")

my_rule(
    name = "foo",
    *args,
    **kwargs,
)
```

As mentioned earlier, the `WORKSPACE` file containes external dependency declarations. An external dependency declared in the `WORKSPACE` file can be referenced in `BUILD` files via the name `@<dependency_name>`.

So if your `WORKSPACE` file contained a rule:

```
external_dependency(
    name = "other_repo",
    *args,
    **kwargs,
)
```

you could reference this in a `BUILD` file via `@other_repo`.

### Go

#### Rules
Bazel has a number of helpful rules for Go in the [rules_go repository](https://github.com/bazelbuild/rules_go).

The rules we can use are `go_library` and `go_binary`.


#### Manual BUILD files
Lets say we wanted to have a repository at `github.com/acme/hello` that contained a library that exported a function `HelloWorld` that prints out "hello world" and a binary that uses the function.

We could create that like so:

```
$ mkdir -p cmd pkg/hello
$ cat > pkg/hello/hello.go << EOF
package hello

import "fmt"

func HelloWorld() {
	fmt.Println("hello world")
}
EOF

$ cat > cmd/main.go << EOF
package main

import "github.com/acme/hello/pkg/hello"

func main() {
	hello.HelloWorld()
}
EOF

$ tree
.
├── cmd
│   └── main.go
└── pkg
    └── hello
        └── hello.go
```

In order to build this with Bazel we'll first have to include the `rules_go` rules in our `WORKSPACE`:

```
$ cat > WORKSPACE << EOF
git_repository(
    name = "io_bazel_rules_go",
    remote = "https://github.com/bazelbuild/rules_go.git",
    commit = "44b3bdf7d3645cbf0cfd786c5f105d0af4cf49ca",
)
load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains")
go_rules_dependencies()
go_register_toolchains()
EOF
```

Then, in order to let Bazel know that our repo is `github.com/acme/hello`, we need to add the following to the file `BUILD`:

```
$ cat > BUILD << EOF
load("@io_bazel_rules_go//go:def.bzl", "go_prefix")
go_prefix("github.com/acme/hello")
EOF
```

Then we have to create two additional `BUILD` files: `pkg/hello/BUILD` and `cmd/BUILD`.

```
$ cat > pkg/hello/BUILD << EOF
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "hello_library",
    srcs = ["hello.go"],
    importpath = "github.com/acme/hello/pkg/hello",
    visibility = ["//visibility:public"],
)
EOF

$ cat > cmd/BUILD << EOF
load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "cmd_library",
    srcs = ["main.go"],
    importpath = "github.com/acme/hello/cmd",
    visibility = ["//visibility:private"],
    deps = [
        "//pkg/hello:hello_library",
    ],
)

go_binary(
    name = "cmd",
    embed = [":cmd_library"],
    importpath = "github.com/acme/hello/cmd",
    visibility = ["//visibility:public"],
)
EOF

$ tree
.
├── BUILD
├── WORKSPACE
├── cmd
│   ├── BUILD
│   └── main.go
└── pkg
  └── hello
      ├── BUILD
      └── hello.go
```

This will create three targets: `//pkg/hello:hello_library`, `//cmd:cmd_library`, and `//cmd:cmd` (aka `//cmd`).

Note that we had to explicitly list all our non standard library dependencies (`deps` in `//cmd:cmd_library`)

Now, if you run `bazel run //cmd`, you would see something like the following:

```
$ bazel run //cmd
INFO: Found 1 target...
Target //cmd:cmd up-to-date:
  bazel-bin/cmd/cmd
INFO: Elapsed time: 0.506s, Critical Path: 0.01s

INFO: Running command line: bazel-bin/cmd/cmd
hello world
```

#### Automatic BUILD file generation with gazelle
That's all great, but it's a lot of manual work. Luckily, `rules_go` includes a binary that does all of this work for us, called `gazelle`.

If instead of creating the two specific files `pkg/hello/BUILD` and `cmd/BUILD`, if we instead wrote `BUILD` as:

```
$ rm -f pkg/hello/BUILD cmd/BUILD BUILD && cat > BUILD << EOF
load("@io_bazel_rules_go//go:def.bzl", "go_prefix", "gazelle")
go_prefix("github.com/acme/hello")
gazelle(name = "gazelle")
EOF
```

You could then run `bazel run //:gazelle`, and `gazelle` would walk the filesystem, generating the proper `BUILD` files (it calls them `BUILD.bazel`) with the `go_library` and `go_binary` rules, with correct dependencies and all.

Then you could run `bazel run //cmd` again just as before, since `gazelle` generated the `go_binary` rule for us at `//cmd:cmd`. Let's try it:

```
$ tree
.
├── BUILD
├── WORKSPACE
├── cmd
│   └── main.go
└── pkg
    └── hello
        └── hello.go

$ bazel run //:gazelle
INFO: Found 1 target...
Target //:gazelle up-to-date:
  bazel-bin/gazelle_script.bash
  bazel-bin/gazelle
INFO: Elapsed time: 0.511s, Critical Path: 0.01s

INFO: Running command line: bazel-bin/gazelle

$ tree
.
├── BUILD
├── WORKSPACE
├── cmd
│   ├── BUILD.bazel
│   └── main.go
└── pkg
    └── hello
        ├── BUILD.bazel
        └── hello.go

$ bazel run //cmd
INFO: Found 1 target...
Target //cmd:cmd up-to-date:
  bazel-bin/cmd/cmd
INFO: Elapsed time: 0.997s, Critical Path: 0.52s

INFO: Running command line: bazel-bin/cmd/cmd
hello world
```

#### External dependencies

Now, say we wanted to add color to our library using the repository [https://github.com/fatih/color](https://github.com/fatih/color).

```
$ rm -f pkg/hello/hello.go && cat > pkg/hello/hello.go << EOF
package hello

import "github.com/fatih/color"

func HelloWorld() {
        color.NoColor = false
        color.Green("hello world")
}
EOF
```

Now if we try to run `bazel run //cmd`, we'll get a failure:


```
$ bazel run //cmd
INFO: Found 1 target...
ERROR: /Users/kevinrosendahl/tmp/bazel/pkg/hello/BUILD.bazel:3:1: GoCompile pkg/hello/darwin_amd64_stripped/go_default_library~/buzz/pkg/hello.a failed (Exit 1).
2017/12/01 21:38:58 missing strict dependencies:
        pkg/hello/hello.go: import of github.com/fatih/color, which is not a direct dependency
Target //cmd:cmd failed to build
Use --verbose_failures to see the command lines of failed build steps.
INFO: Elapsed time: 0.401s, Critical Path: 0.11s
ERROR: Build failed. Not running target.
```

We need to run `gazelle` again so that it can update `/pkg/hello/BUILD.bazel` with the new dependency we introduced.

Run `gazelle` and try again:

```
$ bazel run //:gazelle
INFO: Found 1 target...
Target //:gazelle up-to-date:
  bazel-bin/gazelle_script.bash
  bazel-bin/gazelle
INFO: Elapsed time: 0.767s, Critical Path: 0.03s

INFO: Running command line: bazel-bin/gazelle

$ bazel run //cmd
ERROR: /Users/kevinrosendahl/tmp/bazel/pkg/hello/BUILD.bazel:3:1: no such package '@com_github_fatih_color//': The repository could not be resolved and referenced by '//pkg/hello:go_default_library'.
ERROR: Analysis of target '//cmd:cmd' failed; build aborted: Loading failed.
INFO: Elapsed time: 0.184s
ERROR: Build failed. Not running target.
```

Now Bazel complains that it could not find `@com_github_fatih_color`.

Let's look at the `pkg/hello/BUILD.bazel` file `gazelle` generated for us:

```
$ cat pkg/hello/BUILD.bazel
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["hello.go"],
    importpath = "github.com/acme/hello/pkg/hello",
    visibility = ["//visibility:public"],
    deps = ["@com_github_fatih_color//:go_default_library"],
)
```

As we can see in the `deps` argument, the `go_library` rule is trying to reference `@com_github_fatih_color`. As we said before, targets in the form of `@<name>` are defined in the `WORKSPACE` file.

What's happened here is that `gazelle` recognized we have an external dependency on `github.com/fatih/color`, and is expecting us to have included that dependency in our `WORKSPACE` file.

Let's include it and try again:

```
$ rm -f WORKSPACE && cat > WORKSPACE <<EOF
git_repository(
    name = "io_bazel_rules_go",
    remote = "https://github.com/bazelbuild/rules_go.git",
    commit = "44b3bdf7d3645cbf0cfd786c5f105d0af4cf49ca",
)
load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains", "go_repository")
go_rules_dependencies()
go_register_toolchains()

go_repository(
    name = "com_github_fatih_color",
    tag = "v1.5.0",
    importpath = "github.com/fatih/color",
)
EOF

$ bazel run //cmd
ERROR: /private/var/tmp/_bazel_kevinrosendahl/15450d47838b10e224ba0c221c7915f7/external/com_github_fatih_color/BUILD.bazel:3:1: no such package '@com_github_mattn_go_isatty//': The repository could not be resolved and referenced by '@com_github_fatih_color//:go_default_library'.
ERROR: Analysis of target '//cmd:cmd' failed; build aborted: no such package '@com_github_mattn_go_isatty//': The repository could not be resolved.
INFO: Elapsed time: 4.713s
ERROR: Build failed. Not running target.
```

More complaints. Bazel is now complaining there's no target `@com_github_mattn_go_isatty`. But what's this? We never included anything like that.

That's because Bazel requires you to include the transitive closure of all external dependencies, i.e. even include the dependencies of your dependencies.

Turns out that `github.com/fatih/color` in turn depends on `github.com/mattn/go-colorable` and `github.com/mattn/go-isatty`.

Let's put those in `WORKSPACE` and try one more time:

```
$ rm -f WORKSPACE && cat > WORKSPACE << EOF
git_repository(
    name = "io_bazel_rules_go",
    remote = "https://github.com/bazelbuild/rules_go.git",
    commit = "44b3bdf7d3645cbf0cfd786c5f105d0af4cf49ca",
)
load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains", "go_repository")
go_rules_dependencies()
go_register_toolchains()

go_repository(
    name = "com_github_fatih_color",
    tag = "v1.5.0",
    importpath = "github.com/fatih/color",
)

go_repository(
    name = "com_github_mattn_go_colorable",
    commit = "5411d3eea5978e6cdc258b30de592b60df6aba96",
    importpath = "github.com/mattn/go-colorable",
)

go_repository(
    name = "com_github_mattn_go_isatty",
    commit = "57fdcb988a5c543893cc61bce354a6e24ab70022",
    importpath = "github.com/mattn/go-isatty",
)
EOF

$ bazel run //cmd
 INFO: Found 1 target...
 Target //cmd:cmd up-to-date:
   bazel-bin/cmd/cmd
 INFO: Elapsed time: 6.232s, Critical Path: 0.99s
 
 INFO: Running command line: bazel-bin/cmd/cmd
 hello world
```

I now realize using colorized output isn't a great example seeing as that can't be put in a `.md`, but if you've been following along, you'd now see "hello world" in green.

#### Macros

At the time of writing this, the transitive closure of Lattice's external Go dependencies includes over 60 different repositories. As one could imagine, the `WORKSPACE` file became very bloated and hard to visually parse.

To combat this, custom macro rules were written which when called generate the `go_repository` rules. These rules can be found in the [bazel](../../bazel) directory.

So instead of calling `go_repository` 60 times, we can include the custom macro that generates those rules, and simply call it:

```
load(":bazel/dependencies.bzl", "go_dependencies")
go_dependencies()
```

### Docker

Bazel also has rules building and pushing Docker images in the [rules_docker repository](https://github.com/bazelbuild/rules_docker).

We won't go into as much depth here. By now you should be able to read the documentation available at that repo and get a general idea of what the rules provided are.

Notably, we generate a `container_push` rule in `docker/BUILD` in lattice for each docker image we can build. For example, one docker image we want to build is for `kubernetes-bootstrap-lattice`.

Via the custom rule generator defined in `docker/lattice_targets.bzl`, we generate a `container_push` rule for `kubernetes-bootstrap-lattice` whose `label` is `//docker:push-kubernetes-bootstrap-lattice`.

When you run `bazel run //docker:push-kubernetes-bootstrap-lattice`, it will build `//cmd/kubernetes/bootstrap-lattice`, put it in a docker image, and push it to `gcr.io/lattice-dev/kubernetes-bootstrap-lattice`.

#### IMPORTANT NOTE

Bazel `rules_go`'s support for cross compilation is not yet mature enough to support cross compiling the `go_binary` to Linux to put into the `container_image`. As such `//docker:push-*` should only be run from a Linux box (or container).

See the [docker-hack section](#docker-hack) below for more information about overcoming this.

## Make

Lattice has a Makefile that contains a few convenience targets.

- `make build`
  - Runs gazelle and builds all bazel targets, including docker images (but will not tag, export, or push them) (`bazel build //...`)
- `make clean`
  - Clears the Bazel artifact cache (`bazel clean`)
- `make test`
  - Runs all tests (`bazel test //...`)
- `make gazelle`
  - Runs gazelle (`bazel run //:gazelle`)
- `make docker-push-image IMAGE=<target-name>`
  - Pushes the target production and debug versions of the target to `gcr.io/lattice-dev` (`bazel run //docker:push-$IMAGE && bazel run //docker:push-debug-$IMAGE`)
  - Will fail if not being run on Linux
- `make docker-push-all-images`
  - Pushes all docker images to `gcr.io/lattice-dev`
  - Will fail if not being run on Linux

### docker-hack

As stated above, as of now the cross compile story is not good enough to run `//docker:*` on a OSX box.

This is solved by running a docker container which has Bazel installed and mounting the lattice repo into the container, and running Bazel commands from inside.

There are three options to use this:

- `make docker-hack-enter-build-shell`
  - Drops you into a shell insider the build container. From there you can correctly build/run the `//docker:*` targets.
- `make docker-hack-push-image IMAGE=<target-name>`
  - Enters the build container and runs `make docker-push-image IMAGE=<target-name>`
- `make docker-hack-push-all-images`
  - Enters the build container and runs `make docker-push-all-images`

### Best practices and enforcement

In general, you should always default to simply running `make build`. This will run `gazelle` to make sure you have up-to-date `BUILD` files as well as ensure that everything can compile.

The `pre-push` git hook for the repository will in fact ensure that `make build` and `make test` both succeed, so it behooves you to usually be running `make build` so that your push will be fast. This may or may not be revisited in the future.
