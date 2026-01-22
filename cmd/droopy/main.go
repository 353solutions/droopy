package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"time"
)

var pool *ConnPool

func sockListener(addr string, ch chan<- Message) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			panic(err)
		}
		pool.Add(conn)
		go handler(conn, ch)
	}
}

func handler(conn net.Conn, ch chan<- Message) {
	defer pool.Remove(conn)
	defer conn.Close()

	s := bufio.NewScanner(conn)
	for s.Scan() {
		ch <- Message{"ctrl", s.Text()}
	}

	if err := s.Err(); err != nil && !errors.Is(err, io.EOF) {
		panic(err)
	}
}

func sendEvent(msg string) {
	pool.Broadcast(msg)
}

func stdinListener(ch chan<- Message) {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		ch <- Message{"stdin", s.Text()}
	}

	if err := s.Err(); err != nil {
		panic(err)
	}

	ch <- Message{"stdin", "EOF"}
}

func ticker(ch chan<- Message) {
	for range time.Tick(100 * time.Millisecond) {
		ch <- Message{"ticker", "T"}
	}
}

type Message struct {
	Origin  string
	Payload string
}

const MaxFloor = 4

type MotorState byte

const (
	MotorUp MotorState = iota + 1
	MotorDown
	MotorOff
)

func (s MotorState) String() string {
	switch s {
	case MotorUp:
		return "UP"
	case MotorDown:
		return "DOWN"
	case MotorOff:
		return "OFF"
	}

	return fmt.Sprintf("MotorState(%d)", s)
}

type DoorState byte

const (
	DoorOpening DoorState = iota + 1
	DoorOpen
	DoorClosing
	DoorClosed
)

func (s DoorState) String() string {
	switch s {
	case DoorOpening:
		return "OPENING"
	case DoorOpen:
		return "OPEN"
	case DoorClosing:
		return "CLOSING"
	case DoorClosed:
		return "CLOSED"
	}

	return fmt.Sprintf("DoorState(%d)", s)
}

type Elevator struct {
	// floors start at 1
	panel      [MaxFloor + 1]bool // in car panel
	up         [MaxFloor + 1]bool // up buttons on floors
	down       [MaxFloor + 1]bool // down buttons on floors
	floor      int                // Current floor, starts at 1
	motor      MotorState
	door       DoorState
	stopping   bool
	crashed    bool
	crashCount int // Total crashes this session
	eventTime  int // Start of event such as door opening, move ...
}

func (e *Elevator) crash() {
	if !e.crashed {
		e.crashed = true
		e.crashCount++
	}
}

func (e *Elevator) Reset() {
	for i := range e.panel {
		e.panel[i] = false
	}

	for i := range e.up {
		e.up[i] = false
	}

	for i := range e.down {
		e.down[i] = false
	}

	e.floor = 1
	e.motor = MotorOff
	e.door = DoorClosed
	e.crashed = false
}

const (
	ticksPerFloor = 40
	ticksPerDoor  = 20
	approachTicks = 10
)

// setDoor sets door state, returns crash message.
func (e *Elevator) setDoor(state DoorState) string {
	switch {
	case e.motor != MotorOff:
		e.crash()
		return "crash: door command while moving"
	case e.door == DoorClosed && state == DoorOpening:
		e.door = DoorOpening
		e.eventTime = 0
		return ""
	case e.door == DoorOpen && state == DoorClosing:
		e.door = DoorClosing
		e.eventTime = 0
		return ""
	}

	e.crash()
	return fmt.Sprintf("crash: door %s in state %s", state, e.door)
}

// setMotor sets motor state, returns crash message.
func (e *Elevator) setMotor(state MotorState) string {
	if e.door != DoorClosed {
		e.crash()
		return fmt.Sprintf("crash: motor command while door %s", e.door)
	}

	if e.motor != MotorOff {
		e.crash()
		return fmt.Sprintln("crash: motor command while moving")
	}

	if e.motor == MotorOff && state == MotorOff {
		e.crash()
		return "crash: motor already off"
	}

	e.motor = state
	e.eventTime = 0
	return ""
}

func nextFloor(floor int, motor MotorState) int {
	if motor == MotorUp {
		return floor + 1
	}

	return floor - 1
}

func cmdFloor(cmd string) int {
	return int(cmd[len(cmd)-1] - '0')
}

