package main

import (
	"fmt"
	"github.com/mlab-lattice/system/pkg/cli/command"
)

var (
	bar  string
	buzz string
	qux  string
)

var cmd = &command.Command{
	Name: "foo",
	Flags: command.Flags{
		&command.StringFlag{
			Name:   "bar",
			Target: &bar,
		},
		&command.EmbeddedFlag{
			Name:     "embed",
			Required: true,
			Flags: command.Flags{
				&command.StringFlag{
					Name:   "buzz",
					Target: &buzz,
				},
				&command.StringFlag{
					Name:     "qux",
					Required: true,
					Target:   &qux,
				},
			},
		},
	},
	Run: func(args []string) {
		fmt.Printf("bar: %v\n", bar)
		fmt.Printf("buzz: %v\n", buzz)
		fmt.Printf("qux: %v\n", qux)
	},
}

func main() {
	cmd.Execute()
}
