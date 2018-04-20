# Building

## Dependencies

Only [Bazel](https://bazel.build) is required to build lattice binaries. For building lattice docker images, please see the [docker documentation](docker-images.md).

## Make

To build lattice, after installing Bazel you should simply be able to run `make`.

Lattice's [Makefile](../../Makefile) contains some other convenient targets for common operations.


## Platforms

Lattice can currently be built for darwin and linux.

`make` will build lattice for the platform it is being invoked on.

To build for a specific platform, you can run `make build.darwin` or `make build.linux`. `make build.all` will build lattice for all supported platforms.

The built binaries can be found under the `bazel-bin` directory. You can read more about the directory layouts produced by bazel [here](https://docs.bazel.build/versions/master/output_directories.html).

## Bazel

Lattice uses Bazel heavily. As such, it's recommended that you familiarize yourself with bazel via its [docs](https://docs.bazel.build), at the very least [Concepts and Terminology](https://docs.bazel.build/versions/master/build-ref.html#concepts-and-terminology).

To get you started though, we'll give a brief overview of Bazel and the capabilities being used by Lattice.

### Overview

There's a decent amount to take in here. I'd recommend reading [Commands](#commands) but pushing through any questions you may have, then reading [Go](#go) and following along with the example, referencing back to the information in [Commands](#commands) as needed.

#### Commands

Bazel has two main commands to be aware of:

```shell
$ bazel build <target>
```

and

```shell
$ bazel run <target>
```

`bazel build <target>` will build all of the target's dependencies and the target.

`bazel run <target>` will build then run the target if it is executable.

#### Workspace

The root of the project contains a `WORKSPACE` file.

This file contains a list of all of the external dependencies needed to build targets.

#### Targets

Bazel defines targets for the previously mentioned commands through `BUILD` files containing rules. These rules are defined in a language called [Skylark](https://github.com/google/skylark), which is almost a subset of Python.

A `BUILD` file may then look like this:

```python
my_rule(
    name = "foo",
    *args,
    **kwargs,
)
```

This defines a rule `foo`. The target produced by by this rule depends on the build file's location relative to the `WORKSPACE` file.

For example, say the project looked like this, with the each `BUILD` file looking like above:

```shell
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

```python
git_repository(
    name = "my_other_repo",
    remote = "https://github.com/acme/other_rules.git",
    tag = "0.0.2",
)
```

If this repository contained the following structure:

```shell
$ tree
.
├── BUILD
├── WORKSPACE
└── bar
    ├── BUILD
    └── custom_rules.bzl
```

and `custom_rules.bzl` contained the definition for the `my_rule` example above, in your `BUILD` file, you could load and use this rule:

```python
load("@my_other_repo//bar/custom_rules.bzl", "my_rule")

my_rule(
    name = "foo",
    *args,
    **kwargs,
)
```

As mentioned earlier, the `WORKSPACE` file containes external dependency declarations. An external dependency declared in the `WORKSPACE` file can be referenced in `BUILD` files via the name `@<dependency_name>`.

So if your `WORKSPACE` file contained a rule:

```python
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
Lets say we wanted to have a repository at `github.com/acme/example` that contained a library `pkg/spinner` that exported a function `Spin` that we want to have a little spinner on the CLI and a binary that uses the function.

Let's first just make the shell of the application before adding the spinning functionality. We could create that like so:

```shell
$ mkdir -p cmd pkg/spinner

$ cat > cmd/main.go << EOF
package main

import "github.com/acme/example/pkg/spinner"

func main() {
	spinner.Spin()
}
EOF

$ cat > pkg/spinner/spinner.go << EOF
package spinner

import "fmt"

func Spin() {
    fmt.Println("I can't spin :(")
}
EOF

$ tree
.
├── cmd
│   └── main.go
└── pkg
    └── spinner
        └── spinner.go
```

In order to build this with Bazel we'll first have to include the `rules_go` rules in our `WORKSPACE`:

```shell
$ cat > WORKSPACE << EOF
http_archive(
    name = "io_bazel_rules_go",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.10.1/rules_go-0.10.1.tar.gz",
    sha256 = "4b14d8dd31c6dbaf3ff871adcd03f28c3274e42abc855cb8fb4d01233c0154dc",
)
load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains")
go_rules_dependencies()
go_register_toolchains()
EOF
```

Then we have to create two additional `BUILD` files: `pkg/hello/BUILD` and `cmd/BUILD`.

```shell
$ cat > pkg/spinner/BUILD << EOF
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "spinner_library",
    srcs = ["spinner.go"],
    importpath = "github.com/acme/example/pkg/spinner",
    visibility = ["//visibility:public"],
)
EOF

$ cat > cmd/BUILD << EOF
load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "cmd_library",
    srcs = ["main.go"],
    importpath = "github.com/acme/example/cmd",
    visibility = ["//visibility:private"],
    deps = [
        "//pkg/spinner:spinner_library",
    ],
)

go_binary(
    name = "cmd",
    embed = [":cmd_library"],
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

This will create three targets: `//pkg/spinner:spinner_library`, `//cmd:cmd_library`, and `//cmd:cmd` (aka `//cmd`).

Note that we had to explicitly list all our non standard library dependencies (`deps` in `//cmd:cmd_library`)

Now, if you run `bazel run //cmd`, you would see something like the following:

```shell
 bazel run //cmd
INFO: Analysed target //cmd:cmd (0 packages loaded).
INFO: Found 1 target...
Target //cmd:cmd up-to-date:
  bazel-bin/cmd/darwin_amd64_stripped/cmd
INFO: Elapsed time: 0.491s, Critical Path: 0.03s
INFO: Build completed successfully, 1 total action

INFO: Running command line: bazel-bin/cmd/darwin_amd64_stripped/cmd
I can't spin :(
```

#### Automatic BUILD file generation with gazelle
That's all great, but it's a lot of manual work. Luckily, there's a different bazel repository that does all of this work for us, called `gazelle`.

If we include `gazelle`:

```shell
$ rm -f WORKSPACE && cat > WORKSPACE << EOF
http_archive(
    name = "io_bazel_rules_go",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.10.1/rules_go-0.10.1.tar.gz",
    sha256 = "4b14d8dd31c6dbaf3ff871adcd03f28c3274e42abc855cb8fb4d01233c0154dc",
)
http_archive(
    name = "bazel_gazelle",
    url = "https://github.com/bazelbuild/bazel-gazelle/releases/download/0.10.0/bazel-gazelle-0.10.0.tar.gz",
    sha256 = "6228d9618ab9536892aa69082c063207c91e777e51bd3c5544c9c060cafe1bd8",
)

load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains")
go_rules_dependencies()
go_register_toolchains()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
gazelle_dependencies()
EOF
```

and instead of creating the two specific files `pkg/spinner/BUILD` and `cmd/BUILD` we instead wrote `BUILD` as:

```shell
$ rm -f pkg/hello/BUILD cmd/BUILD && cat > BUILD << EOF
load("@bazel_gazelle//:def.bzl", "gazelle")

gazelle(
    name = "gazelle",
    prefix = "github.com/acme/example",
)
EOF
```

You could then run `bazel run //:gazelle`, and `gazelle` would walk the filesystem, generating the proper `BUILD` files (it calls them `BUILD.bazel`) with the `go_library` and `go_binary` rules, with correct dependencies and all.

Then you could run `bazel run //cmd` again just as before, since `gazelle` generated the `go_binary` rule for us at `//cmd:cmd`. Let's try it:

```shell
$ tree
  .
  ├── BUILD
  ├── WORKSPACE
  ├── cmd
  │   └── main.go
  └── pkg
      └── spinner
          └── spinner.go


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
│   ├── BUILD
│   └── main.go
└── pkg
    └── spinner
        ├── BUILD
        └── spinner.go

$ bazel run //cmd
INFO: Analysed target //cmd:cmd (0 packages loaded).
INFO: Found 1 target...
Target //cmd:cmd up-to-date:
bazel-bin/cmd/darwin_amd64_stripped/cmd
INFO: Elapsed time: 0.475s, Critical Path: 0.01s
INFO: Build completed successfully, 1 total action

INFO: Running command line: bazel-bin/cmd/darwin_amd64_stripped/cmd
I can't spin :(
```

#### External dependencies

Now, say we wanted to add spinning to our library using the repository [https://github.com/briandowns/spinner](https://github.com/briandowns/spinner).

```shell
$ rm -f pkg/spinner/spinner.go && cat > pkg/spinner/spinner.go << EOF
package spinner

import (
    "github.com/briandowns/spinner"
    "time"
)

func Spin() {
    s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)  // Build our new spinner
    s.Start()                                                    // Start the spinner
    time.Sleep(4 * time.Second)                                  // Run for some time to simulate work
    s.Stop()
}
EOF
```

Now if we try to run `bazel run //cmd`, we'll get a failure:


```shell
$ bazel run //cmdd
INFO: Analysed target //cmd:cmd (0 packages loaded).
INFO: Found 1 target...
ERROR: /Users/kevinrosendahl/go/src/github.com/acme/example-bazel/pkg/spinner/BUILD:3:1: GoCompile pkg/spinner/darwin_amd64_stripped/spinner_library~/github.com/acme/example/pkg/spinner.a failed (Exit 1)
2018/03/01 12:46:27 missing strict dependencies:
      pkg/spinner/spinner.go: import of github.com/briandowns/spinner, which is not a direct dependency
Target //cmd:cmd failed to build
Use --verbose_failures to see the command lines of failed build steps.
INFO: Elapsed time: 0.605s, Critical Path: 0.19s
FAILED: Build did NOT complete successfully
ERROR: Build failed. Not running target
```

We need to run `gazelle` again so that it can update `/pkg/spinner/BUILD` with the new dependency we introduced.

Run `gazelle` and try again:

```shell
$ bazel run //:gazelle
INFO: Found 1 target...
Target //:gazelle up-to-date:
  bazel-bin/gazelle_script.bash
  bazel-bin/gazelle
INFO: Elapsed time: 0.767s, Critical Path: 0.03s

INFO: Running command line: bazel-bin/gazelle

$  bazel run //cmd
ERROR: /Users/kevinrosendahl/go/src/github.com/acme/example-bazel/pkg/spinner/BUILD:3:1: no such package '@com_github_briandowns_spinner//': The repository could not be resolved and referenced by '//pkg/spinner:spinner_library'
ERROR: Analysis of target '//cmd:cmd' failed; build aborted: Loading failed
INFO: Elapsed time: 0.367s
FAILED: Build did NOT complete successfully (2 packages loaded)
ERROR: Build failed. Not running target
```

Now Bazel complains that it could not find `@com_github_briandowns_spinner`.

Let's look at the `pkg/spinner/BUILD` file `gazelle` generated for us:

```shell
$ cat pkg/spinner/BUILD
load("@io_bazel_rules_go//go:def.bzl", "go_library")
  
go_library(
    name = "spinner_library",
    srcs = ["spinner.go"],
    importpath = "github.com/acme/example/pkg/spinner",
    visibility = ["//visibility:public"],
    deps = ["@com_github_briandowns_spinner//:go_default_library"],
)
```

As we can see in the `deps` argument, the `go_library` rule is trying to reference `@com_github_briandowns_spinner`. As we said before, targets in the form of `@<name>` are defined in the `WORKSPACE` file.

What's happened here is that `gazelle` recognized we have an external dependency on `github.com/briandowns/spinner`, and is expecting us to have included that dependency in our `WORKSPACE` file.

Let's include it and try again:

```shell
$ rm -f WORKSPACE && cat > WORKSPACE <<EOF
http_archive(
    name = "io_bazel_rules_go",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.10.1/rules_go-0.10.1.tar.gz",
    sha256 = "4b14d8dd31c6dbaf3ff871adcd03f28c3274e42abc855cb8fb4d01233c0154dc",
)
http_archive(
    name = "bazel_gazelle",
    url = "https://github.com/bazelbuild/bazel-gazelle/releases/download/0.10.0/bazel-gazelle-0.10.0.tar.gz",
    sha256 = "6228d9618ab9536892aa69082c063207c91e777e51bd3c5544c9c060cafe1bd8",
)

load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains", "go_repository")
go_rules_dependencies()
go_register_toolchains()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
gazelle_dependencies()

go_repository(
    name = "com_github_briandowns_spinner",
    tag = "1.0",
    importpath = "github.com/briandowns/spinner",
)
EOF

$ bazel run //cmd
ERROR: /private/var/tmp/_bazel_kevinrosendahl/a86530172cf0a051f6446529fb951250/external/com_github_briandowns_spinner/BUILD.bazel:3:1: no such package '@com_github_fatih_color//': The repository could not be resolved and referenced by '@com_github_briandowns_spinner//:go_default_library'
ERROR: Analysis of target '//cmd:cmd' failed; build aborted: Loading failed
INFO: Elapsed time: 9.647s
FAILED: Build did NOT complete successfully (11 packages loaded)
ERROR: Build failed. Not running target 
```

More complaints. Bazel is now complaining there's no target `@com_github_fatih_color`. But what's this? We never included anything like that.

That's because Bazel requires you to include the transitive closure of all external dependencies, i.e. even include the dependencies of your dependencies.

Turns out that `github.com/briandowns/spinner` in turn depends on `github.com/fatih/color`.

Let's put those in `WORKSPACE` and try one more time:

```shell
$ rm -f WORKSPACE && cat > WORKSPACE << EOF
http_archive(
    name = "io_bazel_rules_go",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.10.1/rules_go-0.10.1.tar.gz",
    sha256 = "4b14d8dd31c6dbaf3ff871adcd03f28c3274e42abc855cb8fb4d01233c0154dc",
)
http_archive(
    name = "bazel_gazelle",
    url = "https://github.com/bazelbuild/bazel-gazelle/releases/download/0.10.0/bazel-gazelle-0.10.0.tar.gz",
    sha256 = "6228d9618ab9536892aa69082c063207c91e777e51bd3c5544c9c060cafe1bd8",
)

load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains", "go_repository")
go_rules_dependencies()
go_register_toolchains()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
gazelle_dependencies()

go_repository(
    name = "com_github_briandowns_spinner",
    tag = "1.0",
    importpath = "github.com/briandowns/spinner",
)

go_repository(
    name = "com_github_fatih_color",
    tag = "v1.5.0",
    importpath = "github.com/fatih/color",
)
EOF

$  bazel run //cmd
INFO: Analysed target //cmd:cmd (0 packages loaded).
INFO: Found 1 target...
Target //cmd:cmd up-to-date:
bazel-bin/cmd/darwin_amd64_stripped/cmd
INFO: Elapsed time: 0.317s, Critical Path: 0.01s
INFO: Build completed successfully, 1 total action

INFO: Running command line: bazel-bin/cmd/darwin_amd64_stripped/cmd
\

```

#### Macros

There are a number of macros in the [bazel](../../bazel) directory that make organizing bazel targets easier.

Take a look at the [WORKSPACE](../../WORKSPACE) file or the [docker BUILD file](../../docker/BUILD) for examples of using them.


## Best practices and enforcement

In general, you should default to simply running `make`. This will run `gazelle` to make sure you have up-to-date `BUILD` files as well as ensure that everything can compile. Once you build lattice once, subsequent builds are incremental.

The pre-commit hook for the repository (installable by running `make git.install-hooks`) will check that bazel `BUILD` files are up to date and that all code is formatted.

