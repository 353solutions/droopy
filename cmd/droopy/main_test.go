package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestElevator_setDoor(t *testing.T) {
	type testCase struct {
		doorState  DoorState
		motorState MotorState
		newState   DoorState
	}

	validCases := map[testCase]bool{
		{DoorClosed, MotorOff, DoorOpening}: true,
		{DoorOpen, MotorOff, DoorClosing}:   true,
	}

	for _, door := range []DoorState{DoorOpening, DoorOpen, DoorClosing, DoorClosed} {
		for _, motor := range []MotorState{MotorUp, MotorDown, MotorOff} {
			for _, newState := range []DoorState{DoorOpening, DoorOpen, DoorClosing, DoorClosed} {
				tc := testCase{door, motor, newState}
				_, valid := validCases[tc]
				name := door.String() + ":" + motor.String() + ":" + newState.String()

				t.Run(name, func(t *testing.T) {
					e := &Elevator{
						door:  door,
						motor: motor,
					}

					msg := e.setDoor(newState)
					if !valid {
						if !strings.HasPrefix(msg, "crash:") {
							t.Fatal("expected crash")
						}
						return
					}

					if e.door != newState {
						t.Fatal(e.door)
					}
				})
			}
		}
	}
}

func TestElevator_setMotor(t *testing.T) {
	type testCase struct {
		motorState MotorState
		doorState  DoorState
		newState   MotorState
	}

	validCases := map[testCase]bool{
		{MotorOff, DoorClosed, MotorUp}:   true,
		{MotorOff, DoorClosed, MotorDown}: true,
	}

	for _, motor := range []MotorState{MotorUp, MotorDown, MotorOff} {
		for _, door := range []DoorState{DoorOpening, DoorOpen, DoorClosing, DoorClosed} {
			for _, newState := range []MotorState{MotorUp, MotorDown, MotorOff} {
				tc := testCase{motor, door, newState}
				_, valid := validCases[tc]
				name := motor.String() + ":" + door.String() + ":" + newState.String()

				t.Run(name, func(t *testing.T) {
					e := &Elevator{
						door:  door,
						motor: motor,
					}

					msg := e.setMotor(newState)
					if !valid {
						if !strings.HasPrefix(msg, "crash:") {
							t.Fatal("expected crash")
						}
						return
					}

					if e.motor != newState {
						t.Fatal(e.door)
					}
				})
			}
		}
	}
}

func TestElevetor_HandleButton(t *testing.T) {
	var e Elevator

	var cases = []struct {
		cmds    []string
		buttons []bool
	}{
		{[]string{"P1", "P2", "P3", "P4"}, e.panel[:]},
		{[]string{"U1", "U2", "U3"}, e.up[:]},
		{[]string{"D2", "D3", "D4"}, e.down[:]},
	}

	for _, c := range cases {
		for _, cmd := range c.cmds {
			t.Run(cmd, func(t *testing.T) {
				e.Reset()
				msg := e.Handle(cmd)

				if msg != cmd {
					t.Fatal(msg)
				}

				if c.buttons[cmdFloor(cmd)] != true {
					t.Fatal(c.buttons)
				}
			})
		}
	}
}

func TestElevator_HandleStop(t *testing.T) {
	var e Elevator

	for _, stopping := range []bool{true, false} {
		for _, motor := range []MotorState{MotorUp, MotorDown, MotorOff} {
			name := fmt.Sprintf("%s:%v", motor.String(), stopping)
			t.Run(name, func(t *testing.T) {
				e.Reset()
				e.stopping = stopping
				e.motor = motor

				msg := e.Handle("S")
				if stopping || motor == MotorOff {
					if !strings.HasPrefix(msg, "crash:") {
						t.Fatal("expected crash")
					}
					return
				}

				if strings.HasPrefix(msg, "crash:") {
					t.Fatal("unexpected crash")
				}
			})
		}
	}
}

func TestElevator_HandleClearButton(t *testing.T) {
	var e Elevator

	var cases = []struct {
		setCmd   string
		clearCmd string
		buttons  []bool
	}{
		{"P1", "CP1", e.panel[:]},
		{"P2", "CP2", e.panel[:]},
		{"P3", "CP3", e.panel[:]},
		{"P4", "CP4", e.panel[:]},
		{"U1", "CU1", e.up[:]},
		{"U2", "CU2", e.up[:]},
		{"U3", "CU3", e.up[:]},
		{"D2", "CD2", e.down[:]},
		{"D3", "CD3", e.down[:]},
		{"D4", "CD4", e.down[:]},
	}

	for _, c := range cases {
		t.Run(c.clearCmd, func(t *testing.T) {
			e.Reset()

			// Set button
			msg := e.Handle(c.setCmd)
			if msg != c.setCmd {
				t.Fatalf("set: expected %q, got %q", c.setCmd, msg)
			}

			floor := cmdFloor(c.setCmd)
			if c.buttons[floor] != true {
				t.Fatalf("after set: button %d should be true", floor)
			}

			// Clear button
			msg = e.Handle(c.clearCmd)
			if msg != c.clearCmd {
				t.Fatalf("clear: expected %q, got %q", c.clearCmd, msg)
			}

			if c.buttons[floor] != false {
				t.Fatalf("after clear: button %d should be false", floor)
			}
		})
	}
}

func TestElevator_HandleTick(t *testing.T) {
	t.Skip("TODO")
}
