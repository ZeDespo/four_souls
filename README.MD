# "The Binding of Isaac: Four Souls", PC port.

This work in progress is a dedicated PC port for the card game "The Binding of Isaac: Four Souls."

The project will be built out in stages. Current stage: 0.

1) Build a local only version that relies on command line output to debug the game's mechanics.

2) Build a local only version with a simple UI (no animations, no effects)

3) Build a multiplayer server to enable network play with a front end. 

4) Refine animations. 



PASSIVE TRIGGER EVENTS
0) LUCKY ROCK ONLY: Activating an item (resolvement)

1) Start / End of your turn (resolvement)

2) Damage to self / to monster (preventative and resolvement)

3) Death of player / monster (preventative and resolvement)

4) Rolling dice (resolvement)

5) Intention to Attack a monster (resolvement)

6) Declaring an attack (resolvement)


Damage and Death: 

For cards like Guppy's Hairball, Broken Ankh, The Dead Cat, Guppy's Collar, and 1-UP that prevent damage / death events 
from happening need to be handled differently.