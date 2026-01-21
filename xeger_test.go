package xeger

import (
	"regexp"
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		valid   bool
	}{
		{"empty", "", true},
		{"simple literal", "foo", true},
		{"char class", "^[0-9a-z]+$", true},
		{"brackets", `^[0-9a-z]+\[[0-9]{3,5}\]$`, true},
		{"star", `foo.*`, true},
		{"alternation", `a(x*)b(y|z)c`, true},
		{"punctuation", `[a-z][\.\?!]\s+[A-Z]`, true},
		{"invalid", `[where_is_the_closing_bracket?`, false},
		{"email-like", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, true},
		{"phone-like", `^\d{3}-\d{3}-\d{4}$`, true},
		{"word boundaries", `\bword\b`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := New(tt.pattern, DefaultOptions(), nil)
			if tt.valid {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				// Generate multiple strings and verify they match
				for i := 0; i < 10; i++ {
					s, err := g.Next()
					if err != nil {
						t.Errorf("Next() error: %v", err)
						continue
					}
					matched, err := regexp.MatchString(tt.pattern, s)
					if err != nil {
						t.Errorf("regexp.MatchString error: %v", err)
						continue
					}
					if !matched {
						t.Errorf("generated string %q does not match pattern %q", s, tt.pattern)
					}
				}
			} else {
				if err == nil {
					t.Errorf("expected an error for invalid pattern")
				}
			}
		})
	}
}

func TestRoundtripProperty(t *testing.T) {
	// Core property: every generated string must match the original pattern
	patterns := []string{
		"abc",
		"[0-9]+",
		"[a-z]{3,5}",
		"(foo|bar|baz)",
		"a*b+c?",
		"^[A-Z][a-z]+$",
		".*",
		"\\d{2,4}",
		"[a-zA-Z0-9_]+",
	}

	for _, pattern := range patterns {
		t.Run(pattern, func(t *testing.T) {
			re := regexp.MustCompile(pattern)
			g, err := New(pattern, DefaultOptions(), nil)
			if err != nil {
				t.Fatalf("New() error: %v", err)
			}

			// Generate many strings and verify they all match
			for i := 0; i < 100; i++ {
				s, err := g.Next()
				if err != nil {
					t.Fatalf("Next() error: %v", err)
				}
				if !re.MatchString(s) {
					t.Errorf("generated string %q does not match pattern %q", s, pattern)
				}
			}
		})
	}
}

func TestMaxLength(t *testing.T) {
	opts := DefaultOptions()
	opts.MaxLength = 10

	g, err := New(".*", opts, nil)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	for i := 0; i < 50; i++ {
		s, err := g.Next()
		if err != nil {
			t.Fatalf("Next() error: %v", err)
		}
		if len(s) > opts.MaxLength {
			t.Errorf("generated string length %d exceeds MaxLength %d: %q", len(s), opts.MaxLength, s)
		}
	}
}

func TestMaxRepeat(t *testing.T) {
	opts := DefaultOptions()
	opts.MaxRepeat = 3

	g, err := New("a+", opts, nil)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	// With MaxRepeat=3, a+ should generate at most 3 'a's
	for i := 0; i < 50; i++ {
		s, err := g.Next()
		if err != nil {
			t.Fatalf("Next() error: %v", err)
		}
		count := 0
		for _, r := range s {
			if r == 'a' {
				count++
			}
		}
		if count > opts.MaxRepeat {
			t.Errorf("generated %d 'a's, exceeds MaxRepeat %d: %q", count, opts.MaxRepeat, s)
		}
	}
}

func TestAlternationCoverage(t *testing.T) {
	// Test that alternation options appear across many samples
	pattern := "(a|b|c)"
	g, err := New(pattern, DefaultOptions(), nil)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		s, err := g.Next()
		if err != nil {
			t.Fatalf("Next() error: %v", err)
		}
		seen[s] = true
	}

	// We should see all three options in 100 samples
	if len(seen) < 2 {
		t.Errorf("expected to see multiple alternation options in 100 samples, saw: %v", seen)
	}
}

func TestConvenienceFunction(t *testing.T) {
	s, err := Generate("test")
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if s != "test" {
		t.Errorf("Generate() = %q, want %q", s, "test")
	}

	// Test with pattern that should match
	matched, _ := regexp.MatchString("test", s)
	if !matched {
		t.Errorf("generated string %q does not match pattern", s)
	}
}

func TestCharClass(t *testing.T) {
	patterns := []string{
		"[0-9]",
		"[a-z]",
		"[A-Z]",
		"[0-9a-f]",
		"[a-zA-Z]",
	}

	for _, pattern := range patterns {
		t.Run(pattern, func(t *testing.T) {
			g, err := New(pattern, DefaultOptions(), nil)
			if err != nil {
				t.Fatalf("New() error: %v", err)
			}

			re := regexp.MustCompile("^" + pattern + "$")
			for i := 0; i < 50; i++ {
				s, err := g.Next()
				if err != nil {
					t.Fatalf("Next() error: %v", err)
				}
				if !re.MatchString(s) {
					t.Errorf("generated string %q does not match pattern %q", s, pattern)
				}
			}
		})
	}
}

func TestQuantifiers(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		check   func(string) bool
	}{
		{"star", "a*", func(s string) bool {
			for _, r := range s {
				if r != 'a' {
					return false
				}
			}
			return true
		}},
		{"plus", "a+", func(s string) bool {
			if len(s) == 0 {
				return false
			}
			for _, r := range s {
				if r != 'a' {
					return false
				}
			}
			return true
		}},
		{"question", "a?", func(s string) bool {
			return s == "" || s == "a"
		}},
		{"repeat exact", "a{3}", func(s string) bool {
			return s == "aaa"
		}},
		{"repeat range", "a{2,4}", func(s string) bool {
			if len(s) < 2 || len(s) > 4 {
				return false
			}
			for _, r := range s {
				if r != 'a' {
					return false
				}
			}
			return true
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := New(tt.pattern, DefaultOptions(), nil)
			if err != nil {
				t.Fatalf("New() error: %v", err)
			}

			re := regexp.MustCompile("^" + tt.pattern + "$")
			for i := 0; i < 50; i++ {
				s, err := g.Next()
				if err != nil {
					t.Fatalf("Next() error: %v", err)
				}
				if !re.MatchString(s) {
					t.Errorf("generated string %q does not match pattern %q", s, tt.pattern)
				}
				if !tt.check(s) {
					t.Errorf("generated string %q failed custom check for pattern %q", s, tt.pattern)
				}
			}
		})
	}
}
