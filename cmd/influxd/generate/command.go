package generate

import (
	"github.com/influxdata/influxdb/cmd/influxd/generate/data"
	"github.com/spf13/cobra"
)

// NewCommand creates the new command.
func NewCommand() *cobra.Command {
	base := &cobra.Command{
		Use:   "generate",
		Short: "Commands for generating data directly on the filesystem",
	}

	// List of available sub-commands
	// If a new sub-command is created, it must be added here
	subCommands := []*cobra.Command{
		data.Command,
	}

	base.AddCommand(subCommands...)

	return base
}
