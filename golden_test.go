package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var update = flag.Bool("update", false, "update golden files")

// diff returns a line-by-line diff of two strings
func diff(expected, actual string) string {
	green := color.New(color.FgGreen)
	green.EnableColor()
	red := color.New(color.FgRed)
	red.EnableColor()

	dmp := diffmatchpatch.New()
	a, b, c := dmp.DiffLinesToChars(actual, expected)
	diffs := dmp.DiffMain(a, b, false)
	diffs = dmp.DiffCharsToLines(diffs, c)

	var result []string
	for _, diff := range diffs {
		lines := strings.Split(diff.Text, "\n")
		for _, line := range lines[:len(lines)-1] {
			switch diff.Type {
			case diffmatchpatch.DiffEqual:
				result = append(result, "  "+line)
			case diffmatchpatch.DiffInsert:
				result = append(result, green.Sprint("+ "+line))
			case diffmatchpatch.DiffDelete:
				result = append(result, red.Sprint("- "+line))
			}
		}
	}

	return strings.Join(result, "\n")
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestGolden(t *testing.T) {
	t.Parallel()

	goldenDir, err := filepath.Abs("testdata/golden")
	if err != nil {
		t.Fatalf("Failed to get absolute path of testdata/golden: %v", err)
	}

	moduleDir, err := filepath.Abs("testdata/m")
	if err != nil {
		t.Fatalf("Failed to get absolute path of testdata/m: %v", err)
	}

	if err := os.Chdir(moduleDir); err != nil {
		t.Fatalf("Failed to change directory to %s: %v", moduleDir, err)
	}

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "set",
			args: []string{"coverage_set"},
		},
		{
			name: "count",
			args: []string{"-c", "coverage_count"},
		},
		{
			name: "atomic",
			args: []string{"-c", "coverage_atomic"},
		},
		{
			name: "filter_filename",
			args: []string{"coverage_set", "func1.go"},
		},
		{
			name: "filter_multiple",
			args: []string{"coverage_set", "func1.go", "func2.go"},
		},
		{
			name: "filter_pattern",
			args: []string{"coverage_set", "func*.go"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			goldenFile := filepath.Join(goldenDir, test.name+".golden")

			var buf bytes.Buffer
			run(&buf, "covigon", test.args)
			actual := buf.Bytes()

			if *update {
				// Update golden file
				if err := os.MkdirAll(filepath.Dir(goldenFile), 0755); err != nil {
					t.Fatalf("Failed to create golden file directory: %v", err)
				}
				if err := os.WriteFile(goldenFile, actual, 0644); err != nil {
					t.Fatalf("Failed to update golden file: %v", err)
				}
				t.Logf("Updated golden file: %s", goldenFile)
			} else {
				// Compare with golden file
				expected, err := os.ReadFile(goldenFile)
				if err != nil {
					t.Fatalf("Failed to read golden file: %v", err)
				}

				if !bytes.Equal(expected, actual) {
					// Show diff for better debugging
					t.Errorf("Output mismatch for %s", test.name)
					t.Errorf("Diff:\n%s", diff(string(expected), string(actual)))
				}
			}
		})
	}
}
