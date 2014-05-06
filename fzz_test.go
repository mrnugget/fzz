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
