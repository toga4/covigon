package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/toga4/covigon/internal/app"
)

func main() {
	run(os.Stdout, os.Args[0], os.Args[1:])
}

func run(w io.Writer, cmd string, args []string) {
	flags, err := app.ParseFlags(w, cmd, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		os.Exit(2)
	}

	file, err := os.Open(flags.Filename)
	if err != nil {
		fmt.Fprintf(w, "error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	if err := app.Run(w, file, flags); err != nil {
		fmt.Fprintf(w, "error: %v\n", err)
		os.Exit(1)
	}
}
