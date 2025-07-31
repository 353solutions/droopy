# Going Up

Elevators might seems simple, but the are wickedly tricky to control.
In this workshop,
your job is to write a controller for an elevator that won't cut a passenger in half or dive into the basement.
You are probably going to fail.

This workshop touches upon many practical topics in a hands-on way:
- Design patterns and software architecture
- Event driven programming
- State machines
- Testing (simulation, fuzzing, unit, ...)

You'll need to design, work with a simulator over TCP, and validate that your code is correct.
One of the great outcomes of this workshop is that you'll take the stairs more often.

### The Simulator

During the workshop, you receive events and send commands to `droopy`[^fn-droopy],
an elevator simulator.

The simulator has the following hardware:

- A motor that makes the car go up and down.
- A door that can open and close.
- A panel of 4 buttons inside the elevator car.
- Up and Down request in each of the 4 floors.

People interact with the elevator by pressing buttons.

-  Pn: Press button for floor n in the elevator car
-  Un: Press up button on floor n
-  Dn: Press down button on floor n

The simulator doesn't do anything when a button is pressed.
It does sends an event to the connected controller via TCP.

```
     Person → [Droopy] ↔ [Controller]
         Button       TCP      
```

The simulator sends the following events to the controller:

- Pn: Panel button for floor n pressed
- Un: Up button on floor n pressed
- Dn: Down button floor n pressed
- An: Approaching floor n (still in motion)
- Sn: Stopped at floor n (safe to open door)
- On: Door open on floor n (doors have fully opened)
- Cn: Door closed on floor n (safe to move)

The controller can send the following commands to the simulator:

- MU: Move up
- MD: Move down
- S: Stop at next floor (generates Sn event when stopped)
- DO: Open door (will generate On event when done)
- DC: Close door (will generate Cn event when done)
- CPn: Clear panel button n
- CUn: Clear up button n
- CDn: Clear down button n
- R: Reset

When droopy crashs, I enters a permanent crashed state that can only be reset by sending the reset (R) command.


[^fn-droopy]: https://www.youtube.com/watch?v=20lcB9c-Qa0