// Handle handles a command, returns an event to report (empty string if no event).
func (e *Elevator) Handle(cmd string) string {
	if cmd == "R" { // Reset
		e.Reset()
		return ""
	}

	// Ignore commands when crashed
	if e.crashed {
		return ""
	}

	switch cmd {
	case "P1", "P2", "P3", "P4":
		e.panel[cmdFloor(cmd)] = true
		return cmd
	case "CP1", "CP2", "CP3", "CP4":
		e.panel[cmdFloor(cmd)] = false
		return cmd
	case "U1", "U2", "U3":
		e.up[cmdFloor(cmd)] = true
		return cmd
	case "CU1", "CU2", "CU3":
		e.up[cmdFloor(cmd)] = false
		return cmd
	case "D2", "D3", "D4":
		e.down[cmdFloor(cmd)] = true
		return cmd
	case "CD2", "CD3", "CD4":
		e.down[cmdFloor(cmd)] = false
		return cmd
	case "DO":
		return e.setDoor(DoorOpening)
	case "DC":
		return e.setDoor(DoorClosing)
	case "MU":
		return e.setMotor(MotorUp)
	case "MD":
		return e.setMotor(MotorDown)
	case "S":
		if e.stopping {
			e.crash()
			return "crash: already stopping"
		}

		if e.motor == MotorOff {
			e.crash()
			return "crash: not moving"
		}

		e.stopping = true
	case "T":
		e.eventTime++

		if e.door == DoorOpening || e.door == DoorClosing {
			if e.eventTime <= ticksPerDoor {
				return ""
			}

			var evt string
			if e.door == DoorOpening {
				e.door = DoorOpen
				evt = "O"
			} else {
				e.door = DoorClosed
				evt = "C"
			}
			e.eventTime = 0
			return fmt.Sprintf("%s%d", evt, e.floor)
		}

		if e.motor == MotorUp || e.motor == MotorDown {
			if e.eventTime == ticksPerFloor {
				floor := nextFloor(e.floor, e.motor)
				if floor > MaxFloor {
					e.crash()
					return "crash: out of the roof"
				}

				if floor < 1 {
					e.crash()
					return "crash: into the basement"
				}

				e.floor = floor
				e.eventTime = 0

				if e.stopping {
					e.stopping = false
					e.motor = MotorOff
					return fmt.Sprintf("S%d", e.floor)
				}
			}

			if e.eventTime == ticksPerFloor-approachTicks {
				floor := nextFloor(e.floor, e.motor)
				return fmt.Sprintf("A%d", floor)
			}
		}
	default:
		e.crash()
		return fmt.Sprintf("crash: unknown command - %q", cmd)
	}

	return ""
}

func (e *Elevator) statusStr() string {
	if e.crashed {
		return "CRASH"
	}

	if e.stopping {
		return "STOPPING"
	}

	if e.motor == MotorUp || e.motor == MotorDown {
		return e.motor.String()
	}

	if e.door == DoorClosed || e.door == DoorClosing || e.door == DoorOpen || e.door == DoorOpening {
		return e.door.String()
	}

	panic(fmt.Sprintf("unknown state: %#v", e))
}

func buttonsStr(buttons []bool) string {
	buf := make([]byte, len(buttons)-1) // 0 is a placeholder
	for i, v := range buttons[1:] {
		if v {
			buf[i] = '0' + byte(i+1)
		} else {
			buf[i] = '-'
		}
	}

	return string(buf)
}

func (e *Elevator) String() string {
	var buf bytes.Buffer
	count := pool.Len()
	conn := " "
	if count > 0 {
		conn = "*"
	}

	fmt.Fprintf(&buf, "[%sFLOOR %d", conn, e.floor)
	fmt.Fprintf(&buf, "| %-8s", e.statusStr())
	fmt.Fprintf(&buf, "| P:%s", buttonsStr(e.panel[:]))
	fmt.Fprintf(&buf, "| U:%s", buttonsStr(e.up[:]))
	fmt.Fprintf(&buf, "| D:%s", buttonsStr(e.down[:]))
	fmt.Fprintf(&buf, " ] : ")

	return buf.String()
}

func debug(format string, args ...any) {
	if os.Getenv("DEBUG") == "" {
		return
	}

	fmt.Printf(format, args...)
}

