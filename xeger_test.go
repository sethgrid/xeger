package xeger

import (
	"log"
	"os"
	"testing"
)

func TestEarlyErr(t *testing.T) {
	var tests = []struct {
		Pattern string
		IsValid bool
	}{
		{``, true},
		{`^[0-9a-z]+\[[0-9]{3,5}\]$`, true},
		{`foo.*`, true},
		{`a(x*)b(y|z)c`, true},
		{`[a-z][\.\?!]\s+[A-Z]`, true},
		{`[where_is_the_closing_bracket?`, false},
	}

	for _, test := range tests {
		t.Logf("Test: %s", test.Pattern)
		iRe, err := NewInverseRegex(test.Pattern)
		if test.IsValid && err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if !test.IsValid && err == nil {
			t.Errorf("expected an error and got none")
		}
		if iRe != nil {
			// pass in standard log settings
			iRe.logger = log.New(os.Stderr, "", log.LstdFlags)
			_ = iRe.Generate()
		}
	}
}
