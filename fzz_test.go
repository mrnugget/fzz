package main

import "bytes"
import "testing"

var removeLastCharacterTests = []struct {
	input, expected []byte
}{
	{
		input:    []byte("foo"),
		expected: []byte("fo"),
	},
	{
		input:    []byte("f"),
		expected: nil,
	},
	{
		input:    []byte("foö"),
		expected: []byte("fo"),
	},
	{
		input:    []byte(""),
		expected: []byte(""),
	},
}

func TestRemoveLastCharacter(t *testing.T) {
	for _, tt := range removeLastCharacterTests {
		result := removeLastCharacter(tt.input)
		if tt.expected == nil && result != nil {
			t.Errorf("nil expected. actual: %v", result)
		}
		if len(result) != len(tt.expected) {
			t.Errorf("result slice wrong length. expected: %d, actual: %d", len(tt.expected), len(result))
		}
	}
}

func TestReadCharacter(t *testing.T) {
	test := "föobar"
	source := &bytes.Buffer{}
	source.WriteString(test)

	ch := readCharacter(source)

	for _, c := range test {
		char := <-ch
		if string(char) != string(c) {
			t.Errorf("read character wrong. expected: %q, actual: %q", c, string(char))
		}
	}
}

var extractInputTests = []struct {
	args        []string
	p           string
	resultInput string
	resultArgs  []string
}{
	{
		args:        []string{"ag", "{{}}", "*.go"},
		p:           "{{}}",
		resultInput: "",
		resultArgs:  []string{"ag", "{{}}", "*.go"},
	},
	{
		args:        []string{"ag", "%%", "*.go"},
		p:           "%%",
		resultInput: "",
		resultArgs:  []string{"ag", "%%", "*.go"},
	},
	{
		args:        []string{"ag", "{{foobar}}", "*.go"},
		p:           "{{}}",
		resultInput: "foobar",
		resultArgs:  []string{"ag", "{{}}", "*.go"},
	},
	{
		args:        []string{"ag", "%foobar%", "*.go"},
		p:           "%%",
		resultInput: "foobar",
		resultArgs:  []string{"ag", "%%", "*.go"},
	},
	{
		args:        []string{"ag", "%foobar%", "*.go"},
		p:           "%%",
		resultInput: "foobar",
		resultArgs:  []string{"ag", "%%", "*.go"},
	},
	{
		args:        []string{"ag", "{{foo bar}}", "*.go"},
		p:           "{{}}",
		resultInput: "foo bar",
		resultArgs:  []string{"ag", "{{}}", "*.go"},
	},
	{
		args:        []string{"ag", "%foo bar%", "*.go"},
		p:           "%%",
		resultInput: "foo bar",
		resultArgs:  []string{"ag", "%%", "*.go"},
	},
}

func TestExtractInput(t *testing.T) {
	for _, tt := range extractInputTests {
		rinput, rargs := extractInput(tt.args, tt.p)

		if rinput != tt.resultInput {
			t.Errorf("resultInput wrong. expected: %v, got: %v", tt.resultInput, rinput)
			continue
		}

		if len(rargs) != len(tt.resultArgs) {
			t.Errorf("resultArgs have wrong length. expected: %v, got: %v", len(tt.resultArgs), len(rargs))
			continue
		}

		for i, _ := range tt.resultArgs {
			if rargs[i] != tt.resultArgs[i] {
				t.Errorf("resultArg wrong. expected: %v, got: %v", tt.resultArgs[i], rargs[i])
			}
		}
	}
}
