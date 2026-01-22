package droopy

import (
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func waitForServer(t *testing.T, addr string, timeout time.Duration) {
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

func startElevator(t *testing.T, addr string) {
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "droopy")

	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/droopy")
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
	lst, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}

	defer lst.Close()
	return lst.Addr().(*net.TCPAddr).Port
}

func TestClient(t *testing.T) {
	port := freePort(t)
	addr := fmt.Sprintf(":%d", port)
	startElevator(t, addr)

	// Create client
	c, err := NewClient(WithAddr("localhost" + addr))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer c.Close()

	if err := c.Send("R"); err != nil {
		t.Fatalf("failed to send reset: %v", err)
	}

	if err := c.Send("MU"); err != nil {
		t.Fatalf("failed to send motor up: %v", err)
	}

	evt, err := c.Recv()
	if err != nil {
		t.Fatalf("failed to receive approaching event: %v", err)
	}
	if evt != "A2" {
		t.Errorf("expected approaching event A2, got %q", evt)
	}

	if err := c.Send("S"); err != nil {
		t.Fatalf("failed to send stop: %v", err)
	}

	// Receive stopped event (S2 - stopped at floor 2)
	evt, err = c.Recv()
	if err != nil {
		t.Fatalf("failed to receive stopped event: %v", err)
	}
	if evt != "S2" {
		t.Errorf("expected stopped event S2, got %q", evt)
	}
}
