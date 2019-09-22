// Package help is the help subcommand of the influxd command.
package help

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Command displays help for command-line sub-commands.
type Command struct {
	Stdout io.Writer
}

// NewCommand returns a new instance of Command.
func NewCommand() *Command {
	return &Command{
		Stdout: os.Stdout,
	}
}

// Run executes the command.
func (cmd *Command) Run(args ...string) error {
	fmt.Fprintln(cmd.Stdout, strings.TrimSpace(usage))
	return nil
}

const usage = `
Tools for performing operations on 1.x shards.

Usage: influx_shard command [arguments]

The commands are:

    import               imports a shard into a 1.x OSS database
    help                 display this help message

Use "influx_tools command -help" for more information about a command.
`
