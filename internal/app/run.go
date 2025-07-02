package app

import (
	"io"
)

func Run(writer io.Writer, reader io.Reader, flags *Flags) error {
	profiles, err := ParseProfiles(reader)
	if err != nil {
		return err
	}

	printer := NewPrinter(writer, flags)

	return printer.PrintCoverage(profiles)
}
