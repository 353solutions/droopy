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

func TestElevator_HandleTick(t *testing.T) {
	t.Skip("TODO")
}
