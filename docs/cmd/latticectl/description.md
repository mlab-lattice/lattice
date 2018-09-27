The Lattice Command Line Interface is a tool for managing your lattices via the shell.

### Installation

You can download the Lattice CLI from the [download page](https://alpha.lattice.mlab.com/dashboard/downloads/cli). You can also install it on MacOS using [Homebrew](https://brew.sh/):

```
$ brew install mlab-lattice/lattice/lattice-cli
```

Once installed, you can view a list of commands by running `lattice help`.

To use the Lattice CLI, first generate an _Access Key ID_ and _Access Key Secret_ on the [Access Keys page](https://alpha.lattice.mlab.com/dashboard/me/access-keys). Run `lattice lattices` to generate a config file. You will be prompted to enter the _Access Key ID_ and _Access Key Secret_ you just generated.

You will now be able to run Lattice CLI commands. Try making a new lattice by running:

```
$ lattice lattices:create --provider AWS --region us-east-1 --name my-first-lattice
```

You can then see the status of this lattice by running:

```
$ lattice lattices:status --lattice my-first-lattice
```

To learn more about how to get started, see the [Getting Started Guide](/getting-started) and the [Official Lattice Tutorial](/tutorial).

### Help

You can view a list of available commands by running `lattice help`.

To get detailed instructions on specific commands, pass the name of the command to `lattice help`. For example, to learn more about the usage of `systems:create`, run `lattice help systems:create`:

```
Create a new system

Usage:
  lattice systems:create [flags]

Flags:
      --config string       the config file to use
      --definition string   the repository containing the system definition
  -h, --help                help for systems:create
      --lattice string      the lattice to act on
      --name string         the name of the system to create
```

You can also append the `-h, --help` flag to get information on commands.

### Output Formats

You can set the output format of lattice commands using the `-o, --output` flag. The options are `table` and `json`. `table` is the default output format. It is a human readable table that can also easily be used with grep and AWK. For example:

```
$ lattice lattices

    Name    | Provider |  Region   |                                 Address                                 |  Status
------------|----------|-----------|-------------------------------------------------------------------------|-----------
 production | AWS      | us-east-1 | http://lattice-ed9e558167-master-1366054109.us-east-1.elb.amazonaws.com | succeeded
 staging    | AWS      | us-east-1 | http://lattice-301657cb4c-master-1469449927.us-east-1.elb.amazonaws.com | succeeded
```

The JSON output will output a JSON object. The JSON objects are easily machine-readable. If you are piping output into another program, we recommend using the JSON output. JSON output can easily be formatted and manipulated using the `jq` tool. For example:

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

### Output Highlighting

Unique identifiers of resources are colored cyan. Note that this may display as a different color depending on your terminal settings.

### The Watch Flag

Many commands in the lattice CLI are asynchronous. When you deploy a system using `lattice systems:deploy`, lattice enqueues the deploy and then the command returns. To watch the progress of commands, use the `-w, --watch` flag. This will update the status of a command every five seconds. When a command reaches a terminating state (e.g. a success or a failure), the command will exit with the appropriate exit code. Watching operations such as `lattice systems:build` and `lattice systems:deploy` will exit when they reach a terminating state. Watching the status of resources, such as `lattice systems` and `lattice services:status` will not exit until lattice receives a SIGTERM. See the documentation for commands to learn which commands can terminate when watched.

### Config Files

Lattice will generate a config file from your _Access Key ID_ and _Access Key Secret_. By default this config file is located at `~/.config/lattice/config.json`. You can instruct lattice to use a different config file by passing the path of the config file with the `--config` option.

### Context

Rather than including the `--lattice` and `--system` flags with every command, you can set the lattice and system you are currently working on in the context. This is useful if you are issuing several commands in a row on the same lattice or system. You can also set a context by running `lattice context:set`. This will create a context file located at `~/.config/lattice/context.json`. You can always override the current context by manually setting the `--lattice` and/or `--system` flags.

### Usage with AWK

The default table output is easily parsed by awk. For example, the output for `lattice lattices:status` is:

```
    Name    | Provider |  Region   |                                 Address                                 |  Status
------------|----------|-----------|-------------------------------------------------------------------------|-----------
 production | AWS      | us-east-1 | http://lattice-ed9e558167-master-1366054109.us-east-1.elb.amazonaws.com | succeeded

Lattice production is stable.
```

This can be piped to awk using the separator `'|'`. To pick out the status column, you can use this awk command:

```
$ lattice lattices:status --lattice production | awk -F '|' 'FNR == 4 {print $5}' | sed 's/ //g'
succeeded
```

Here we also pipe to sed to trim any whitespace.

The `-w, --watch` flag with table output is not usable with AWK. Use JSON output with JQ instead.

### Usage with JQ

The JSON output with `-o, --output json` can be piped to the `jq` tool. JQ is a powerful tool for parsing and manipulating JSON. For example, piping the above output to JQ:

```
$ lattice lattices:status --lattice timl -o json | jq -r '.[0].state'
succeeded
```

Using the `-w, --watch` flag will stream the status every five seconds. The pipe to JQ will remain open and JQ will output the result every 5 seconds.

```
$ lattice lattices:status --lattice timl -o json -w | jq -r '.[0].state'
succeeded
succeeded
succeeded
```
