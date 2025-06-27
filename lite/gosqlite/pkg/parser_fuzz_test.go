package pkg

import (
	"testing"
)

func FuzzParser(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("parser panicked: %v", r)
			}
		}()

		l := NewTokenizer(string(data), 1024)
		p := NewParser(l, 100, 10)
		_ = p.ParseProgram()

		// Check for tokenizer errors
		if len(l.Errors()) > 0 {
			// We expect some inputs to cause tokenizer errors, so we don't fail the test.
			// t.Logf("tokenizer errors: %v", l.Errors())
		}

		// Check for parser errors
		if len(p.Errors()) > 0 {
			// We expect some inputs to cause parser errors, so we don't fail the test.
			// t.Logf("parser errors: %v", p.Errors())
		}

		// TODO: Add more sophisticated checks for memory leaks, infinite loops, and incorrect output.
		// This would involve tracking memory usage, execution time, and comparing ASTs for valid inputs.

		// TODO: Integrate with CI/CD for continuous fuzzing.
		// TODO: Implement corpus management to store interesting inputs.
	})
}
