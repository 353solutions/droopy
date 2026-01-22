package main

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func waitForServer(t *testing.T, addr string, timeout time.Duration) {
	t.Helper()
	addr = "localhost" + addr
	start := time.Now()
	for time.Since(start) <= timeout {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			return
		}

		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("server on %s did not start after %v", addr, timeout)
}

func startSimulator(t *testing.T, addr string) {
	t.Helper()
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "droopy")

	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build droopy: %v\n%s", err, out)
	}

	cmd := exec.Command(binPath, "-addr", addr)
	if _, err := cmd.StdinPipe(); err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start droopy: %v", err)
	}
	t.Cleanup(func() {
		if err := cmd.Process.Kill(); err != nil {
			t.Logf("warning: can't kill %d - %v", cmd.Process.Pid, err)
		}
	})

	waitForServer(t, addr, 2*time.Second)
}

func freePort(t *testing.T) int {
	t.Helper()
	lst, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}

	defer lst.Close()
	return lst.Addr().(*net.TCPAddr).Port
}

func TestPlayCmd(t *testing.T) {
	port := freePort(t)
	addr := fmt.Sprintf(":%d", port)
	startSimulator(t, addr)

	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "droopy-play")

	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build droopy: %v\n%s", err, out)
	}

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "send command",
			input: `SEND R
SEND MU`,
			expected: []string{
				"> SEND R",
				"> SEND MU",
			},
		},
		{
			name: "sleep command",
			input: `SEND R
SLEEP 100ms`,
			expected: []string{
				"> SEND R",
				"> SLEEP 100ms",
			},
		},
		{
			name: "wait for event",
			input: `SEND R
SEND MU
WAIT A2`,
			expected: []string{
				"> SEND R",
				"> SEND MU",
				"> WAIT A2",
				"< A2",
			},
		},
		{
			name: "complex scenario",
			input: `SEND R
SEND MU
WAIT A2
SEND S
WAIT S2
SEND DO`,
			expected: []string{
				"> SEND R",
				"> SEND MU",
				"> WAIT A2",
				"< A2",
				"> SEND S",
				"> WAIT S2",
				"< S2",
				"> SEND DO",
			},
		},
		{
			name: "comments and blank lines",
			input: `# This is a comment
SEND R

# Start moving up
SEND MU
WAIT A2`,
			expected: []string{
				"> SEND R",
				"> SEND MU",
				"> WAIT A2",
				"< A2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, binPath, "play", "-addr", "localhost"+addr)
			cmd.Stdin = strings.NewReader(tt.input)

			output, err := cmd.CombinedOutput()
			if err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					t.Fatalf("command timed out: %v\nOutput:\n%s", err, output)
				}
			}

			outputStr := string(output)
			for _, want := range tt.expected {
				if !strings.Contains(outputStr, want) {
					t.Errorf("expected %q in output:\n%s", want, outputStr)
				}
			}
		})
	}
}

func TestPlayCmd_Errors(t *testing.T) {
	port := freePort(t)
	addr := fmt.Sprintf(":%d", port)
	startSimulator(t, addr)

	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "droopy-play")

	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build droopy: %v\n%s", err, out)
	}

	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "missing command for SEND",
			input:         "SEND",
			expectedError: "SEND requires a command",
		},
		{
			name:          "missing duration for SLEEP",
			input:         "SLEEP",
			expectedError: "SLEEP requires a duration",
		},
		{
			name:          "invalid duration for SLEEP",
			input:         "SLEEP invalid",
			expectedError: "parsing duration",
		},
		{
			name:          "missing event for WAIT",
			input:         "WAIT",
			expectedError: "WAIT requires an event",
		},
		{
			name:          "unknown command",
			input:         "UNKNOWN",
			expectedError: "unknown command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, binPath, "play", "-addr", "localhost"+addr)
			cmd.Stdin = strings.NewReader(tt.input)

			output, err := cmd.CombinedOutput()
			if err == nil {
				t.Fatalf("expected error but command succeeded. Output:\n%s", output)
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, tt.expectedError) {
				t.Errorf("expected error containing %q, got:\n%s", tt.expectedError, outputStr)
			}
		})
	}
}
