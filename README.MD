# "The Binding of Isaac: Four Souls", PC port.
This work in progress is a dedicated PC port for the card game 
[_The Binding of Isaac: Four Souls_](http://foursoulsspoiler.com/).
I have not provided a license for this repository, therefore, if you are not the project's owner,
you are only authorized to read and run this repository.

### Objective
Create a one-for-one recreation of _Four Souls_ without compromising the game's 
[rules](http://pop-life.com/foursouls/rule-book.php), while simultaneously automating several key, yet mundane,
mechanics such as checking card activation requirements, rolling / applying dice rolls to card effects & battle, and
card placement management.

### Current Phase
This project is not in a runnable state. The main goal at this point is to finish programming the basic rules of the game
along with creating unit tests that ensure: 
1) Each card in the base game works as intended
2) Each type of event (starting a turn, battling a monster, rolling dice) will resolve as intended, independently
of other events.
3) The basic mechanics of the game (initiating an action, prompting players for a reaction, stacking events) work as
intended.

The game will exist as a CLI application until the project is as free as obvious bugs as can be. 

### Future plans

Once the current phase is over: 
1) Provide doc strings to each basic mechanic and event function (cards are already documented)
2) Create the multiplayer server and test in a local environment (one game, two separate players) 
3) Implement goroutines to lessen wait times for computationally expensive functions
4) Start work on the frontend (Typescript) and enable browser play.


