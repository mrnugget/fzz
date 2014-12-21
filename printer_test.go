package main

import (
	"bytes"
	"testing"
)

func TestPrintNormal(t *testing.T) {
	testTarget := new(bytes.Buffer)
	testLine := "foobar\n"
	printer := NewPrinter(testTarget, 80, 20)

	n, err := printer.Print(testLine)
	if err != nil {
		t.Errorf("error printing: %s", err)
	}
	if n != len(testLine) {
		t.Errorf("bytes written wrong. expected: %d, got: %d", len(testLine), n)
	}

	output := testTarget.String()
	if output != "\n"+testLine {
		t.Errorf("wrong output written")
	}
}

func TestPrintTooLong(t *testing.T) {
	testTarget := new(bytes.Buffer)
	testLine := "foobar\n"
	printer := NewPrinter(testTarget, 99, 2)

	for i := 0; i < printer.maxRow+1; i++ {
		_, err := printer.Print(testLine)
		if err != nil {
			t.Errorf("error printing: %s", err)
		}
	}

	// Cut of the newline char on the last line
	if testTarget.String() != "\nfoobar\nfoobar" {
		t.Errorf("wrong output written: %q", testTarget.String())
	}
}

func TestPrintTooWide(t *testing.T) {
	testTarget := new(bytes.Buffer)
	testLine := "foobar\n"
	printer := NewPrinter(testTarget, 3, 99)

	expectedN := printer.maxCol + 1
	expectedStr := "\nfoo\n"

	n, err := printer.Print(testLine)
	if err != nil {
		t.Errorf("error printing: %s", err)
	}

	// Prints maxcol + '\n'
	if n != expectedN {
		t.Errorf("bytes written wrong. expected: %d, got: %d", expectedN, n)
	}

	output := testTarget.String()
	if output != expectedStr {
		t.Errorf("wrong output written. expected: %q, got: %q", expectedStr, output)
	}
}

func TestPrintTooWideTooLong(t *testing.T) {
	testTarget := new(bytes.Buffer)
	testLines := []string{
		"foobarbarfoo\n",
		"foobarbarfoo\n",
		"foobarbarfoo\n",
		"foobarbarfoo\n",
		"foobarbarfoo\n",
	}
	expectedStr := "\nfoobarbarf\nfoobarbarf\nfoobarbarf\nfoobarbarf"

	printer := NewPrinter(testTarget, 10, 4)

	for _, line := range testLines {
		_, err := printer.Print(line)
		if err != nil {
			t.Errorf("error printing: %s", err)
		}
	}
	output := testTarget.String()
	if output != expectedStr {
		t.Errorf("wrong output written. expected: %q, got: %q", expectedStr, output)
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
