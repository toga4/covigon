package app

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/tools/cover"
)

var (
	colorRed   = color.New(color.FgRed)
	colorGreen = color.New(color.FgGreen)
)

type Printer struct {
	writer    io.Writer
	showCount bool
	filters   []string
}

func NewPrinter(w io.Writer, flags *Flags) *Printer {
	// Force color output if Color option is set
	// Otherwise, let fatih/color decide based on output destination
	if flags.ForceColor {
		color.NoColor = false
	}

	return &Printer{
		writer:    w,
		showCount: flags.ShowCount,
		filters:   flags.Filters,
	}
}

func (p *Printer) PrintCoverage(profiles []*Profile) error {
	for _, profile := range profiles {
		if p.shouldIncludeProfile(profile) {
			if err := p.printProfile(profile); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Printer) shouldIncludeProfile(profile *Profile) bool {
	// If no filters are specified, include all profiles
	if len(p.filters) == 0 {
		return true
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		// If we can't get the current directory, fall back to including everything
		return true
	}

	// Convert ResolvedPath to relative path from current working directory
	relPath, err := filepath.Rel(cwd, profile.ResolvedPath)
	if err != nil {
		// If conversion fails, use absolute path
		relPath = profile.ResolvedPath
	}

	// Get filename only
	filename := filepath.Base(profile.ResolvedPath)

	// Check if any filter matches (OR logic)
	for _, filter := range p.filters {
		// Check against relative path
		if matched, _ := filepath.Match(filter, relPath); matched {
			return true
		}
		// Check against filename only
		if matched, _ := filepath.Match(filter, filename); matched {
			return true
		}
	}

	return false
}

func (p *Printer) printProfile(profile *Profile) error {
	sourceContent, err := os.ReadFile(profile.ResolvedPath)
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %w", profile.ResolvedPath, err)
	}

	// Get boundaries for processing
	boundaries := profile.Boundaries(sourceContent)

	// Calculate total line count
	totalLines := bytes.Count(sourceContent, []byte("\n")) + 1
	lineNumWidth := max(len(strconv.Itoa(totalLines)), 4)

	// Calculate max count for alignment
	maxCountDigits := 0
	maxCount := 0
	for _, block := range profile.Blocks {
		if block.Count > maxCount {
			maxCount = block.Count
		}
	}
	if maxCount > 0 {
		maxCountDigits = len(strconv.Itoa(maxCount))
	}

	// Generate line headers
	lineHeaders := p.generateLineHeaders(sourceContent, boundaries, lineNumWidth, maxCountDigits)

	// Generate lines with coverage info
	lines := p.generateLines(sourceContent, boundaries)

	// Render all lines
	fmt.Fprint(p.writer, "\n[")
	fmt.Fprint(p.writer, profile.FileName)
	fmt.Fprintln(p.writer, "]")
	for i := range lineHeaders {
		fmt.Fprint(p.writer, lineHeaders[i])
		fmt.Fprint(p.writer, lines[i])
		fmt.Fprintln(p.writer)
	}

	return nil
}

func (p *Printer) generateLineHeaders(src []byte, boundaries []cover.Boundary, lineNumWidth int, maxCountDigits int) []string {
	srcLines := bytes.Split(src, []byte("\n"))
	lineHeaders := make([]string, 0, len(srcLines))

	offsetHasCovered := false
	offsetHasUncovered := false
	offsetMaxCount := 0

	boundaryIndex := 0
	endOffset := 0
	for i, srcLine := range srcLines {
		endOffset = endOffset + len(srcLine) + 1 // +1 for newline

		lineHasCovered := offsetHasCovered
		lineHasUncovered := offsetHasUncovered
		lineMaxCount := offsetMaxCount

		for boundaryIndex < len(boundaries) && boundaries[boundaryIndex].Offset <= endOffset {
			b := boundaries[boundaryIndex]
			if b.Start {
				offsetHasCovered = b.Count > 0
				offsetHasUncovered = b.Count == 0
				offsetMaxCount = max(offsetMaxCount, b.Count)
				lineHasCovered = lineHasCovered || offsetHasCovered
				lineHasUncovered = lineHasUncovered || offsetHasUncovered
				lineMaxCount = max(lineMaxCount, offsetMaxCount)
			} else {
				offsetHasCovered = false
				offsetHasUncovered = false
				offsetMaxCount = 0
			}
			boundaryIndex++
		}

		lineHeaders = append(lineHeaders, p.generateLineHeader(lineNumWidth, i+1, lineHasCovered, lineHasUncovered, maxCountDigits, lineMaxCount))
	}

	return lineHeaders
}

func (p *Printer) generateLineHeader(lineNumWidth int, lineNum int, hasCovered bool, hasUncovered bool, maxCountDigits int, maxCount int) string {
	var buf bytes.Buffer

	// Print line number and space
	fmt.Fprintf(&buf, "%*d ", lineNumWidth, lineNum)

	// Print covered indicator
	if hasCovered {
		buf.WriteString(colorGreen.Sprint("+"))
	} else {
		buf.WriteString(" ")
	}

	// Print uncovered indicator
	if hasUncovered {
		buf.WriteString(colorRed.Sprint("-"))
	} else {
		buf.WriteString(" ")
	}

	// Print space
	buf.WriteString(" ")

	// Print count
	if p.showCount && maxCountDigits > 0 {
		if maxCount > 0 {
			buf.WriteString(colorGreen.Sprint(fmt.Sprintf("%*d", maxCountDigits, maxCount)))
		} else {
			buf.WriteString(strings.Repeat(" ", maxCountDigits))
		}

		// Print space
		buf.WriteString(" ")
	}

	return buf.String()
}

func (p *Printer) generateLines(src []byte, boundaries []cover.Boundary) []string {
	if len(boundaries) == 0 {
		return strings.Split(string(src), "\n")
	}

	boundary := boundaries[0]

	var buf bytes.Buffer

	// Write the part of the source code before the first boundary
	buf.Write(src[:boundary.Offset])

	for _, nextBoundary := range boundaries[1:] {
		block := src[boundary.Offset:nextBoundary.Offset]
		if boundary.Start {
			// write the block with color based on the count
			col := colorRed
			if boundary.Count > 0 {
				col = colorGreen
			}
			for line := range bytes.Lines(block) {
				buf.WriteString(col.Sprint(string(line)))
			}
		} else {
			// write the block without color
			buf.Write(block)
		}

		// Move to the next boundary
		boundary = nextBoundary
	}

	// Write the part of the source code after the last boundary
	buf.Write(src[boundary.Offset:])

	// replace tabs with 4 spaces for column consistency
	indented := strings.ReplaceAll(buf.String(), "\t", "    ")

	return strings.Split(indented, "\n")
}
