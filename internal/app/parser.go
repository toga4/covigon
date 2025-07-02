package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os/exec"
	"path"
	"path/filepath"
	"slices"

	"golang.org/x/tools/cover"
)

type Profile struct {
	*cover.Profile

	ResolvedPath string
}

func ParseProfiles(r io.Reader) ([]*Profile, error) {
	coverProfiles, err := cover.ParseProfilesFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse coverage profile: %w", err)
	}

	// Get unique directories
	dirSet := make(map[string]struct{})
	for _, profile := range coverProfiles {
		dirSet[path.Dir(profile.FileName)] = struct{}{}
	}
	dirs := slices.Collect(maps.Keys(dirSet))

	output, err := runGoList(dirs)
	if err != nil {
		return nil, fmt.Errorf("failed to run go list: %w", err)
	}

	type Pkg struct {
		ImportPath string `json:"ImportPath"`
		Dir        string `json:"Dir"`
		Error      *struct {
			Err string `json:"Err"`
		} `json:"Error,omitempty"`
	}

	// Parse multiple JSON objects using json.Decoder
	pkgs := make(map[string]*Pkg)
	decoder := json.NewDecoder(bytes.NewReader(output))
	for {
		var pkg Pkg
		if err := decoder.Decode(&pkg); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode go list output: %w", err)
		}
		// Map ImportPath to Pkg
		pkgs[pkg.ImportPath] = &pkg
	}

	// Resolve file paths
	profiles := make([]*Profile, 0, len(coverProfiles))
	for _, profile := range coverProfiles {
		dir := path.Dir(profile.FileName)
		pkg, exists := pkgs[dir]
		if !exists {
			return nil, fmt.Errorf("package not found for directory: %s", dir)
		}

		// Check for error
		if pkg.Error != nil {
			return nil, fmt.Errorf("package error: %s", pkg.Error.Err)
		}

		// Combine Dir with filename basename
		profiles = append(profiles, &Profile{
			Profile:      profile,
			ResolvedPath: filepath.Join(pkg.Dir, filepath.Base(profile.FileName)),
		})
	}

	return profiles, nil
}

func runGoList(dirs []string) ([]byte, error) {
	args := append([]string{"list", "-e", "-json"}, dirs...)
	cmd := exec.Command("go", args...)
	return cmd.Output()
}
