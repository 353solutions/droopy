package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/353solutions/droopy"
)

func playCmd(args []string) error {
	fs := flag.NewFlagSet("play", flag.ExitOnError)
	var addr string
	fs.StringVar(&addr, "addr", ":10000", "simulator address")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: droopy play [options]\n")
		fmt.Println("Options:")
		fs.PrintDefaults()
		fmt.Println()
		fmt.Println("Commands (read from stdin):")
		fmt.Println("  SEND <cmd>       - send command to the simulator")
		fmt.Println("  SLEEP <duration> - sleep for duration (e.g., 3s, 100ms)")
		fmt.Println("  WAIT <event>     - wait for event (e.g., A2, O1)")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	client, err := droopy.NewClient(droopy.WithAddr(addr))
	if err != nil {
		return fmt.Errorf("connecting to simulator: %w", err)
	}
	defer client.Close()

	scanner := bufio.NewScanner(os.Stdin)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		fmt.Printf("> %s\n", line)

		cmd := parts[0]
		switch cmd {
		case "SEND":
			if len(parts) < 2 {
				return fmt.Errorf("line %d: SEND requires a command", lineNum)
			}

			command := parts[1]
			if err := client.Send(command); err != nil {
				return fmt.Errorf("line %d: sending command: %w", lineNum, err)
			}

		case "SLEEP":
			if len(parts) < 2 {
				return fmt.Errorf("line %d: SLEEP requires a duration", lineNum)
			}

			duration, err := time.ParseDuration(parts[1])
			if err != nil {
				return fmt.Errorf("line %d: parsing duration: %w", lineNum, err)
			}

			time.Sleep(duration)
		case "WAIT":
			if len(parts) < 2 {
				return fmt.Errorf("line %d: WAIT requires an event", lineNum)
			}

			event := parts[1]
			for {
				received, err := client.Recv()
				if err != nil {
					return fmt.Errorf("line %d: receiving event: %w", lineNum, err)
				}

				fmt.Printf("< %s\n", received)

				if received == event {
					break
				}
			}
		default:
			return fmt.Errorf("line %d: unknown command %q", lineNum, cmd)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	return nil
}
