# covigon

A Go coverage visualizer that displays coverage information directly in your source code with color-coded output.

## Installation

```bash
go install github.com/toga4/covigon@latest
```

## Usage

```bash
# Generate coverage profile
go test -coverprofile=coverage.out

# Display coverage
covigon coverage.out

# Show execution counts
covigon -c coverage.out

# Force color output (useful for CI/pipes)
covigon --color coverage.out

# Filter by filename
covigon coverage.out main.go

# Filter by pattern
covigon coverage.out "*.go"

# Multiple filters (OR condition)
covigon coverage.out main.go util.go
```

## Example Output

```
[main.go]
   1 +  func main() {
   2 +      fmt.Println("Hello, World!")
   3 -      unreachableCode()
   4    }
```

- `+` indicates covered lines (green)
- `-` indicates uncovered lines (red)
- Numbers show execution count when using `-c` flag

## Options

- `-c, --count`: Show execution count for covered lines
- `--color`: Force color output
- `-h, --help`: Show help message

## Filtering

You can filter files by specifying one or more filter patterns after the coverage file:

```bash
covigon coverage.out [FILTER...]
```

Filters are applied using Go's `filepath.Match` function and work against:
- File paths relative to the current working directory
- Filename only (basename)

Multiple filters use OR logic - files matching any filter will be included.

## License

MIT