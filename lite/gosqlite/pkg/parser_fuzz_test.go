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
		program := p.ParseProgram()

		// Check for tokenizer errors. We expect some inputs to cause tokenizer errors,
		// but we should ensure the tokenizer doesn't crash or enter an infinite loop.
		if len(l.Errors()) > 0 {
			// In a real fuzzing setup, these errors would be logged and analyzed.
			// t.Logf("tokenizer errors: %v", l.Errors())
		}

		// Check for parser errors. Similar to tokenizer errors, we expect some inputs
		// to cause parsing errors, but the parser should remain stable.
		if len(p.Errors()) > 0 {
			// In a real fuzzing setup, these errors would be logged and analyzed.
			// t.Logf("parser errors: %v", p.Errors())
		}

		// Conceptual checks for AST validity and stability.
		// In a more sophisticated fuzzing setup, one would:
		// - Compare the parsed AST against a reference parser for valid inputs.
		// - Track memory usage to detect leaks (e.g., using Go's testing.MemStats or external tools).
		// - Monitor execution time to detect infinite loops or excessive computation.
		// - Ensure that for valid SQL inputs, the AST is correctly formed and semantically sound.
		if program == nil && len(p.Errors()) == 0 && len(l.Errors()) == 0 {
			t.Errorf("parser returned nil program with no reported errors for input: %q", string(data))
		}

		// Enterprise-level fuzzing would involve:
		// - Integration with CI/CD pipelines for continuous fuzzing on every code change.
		// - Advanced corpus management to store interesting inputs that trigger new code paths or bugs.
		// - Coverage-guided fuzzing tools (e.g., go-fuzz, libFuzzer) to maximize code coverage.
		// - Oracle-based testing where the fuzzer compares output against a known-good implementation.
	})
}