func sigHandler(ch chan<- Message) {
	sch := make(chan os.Signal, 1)
	signal.Notify(sch, os.Interrupt)
	<-sch
	ch <- Message{"signal", "Q"}
}

var (
	version     string = "v0.11.0" // filled by goreleaser
	showVersion bool
	simAddr     = ":10000"

	//go:embed help.txt
	help string
)

func validateAddr(addr string) error {
	i := strings.Index(addr, ":")
	if i == -1 {
		return fmt.Errorf("%q: missing ':'", addr)
	}

	port, err := strconv.Atoi(addr[i+1:])
	if err != nil {
		return fmt.Errorf("%q: bad port - %w", addr, err)
	}

	if port < 0 || port > 65_535 {
		return fmt.Errorf("%q: bad port number", addr)
	}

	return nil
}

func farewellMessage(crashCount int) string {
	switch {
	case crashCount == 0:
		return "Perfect run! Not a single crash. You're going up in the world! 🎯"
	case crashCount == 1:
		return "Just 1 crash. Everyone makes mistakes on their first day. 🤕"
	case crashCount <= 3:
		return fmt.Sprintf("%d crashes. The elevator inspector wants a word with you. 📋", crashCount)
	case crashCount <= 5:
		return fmt.Sprintf("%d crashes. OSHA is preparing paperwork. 📝", crashCount)
	case crashCount <= 10:
		return fmt.Sprintf("%d crashes. The building is considering stairs only. 🚶", crashCount)
	case crashCount <= 15:
		return fmt.Sprintf("%d crashes. Your insurance premiums just went UP. 📈", crashCount)
	case crashCount <= 20:
		return fmt.Sprintf("%d crashes. NASA called - they need you to NOT work on rockets. 🚀❌", crashCount)
	case crashCount <= 30:
		return fmt.Sprintf("%d crashes. The passengers are forming a support group. 😰", crashCount)
	case crashCount <= 50:
		return fmt.Sprintf("%d crashes. The elevator is filing for emancipation. 🏃", crashCount)
	default:
		return fmt.Sprintf("%d crashes. At this point, it's impressive. Impressively bad. 💀", crashCount)
	}
}

func runCmd() {
	pool = NewConnPool()

	flag.BoolVar(&showVersion, "version", false, "show version and exit")
	flag.StringVar(&simAddr, "addr", simAddr, "simulator address")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]\n", path.Base(os.Args[0]))
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println(help)
	}
	flag.Parse()

	if showVersion {
		fmt.Printf("%s version %s\n", path.Base(os.Args[0]), version)
		os.Exit(0)
	}

	if err := validateAddr(simAddr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	debug("address: %s\n", simAddr)

	ch := make(chan Message)

	go sockListener(simAddr, ch)
	go stdinListener(ch)
	go sigHandler(ch)
	go ticker(ch)

	var e Elevator
	e.Reset()
	defer func() {
		fmt.Println()
		fmt.Println(farewellMessage(e.crashCount))
	}()

	lastState := e.String()
	fmt.Print(lastState)
	for msg := range ch {
		if msg.Payload != "T" {
			debug("%-5s: %s\n", msg.Origin, msg.Payload)
		}

		if msg.Payload == "EOF" || msg.Payload == "Q" {
			return
		}

		if msg.Origin == "ctrl" {
			debug("c:%s", msg.Payload)
		}

		var evt string
		switch msg.Payload {
		case "":
			// Ignore user hitting Enter
		case "H":
			fmt.Println(help)
		case "Q":
			return
		default:
			evt = e.Handle(msg.Payload)
			if evt != "" && !strings.HasPrefix(evt, "crash:") {
				debug("event: %s\n", evt)
				sendEvent(evt)
			}
		}

		state := e.String()
		if state != lastState || msg.Payload == "" || msg.Payload == "H" { // Empty message -> user hit Enter
			if msg.Origin != "stdin" || msg.Origin == "ctrl" {
				fmt.Println()
			}
			if strings.HasPrefix(evt, "crash:") {
				fmt.Printf("\033[31m%s\033[39m\n", evt)
			}
			fmt.Print(state)
			lastState = state
		}
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "play" {
		if err := playCmd(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
		return
	}

	runCmd()
}
