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
	printer := NewPrinter(testTarget, 80, 1)

	for i := 0; i < printer.maxRow+1; i++ {
		_, err := printer.Print(testLine)
		if err != nil {
			t.Errorf("error printing: %s", err)
		}
	}

	if testTarget.String() != "\n"+testLine {
		t.Errorf("wrong output written: %s", testTarget.String())
	}
}

func TestPrintTooWide(t *testing.T) {
	testTarget := new(bytes.Buffer)
	testLine := "foobar\n"
	printer := NewPrinter(testTarget, 3, 1)

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
	expected := "\n" + testLine + "\n" + testLine

	if output != expected {
		t.Errorf("wrong output written. expected: %q, got: %q", expected, output)
	}
}
