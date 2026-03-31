package logger

import (
	"bytes"
	"os"
	"testing"
)

func TestSetDebug(t *testing.T) {
	SetDebug(true)
	if !IsDebug() {
		t.Error("Expected debug to be true after SetDebug(true)")
	}

	SetDebug(false)
	if IsDebug() {
		t.Error("Expected debug to be false after SetDebug(false)")
	}
}

func TestDebugOutput(t *testing.T) {
	SetDebug(true)

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Debug("test message %s", "arg")

	w.Close()
	os.Stdout = oldStdout

	buf.ReadFrom(r)
	output := buf.String()

	if output == "" {
		t.Error("Expected debug output, got empty string")
	}
}

func TestInfo(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Info("info message")

	w.Close()
	os.Stdout = oldStdout

	buf.ReadFrom(r)
	output := buf.String()

	if output == "" {
		t.Error("Expected info output, got empty string")
	}
}

func TestWarn(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Warn("warn message")

	w.Close()
	os.Stdout = oldStdout

	buf.ReadFrom(r)
	output := buf.String()

	if output == "" {
		t.Error("Expected warn output, got empty string")
	}
}

func TestError(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Error("error message")

	w.Close()
	os.Stdout = oldStdout

	buf.ReadFrom(r)
	output := buf.String()

	if output == "" {
		t.Error("Expected error output, got empty string")
	}
}

func TestDebugOnlyOutputsWhenEnabled(t *testing.T) {
	SetDebug(false)

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Debug("should not appear")

	w.Close()
	os.Stdout = oldStdout

	buf.ReadFrom(r)
	output := buf.String()

	if output != "" {
		t.Errorf("Expected no output when debug is disabled, got: %s", output)
	}
}

func TestFormatArgs(t *testing.T) {
	SetDebug(false)

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Info("test %d %s", 123, "hello")

	w.Close()
	os.Stdout = oldStdout

	buf.ReadFrom(r)
	output := buf.String()

	if output == "" {
		t.Error("Expected info output, got empty string")
	}
}
