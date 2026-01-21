package xeger

import (
	"math/rand"
	"regexp/syntax"
	"time"
)

// Options controls the behavior of the generator.
type Options struct {
	// MaxLength is the maximum total length of generated strings.
	// If 0, no limit is enforced.
	MaxLength int

	// MaxRepeat is the maximum repetitions for unbounded quantifiers (*, +, {n,}).
	// If 0, defaults to 10.
	MaxRepeat int

	// ShortBias biases repetitions toward smaller values when true.
	ShortBias bool
}

// DefaultOptions returns sensible default options.
func DefaultOptions() Options {
	return Options{
		MaxLength: 64,
		MaxRepeat: 10,
		ShortBias: true,
	}
}

// Generator generates strings matching a regular expression pattern.
type Generator struct {
	tree *syntax.Regexp
	rng  *rand.Rand
	opts Options
}

// New creates a new Generator for the given pattern.
// If rng is nil, a new random source is created using the current time.
func New(pattern string, opts Options, rng *rand.Rand) (*Generator, error) {
	re, err := syntax.Parse(pattern, syntax.Perl)
	if err != nil {
		return nil, err
	}
	re = re.Simplify()

	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	return &Generator{
		tree: re,
		rng:  rng,
		opts: opts,
	}, nil
}

// Next generates the next matching string.
func (g *Generator) Next() (string, error) {
	buf := g.gen(g.tree, make([]rune, 0))
	return string(buf), nil
}

// Generate is a convenience function that generates a single string from a pattern.
func Generate(pattern string) (string, error) {
	g, err := New(pattern, DefaultOptions(), nil)
	if err != nil {
		return "", err
	}
	return g.Next()
}

// gen recursively generates runes from the regex AST.
func (g *Generator) gen(re *syntax.Regexp, buf []rune) []rune {
	// Check max length constraint
	if g.opts.MaxLength > 0 && len(buf) >= g.opts.MaxLength {
		return buf
	}

	switch re.Op {
	case syntax.OpNoMatch:
		// Never matches, return empty
		return buf

	case syntax.OpEmptyMatch:
		// Matches empty string
		return buf

	case syntax.OpLiteral:
		// Append literal runes
		return append(buf, re.Rune...)

	case syntax.OpCharClass:
		// Pick a character from the class
		r := g.pickFromClass(re.Rune)
		return append(buf, r)

	case syntax.OpAnyCharNotNL:
		// Any character except newline
		// Pick from printable ASCII range for simplicity
		r := rune(32 + g.rng.Intn(95)) // 32-126 (printable ASCII)
		return append(buf, r)

	case syntax.OpAnyChar:
		// Any character including newline
		if g.rng.Intn(10) == 0 {
			return append(buf, '\n')
		}
		r := rune(32 + g.rng.Intn(95))
		return append(buf, r)

	case syntax.OpBeginLine, syntax.OpBeginText:
		// Anchors don't generate characters
		return buf

	case syntax.OpEndLine, syntax.OpEndText:
		// Anchors don't generate characters
		return buf

	case syntax.OpWordBoundary, syntax.OpNoWordBoundary:
		// Word boundaries don't generate characters
		return buf

	case syntax.OpConcat:
		// Concatenate all sub-expressions
		for _, sub := range re.Sub {
			buf = g.gen(sub, buf)
			if g.opts.MaxLength > 0 && len(buf) >= g.opts.MaxLength {
				break
			}
		}
		return buf

	case syntax.OpAlternate:
		// Pick one of the alternatives
		if len(re.Sub) == 0 {
			return buf
		}
		i := g.rng.Intn(len(re.Sub))
		return g.gen(re.Sub[i], buf)

	case syntax.OpCapture:
		// Treat capture groups as their contents
		if len(re.Sub) == 0 {
			return buf
		}
		// For captures, generate from the first (and typically only) sub-expression
		return g.gen(re.Sub[0], buf)

	case syntax.OpStar:
		// Zero or more
		return g.genRepeat(re.Sub[0], buf, 0, g.effectiveMaxRepeat())

	case syntax.OpPlus:
		// One or more
		return g.genRepeat(re.Sub[0], buf, 1, g.effectiveMaxRepeat())

	case syntax.OpQuest:
		// Zero or one
		if g.rng.Intn(2) == 1 {
			return g.gen(re.Sub[0], buf)
		}
		return buf

	case syntax.OpRepeat:
		// {min,max} repetition
		min, max := re.Min, re.Max
		if max < 0 {
			max = g.effectiveMaxRepeat()
		}
		return g.genRepeat(re.Sub[0], buf, min, max)

	default:
		// Unknown operation, return as-is
		return buf
	}
}

// genRepeat generates a repeated sub-expression, appending to the existing buffer.
func (g *Generator) genRepeat(sub *syntax.Regexp, buf []rune, min, max int) []rune {
	if max < min {
		max = min
	}
	n := g.pickRepetition(min, max)
	for i := 0; i < n; i++ {
		buf = g.gen(sub, buf)
		if g.opts.MaxLength > 0 && len(buf) >= g.opts.MaxLength {
			break
		}
	}
	return buf
}

// effectiveMaxRepeat returns the effective maximum repeat count.
func (g *Generator) effectiveMaxRepeat() int {
	if g.opts.MaxRepeat > 0 {
		return g.opts.MaxRepeat
	}
	return 10
}

// pickRepetition chooses a repetition count between min and max.
func (g *Generator) pickRepetition(min, max int) int {
	if min == max {
		return min
	}
	if g.opts.ShortBias {
		// Geometric-ish bias toward smaller values
		n := min
		for n < max && g.rng.Intn(2) == 1 {
			n++
		}
		return n
	}
	return min + g.rng.Intn(max-min+1)
}

// pickFromClass picks a random rune from a character class.
// The ranges slice contains pairs [lo0, hi0, lo1, hi1, ...]
func (g *Generator) pickFromClass(ranges []rune) rune {
	if len(ranges) == 0 {
		return ' '
	}

	// Handle negated character classes
	// If first range is 0 and last is MaxRune, it's likely negated
	if len(ranges) >= 2 && ranges[0] == 0 && ranges[len(ranges)-1] == 0x10FFFF {
		// For negated classes, we need to pick from the complement
		// This is complex, so for now we'll pick from a common subset
		// In practice, negated classes are less common
		return g.pickFromRanges(ranges[1 : len(ranges)-1])
	}

	return g.pickFromRanges(ranges)
}

// pickFromRanges picks a random rune from the given ranges.
func (g *Generator) pickFromRanges(ranges []rune) rune {
	// Compute total size of all ranges
	total := 0
	for i := 0; i < len(ranges); i += 2 {
		if i+1 >= len(ranges) {
			break
		}
		lo, hi := ranges[i], ranges[i+1]
		if hi >= lo {
			total += int(hi-lo) + 1
		}
	}

	if total == 0 {
		return ' '
	}

	// Pick a random position
	k := g.rng.Intn(total)

	// Find which range contains position k
	for i := 0; i < len(ranges); i += 2 {
		if i+1 >= len(ranges) {
			break
		}
		lo, hi := ranges[i], ranges[i+1]
		if hi < lo {
			continue
		}
		size := int(hi-lo) + 1
		if k < size {
			return lo + rune(k)
		}
		k -= size
	}

	return ' ' // fallback
}
