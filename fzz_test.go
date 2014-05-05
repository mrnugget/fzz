package main

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
