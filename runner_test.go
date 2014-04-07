package main

import "bytes"
import "testing"

func TestRunWithInput(t *testing.T) {
	testTarget := new(bytes.Buffer)
	kill := make(chan bool)
	testInput := []byte{'f', 'o', 'o'}

	runner := &Runner{
		target: testTarget,
		template: "echo {{}} bar",
		placeholder: "{{}}",
		maxCol: 80,
		maxRow: 20,
	}

	runner.runWithInput(testInput, kill)

	output := testTarget.String()
	if output != "foo bar\n" {
		t.Errorf("unexpected output: %v", output)
	}
}
