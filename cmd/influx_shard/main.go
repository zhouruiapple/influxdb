// The influx_shard command performs operations with 1.x shards.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/influxdata/influxdb/cmd"
	"github.com/influxdata/influxdb/cmd/influx_shard/help"
	"github.com/influxdata/influxdb/cmd/influx_shard/importer"
	_ "github.com/influxdata/influxdb/tsdb/engine"
)

func main() {
	m := NewMain()
	if err := m.Run(os.Args[1:]...); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Main represents the program execution.
type Main struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// NewMain returns a new instance of Main.
func NewMain() *Main {
	return &Main{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

// Run determines and runs the command specified by the CLI args.
func (m *Main) Run(args ...string) error {
	name, args := cmd.ParseCommandName(args)

	// Extract name from args.
	switch name {
	case "", "help":
		if err := help.NewCommand().Run(args...); err != nil {
			return fmt.Errorf("help failed: %s", err)
		}
	case "import":
		c := importer.NewCommand()
		if err := c.Run(args); err != nil {
			return fmt.Errorf("import failed: %s", err)
		}
	default:
		return fmt.Errorf(`unknown command "%s"`+"\n"+`Run 'influx_shard help' for usage`+"\n\n", name)
	}

	return nil
}
