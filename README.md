# droopy - An Elevator Simulator

![](droopy.png)

Droopy is an elevator simulator used in various workshops.
It's based on [lifty][lifty] by the amazing David Beazley. 

Droopy resides in a 4-floor building. It has the following hardware:
- A door that can open and close
- A panel of 4 buttons inside the car
- Up and down buttons at each floor
- A motor that can go up and down

People can interact with Droppy by pressing buttons.
- Pn: Press button for floor n inside the car
- Un: Press up button on floor n
- Dn: Press down button on floor n

Droopy won't do anything when a button is pressed,
it'll notify the controller via TCP socket on that a button was pressed.
The controller can send command telling Droopy what do do.


```
          +------+  --- event ---> +----------+
Jane ---> |Droopy|                 |Controller| 
          +------+  <-- command -- +----------+
   button                 TCP
```


Events from Droopy to the controller:

- Pn: Panel button for floor n was pressed
- Un: Up button on floor n was pressed
- Dn: Down button floor n was pressed
- An: Approaching floor n (still in motion)
- Sn: Stopped at floor n (safe to open door)
- On: Door open on floor n (doors have fully opened)
- Cn: Door closed on floor n (now safe to move)

Command from the controller to Droopy:

- MU: Start moving up
- MD: Start moving down
- S: Stop at the next floor (generates Sn event when stopped)
- DO: Open door (generates On event when done)
- DC: Close the door (generates Cn event when done)
- CPn: Clear panel button n
- CUn: Clear up button n
- CDn: Clear down button n
- R: Reset

If the controller sends Droopy an unsafe state (say open door when moving),
Droopy will move into a crashed state and stop responding to any commands.
You can reset Droopy by entering the "R" (reset) command.

You should write a controller program that runs that elevator and:
1. Works like an actual elevator
2. Never crashes the elevator


[lifty]: https://github.com/dabeaz/lifty
