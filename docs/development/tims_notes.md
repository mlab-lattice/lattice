# Tim's notes

Collection of scratch notes on using Lattice

## latticectl installation

Clone the lattice repo:

```
git clone git@github.com:mlab-lattice/lattice.git
```

Make sure Bazel is installed: https://docs.bazel.build/versions/master/install-os-x.html

To build latticectl, run:

```
bazel build //cmd/latticectl
```

Then the latticectl binary will be available in `bazel-bin/cmd/latticectl/darwin_amd64_stripped/latticectl` (if you're on OSX).
Run this binary directly (you can create an alias to it).

It's also possible to build and run commands in one line:

```
bazel run //cmd/latticectl
```

(Though if you're not actively developing, there's not much point to this since it's slower than building the binary once)

## Installing laas CLI

Install using instructions here:

https://staging.lattice.mlab.com/dashboard/downloads/cli

First thing is run:

```
lattice generate-config
```

You will be prompted for an access key ID and secret which you can generate here: https://staging.lattice.mlab.com/dashboard/me/access-keys

## Creating a lattice

You can create a lattice using the UI or using the laas cli:

```
lattice lattices create --provider AWS --region us-east-1 --name my-lattice
```

Note how latticectl separates commands using `:` whereas laas uses spaces. This is a holdover from experimentation on how to
structure the CLIs, both will use colons in the future.

## latticectl config

Get your lattice URI - this can only be done with the CLI currently (there's an open ticket to add it to the UI):

```
lattice lattices
```

Take the lattice URI of your lattice and add it to your latticectl context:

```
latticectl context:set --lattice <lattice-URI>
```

## Creating a system definition

You can write a system definition and upload it to a GitHub repo. The main file should be `lattice.yml`. Easiest thing
is to make the repo public (public system definition repos are considered kosher).

## Creating a system

Create a system with the UI or with the CLI. Give the system a name and the defintition repo (which is the github URL + `.git`).

E.g.

```
latticectl systems:create --name petflix --definition https://github.com/mlab-lattice/system__petflix.git
```

Then you can deploy a version of the system with:

```
latticectl systems:deploy --version <tag> --system <system-name>
```
