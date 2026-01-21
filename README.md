# Xeger

**Xeger** (pronounced *zeh-ger*, like "zeh Germans") is a Go library and CLI tool that generates example strings matching a given regular expression. Think of it as the "reverse" of regex: instead of testing if a string matches a pattern, Xeger produces valid strings from patterns. Handy for reverse-testing your regex patterns or generating test data.

## Quick Examples

Generate realistic test data in seconds:

```bash
# Email addresses with popular domains
$ xeger -n 5 "seth\+[a-z0-9]{4,8}@(gmail|yahoo|example)\.com"
seth+kdho@yahoo.com
seth+h8hk@example.com
seth+vwno@yahoo.com
seth+yju8@yahoo.com
seth+3zuh@example.com

# Phone numbers
$ xeger -n 3 "\+1-[0-9]{3}-[0-9]{3}-[0-9]{4}"
+1-774-045-9094
+1-829-087-5394
+1-954-009-0556

# Names with middle initial
$ xeger -n 3 "[A-Z][a-z]+ [A-Z]\. [A-Z][a-z]+"
Df Y. Bndx
Lww B. Gf
Iwq V. Lha

# Random alphanumeric IDs
$ xeger -n 5 "[A-Z0-9]{8}"
A3B7C9D1
X2Y4Z6W8
M5N7P9Q1
R3S5T7U9
V1W3X5Y7
```

Perfect for testing, fuzzing, documentation examples, and generating synthetic data!

## Features

- **Go Package**: Importable library for programmatic use
- **CLI Tool**: Command-line executable for quick string generation
- **Full Regex Support**: Handles literals, character classes, alternation, quantifiers, and more
- **Configurable**: Control max length, repetition limits, and bias toward shorter strings
- **Well Tested**: Comprehensive test suite with roundtrip property validation

## Installation

```bash
go get github.com/sethgrid/xeger
```

Or clone the repository:

```bash
git clone https://github.com/sethgrid/xeger.git
cd xeger
go build ./cmd/xeger
```

## Usage

### As a Go Package

#### Simple Usage

```go
package main

import (
    "fmt"
    "github.com/sethgrid/xeger"
)

func main() {
    // Generate a single string with defaults
    str, err := xeger.Generate("[0-9]{3}-[0-9]{2}-[0-9]{4}")
    if err != nil {
        panic(err)
    }
    fmt.Println(str) // e.g., "123-45-6789"
}
```

#### Advanced Usage

```go
package main

import (
    "fmt"
    "math/rand"
    "time"
    "github.com/sethgrid/xeger"
)

func main() {
    // Create a generator with custom options
    opts := xeger.Options{
        MaxLength: 100,
        MaxRepeat: 5,
        ShortBias: true,
    }
    
    rng := rand.New(rand.NewSource(time.Now().UnixNano()))
    g, err := xeger.New("^[A-Z][a-z]+$", opts, rng)
    if err != nil {
        panic(err)
    }
    
    // Generate multiple strings
    for i := 0; i < 10; i++ {
        str, err := g.Next()
        if err != nil {
            panic(err)
        }
        fmt.Println(str)
    }
}
```

### As a CLI Tool

```bash
# Generate a single string
xeger "[0-9]{3}-[0-9]{2}-[0-9]{4}"

# Generate multiple strings
xeger -n 5 "^[A-Z][a-z]{2,5}[0-9]{2}$"

# Control max length
xeger -maxlen 20 ".*"

# Combine flags
xeger -n 10 -maxlen 50 "[a-zA-Z0-9_]+"
```

#### CLI Flags

- `-n N`: Number of strings to generate (default: 1)
- `-maxlen N`: Maximum length of generated strings (default: 64)

## Options

The `Options` struct controls generator behavior:

- **MaxLength** (int): Maximum total length of generated strings. If 0, no limit is enforced. Default: 64
- **MaxRepeat** (int): Maximum repetitions for unbounded quantifiers (`*`, `+`, `{n,}`). If 0, defaults to 10. Default: 10
- **ShortBias** (bool): When true, biases repetitions toward smaller values using a geometric distribution. Default: true

## Supported Regex Features

- ✅ Literal strings
- ✅ Character classes (`[a-z]`, `[0-9]`, etc.)
- ✅ Alternation (`a|b|c`)
- ✅ Quantifiers (`*`, `+`, `?`, `{n}`, `{n,m}`)
- ✅ Concatenation
- ✅ Capture groups (treated as their contents)
- ✅ Anchors (`^`, `$`) - don't generate characters
- ✅ Word boundaries (`\b`) - don't generate characters
- ✅ Any character (`.`)

## Limitations

The following features are not supported in v1:

- Backreferences (`\1`, etc.)
- Lookahead/lookbehind assertions
- PCRE extensions beyond Go's standard regexp syntax

## Examples

```go
// Phone number pattern
xeger.Generate(`^\d{3}-\d{3}-\d{4}$`)
// Output: "123-456-7890"

// Email-like pattern
xeger.Generate(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
// Output: "user@example.com"

// Alternation
xeger.Generate(`(red|green|blue)`)
// Output: "green" (randomly chosen)

// Character class with quantifier
xeger.Generate(`[0-9a-f]{8}`)
// Output: "a1b2c3d4"
```

## Testing

Run the test suite:

```bash
go test ./...
```

The test suite includes:
- Roundtrip property tests (every generated string matches the original pattern)
- MaxLength constraint validation
- MaxRepeat constraint validation
- Alternation coverage tests
- Character class tests
- Quantifier tests

## Project Structure

```
xeger/
├── cmd/
│   └── xeger/          # CLI executable
├── xeger.go            # Main package API
├── xeger_test.go       # Test suite
├── go.mod
└── README.md
```

## License

[Add your license here]

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Roadmap

- [ ] Deterministic enumeration mode
- [ ] Weighted alternation
- [ ] Minimal example generation (shrinkers)
- [ ] Better support for negated character classes
- [ ] Performance optimizations for large patterns
