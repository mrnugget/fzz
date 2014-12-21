package main

import (
	"bytes"
	"testing"
)

var printTests = []struct {
	lines    []string
	cols     int
	rows     int
	expected string
}{
	{ // Normal
		[]string{"foobar\n"},
		80,
		20,
		"\nfoobar\n",
	},
	{ // Too many lines
		[]string{"foobar\n", "foobar\n", "foobar\n"},
		99,
		2,
		"\nfoobar\nfoobar",
	},
	{ // Lines too wide
		[]string{"foobar\n"},
		3,
		99,
		"\nfoo\n",
	},
	{
		[]string{"xxxYYYzzz\n", "xxxYYYzzz\n", "xxxYYYzzz\n", "xxxYYYzzz\n", "xxxYYYzzz\n"},
		6,
		4,
		"\nxxxYYY\nxxxYYY\nxxxYYY\nxxxYYY",
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
