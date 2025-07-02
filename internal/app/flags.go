package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

type Flags struct {
	ShowCount  bool
	ForceColor bool
	Filename   string
	Filters    []string
}

func ParseFlags(w io.Writer, cmd string, args []string) (*Flags, error) {
	var flags Flags

	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	fs.SetOutput(w)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), `Usage: %s [OPTIONS...] FILE [FILTER...]

Options:
  -c, --count  Show execution count for covered lines
  --color      Force color output
  -h, --help   Show this help message

`, cmd)
	}

	// short flags
	fs.BoolVar(&flags.ShowCount, "c", false, "")

	// long flags
	fs.BoolVar(&flags.ShowCount, "count", false, "")
	fs.BoolVar(&flags.ForceColor, "color", false, "")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if fs.NArg() < 1 || fs.Arg(0) == "" {
		msg := "missing file argument"
		fmt.Fprintln(w, msg)
		fs.Usage()
		return nil, errors.New(msg)
	}

	flags.Filename = fs.Arg(0)

	// FILTER arguments
	if fs.NArg() > 1 {
		flags.Filters = fs.Args()[1:]
	}

	return &flags, nil
}
