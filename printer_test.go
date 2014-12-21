package main

import (
	"bytes"
	"testing"
)

var input = []string{
	"xxxYYYzzz\n",
	"xxxYYYzzz\n",
	"xxxYYYzzz\n",
	"xxxYYYzzz\n",
	"xxxYYYzzz\n",
}

var printTests = []struct {
	lines    []string
	cols     int
	rows     int
	expected string
}{
	{ // Normal
		input,
		999,
		999,
		"\nxxxYYYzzz\nxxxYYYzzz\nxxxYYYzzz\nxxxYYYzzz\nxxxYYYzzz\n",
	},
	{ // Too many lines
		input,
		999,
		3,
		"\nxxxYYYzzz\nxxxYYYzzz\nxxxYYYzzz",
	},
	{ // Lines too wide
		input,
		3,
		999,
		"\nxxx\nxxx\nxxx\nxxx\nxxx\n",
	},
	{ // Too many too wide lines
		input,
		3,
		3,
		"\nxxx\nxxx\nxxx",
	},
}

func TestPrint(t *testing.T) {
	for _, tt := range printTests {
		target := new(bytes.Buffer)
		printer := NewPrinter(target, tt.cols, tt.rows)

		for _, line := range tt.lines {
			_, err := printer.Print(line)
			if err != nil {
				t.Errorf("error printing: %s", err)
			}
		}

		actual := target.String()
		if actual != tt.expected {
			t.Errorf("wrong output written. got=%q, expected=%q", actual, tt.expected)
		}
	}
}

func TestReset(t *testing.T) {
	testTarget := new(bytes.Buffer)
	testLine := "foobar\n"
	printer := NewPrinter(testTarget, 80, 1)

	_, err := printer.Print(testLine)
	if err != nil {
		t.Errorf("error printing: %s", err)
	}

	printer.Reset()

	_, err = printer.Print(testLine)
	if err != nil {
		t.Errorf("error printing: %s", err)
	}

	output := testTarget.String()
	// Cut of the newline char on the last line
	expected := "\n" + testLine + testLine[:len(testLine)-1]

	if output != expected {
		t.Errorf("wrong output written. expected: %q, got: %q", expected, output)
	}
}
