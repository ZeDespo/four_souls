package four_souls

import (
	"errors"
	"math/rand"
)

// A slice of cards that act as the
// deck / discard pile of each card type
//
// Most methods here are based off: https://github.com/golang/go/wiki/SliceTricks
type deck []card

// Place a card at the end of the slice (top of the deck)
func (d *deck) append(c ...card) {
	*d = append(*d, c...)
}

// Delete some card at certain index
// Since this method doesn't perform any checks, be
// certain that the index is in bounds before invoking this
// method directly
//
// This method accounts for pointers to prevent memory leaks
// Preserves order
func (d *deck) delete(i uint8) {
	copy((*d)[i:], (*d)[i+1:])
	(*d)[len(*d)-1] = lootCard{} // Any type that implements card can go here
	*d = (*d)[:len(*d)-1]
}

// Returns the length of the deck in uint8
func (d deck) len() uint8 {
	return uint8(len(d))
}

// If we have two decks of cards, merge them together
// Make sure they are the same type!
// if onTop, append, else, prepend
// if shuffle, shuffle the deck after merging
func (d *deck) merge(d2 deck, onTop bool, shuffle bool) {
	if onTop {
		*d = append(*d, d2...)
	} else {
		*d = append(d2, *d...)
	}
	if shuffle {
		d.shuffle()
	}
}

func (d deck) peek() (card, error) {
	var c card
	err := errors.New("empty deck")
	l := d.len()
	if l > 0 {
		c, err = d[l-1], nil
	}
	return c, err
}

// Pop a card from the last index in the
// slice (akin to drawing the card)
// return: the card and an error if the card was not found
func (d *deck) pop() (card, error) {
	var c card
	err := errors.New("empty deck")
	l := d.len()
	if l > 0 {
		c, err = (*d)[l-1], nil
		d.delete(l - 1)
	}
	return c, err
}

// Pop a card based off it's id, no matter where
// it exists in the slice
// return: the card and an error if the card was not found
func (d *deck) popById(cardId uint16) (card, error) {
	var c card
	var i uint8
	var err error
	if c, i, err = d.search(cardId); err == nil {
		d.delete(i)
	}
	return c, err
}

// Pop a card based off its index
// return: the card and an error if the card was not found
func (d *deck) popByIndex(i uint8) (card, error) {
	var c card
	err, l := errors.New("index out of bounds"), d.len()
	if l > 0 {
		if i > 0 && i < l {
			c, err = (*d)[i], nil
			d.delete(i)
		}
	}
	return c, err
}

// Add a card to the beginning of the slice
// (bottom of the deck)
func (d *deck) prepend(c ...card) {
	*d = append(c, *d...)
}

// From a slice of ids, find the indexes of all the cards with the id
// Will return a
func (d deck) scan(ids ...uint16) ([]card, map[uint16]uint8) {
	size := len(ids)
	idSet := make(map[uint16]struct{}, size)
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	idMaps := make(map[uint16]uint8, size) // key: id, value: index
	foundCards := make([]card, 0, size)
	for i, c := range d {
		if _, ok := idSet[c.getId()]; ok {
			foundCards = append(foundCards, c)
			idMaps[c.getId()] = uint8(i)
		}
	}
	return foundCards, idMaps
	return foundCards, idMaps
}

// Find a card in the slice based off its id
// return: if error is nil, returns the card and the index it was found
func (d deck) search(cardId uint16) (card, uint8, error) {
	var c card
	err := errors.New("card not found")
	var i uint8
	for i = 0; i < d.len(); i++ {
		card := d[i]
		if card.getId() == cardId {
			c, err = card, nil
		}
	}
	return c, i, err
}

// Shuffle the deck in place
func (d *deck) shuffle() {
	rand.Shuffle(len(*d), func(i, j int) { (*d)[i], (*d)[j] = (*d)[j], (*d)[i] })
}

// Necessary only for the monster card.
// If a player attacks the deck instead of one of the two active monster zones,
// and if that attacked card is a monster, it will be overlayed on
// on of the monster zones. The sole purpose of this structure is to provide this
// functionality.
type activeSlot []monsterCard

func (as activeSlot) isEmpty() bool {
	return len(as) == 0
}

// Check the top of the stack
// Will be used either for effects targeting the monster
// or for attack.
func (as activeSlot) peek() *monsterCard {
	var monster monsterCard
	length := len(as)
	if length > 0 {
		monster = as[length-1]
	}
	return &monster
}

// Put a new monster in a monster zone, either
// from the deck or the discardPile pile.
func (as *activeSlot) push(item monsterCard) {
	*as = append(*as, item)
}

// pop a monster (destroy it)
func (as *activeSlot) pop() monsterCard {
	var item monsterCard
	length := len(*as)
	if length > 0 {
		idx := length - 1
		item, *as = (*as)[idx], (*as)[:idx]
	}
	return item
}

func (as activeSlot) Size() int {
	return len(as)
}

// Very similar to a linked list deckNode, except the top pointer to the end of the list
type eventNode struct {
	id    uint
	event event
	next  *eventNode
	top   *eventNode // Points to the top of the stack (head points to tail)
}

func (en eventNode) checkActivateItemEvent() error {
	var err = errors.New("not an activate item event")
	if activate, ok := en.event.e.(activateEvent); ok {
		if _, ok := activate.c.(*treasureCard); ok {
			err = nil
		}
	}
	return err
}

func (p player) checkAttackingPlayer() error {
	var err error
	if !p.inBattle {
		err = errors.New("not the attacking player")
	}
	return err
}

func (en eventNode) checkAttackRoll(expected uint8, nextEn eventNode) error {
	var err error
	if err = en.checkDiceRoll(expected); err == nil {
		if _, ok := nextEn.event.e.(declareAttackEvent); !ok {
			err = errors.New("not an attack dice roll")
		}
	}
	return err
}

// Check if the monster dealt damage ONLY
func (en eventNode) checkDamageFromMonster(mId uint16) (damageEvent, error) {
	var damage damageEvent
	var ok bool
	var err = errors.New("not a damage from monster event")
	if damage, ok = en.event.e.(damageEvent); ok {
		if damage.monster != nil {
			if mId == 0 || damage.monster.id == mId {
				err = nil
			}
		}
	}
	return damage, err
}

// Check if an event deckNode is a "damage to player's character" event.
func (en eventNode) checkDamageToPlayer(cId uint16) (damageEvent, error) {
	var damage damageEvent
	var ok bool
	var err = errors.New("not a damage to self event")
	if damage, ok = en.event.e.(damageEvent); ok {
		if damage.target.getId() == cId {
			err = nil
		}
	}
	return damage, err
}

// Check if a monster dealt damage to a specific player based off their ids.
func (en eventNode) checkDamageToPlayerFromMonster(cId, mId uint16) (damageEvent, error) {
	damage, err := en.checkDamageToPlayer(cId)
	if err == nil && damage.monster != nil {
		err = errors.New("not a damage event due to combat")
		if damage.monster.id == mId {
			err = nil
		}
	}
	return damage, err
}

// Check if damage to some monster occurred
func (en eventNode) checkDamageToMonster() (damageEvent, error) {
	var damage damageEvent
	var ok bool
	var err = errors.New("not a damage to monster event")
	if damage, ok = en.event.e.(damageEvent); ok {
		if _, ok = damage.target.(*monsterCard); ok {
			err = nil
		}
	}
	return damage, err
}

// Check if a specific monster took some kind of damage
func (en eventNode) checkDamageToSpecificMonster(id uint16) (damageEvent, error) {
	var damage damageEvent
	var err = errors.New("not a damage to given monster id")
	if damage, err = en.checkDamageToMonster(); err == nil {
		if damage.target.getId() != id {
			err = errors.New("not a damage to given monster id")
		}
	}
	return damage, err
}

func (en eventNode) checkDeath(p *player) error {
	err := errors.New("not a deathPenalty to self event")
	if _, ok := en.event.e.(deathOfCharacterEvent); ok {
		if p != nil {
			if en.event.p.Character.id == p.Character.id {
				err = nil
			}
		}
		err = nil
	}
	return err
}

func (en eventNode) checkDeclareAttack(p *player) (declareAttackEvent, error) {
	var declareAttack declareAttackEvent
	var ok bool
	var err = errors.New("not a declare attack event")
	if declareAttack, ok = en.event.e.(declareAttackEvent); ok {
		if en.event.p.Character.id == p.Character.id {
			err = nil
		}
	}
	return declareAttack, err
}

func (en eventNode) checkDiceRoll(expected uint8) error {
	var dre diceRollEvent
	var ok bool
	var err = errors.New("not a valid dice roll event")
	if dre, ok = en.event.e.(diceRollEvent); ok {
		if expected == 0 || dre.n == expected {
			err = nil
		}
	}
	return err
}

func (en eventNode) checkEndOfTurn(p *player) error {
	var ok bool
	var err = errors.New("not a valid end of turn event")
	if _, ok = en.event.e.(endTurnEvent); ok {
		if en.event.p.Character.id == p.Character.id {
			err = nil
		}
	}
	return err
}

func (en eventNode) checkIntentionToAttack() (intentionToAttackEvent, error) {
	var intention intentionToAttackEvent
	var err = errors.New("not an intention to attack event")
	var ok bool
	if intention, ok = en.event.e.(intentionToAttackEvent); ok {
		err = nil
	}
	return intention, err
}

func (en eventNode) checkStartOfTurn(p *player) error {
	var ok bool
	var err = errors.New("not a valid start turn event")
	if _, ok = en.event.e.(startOfTurnEvent); ok {
		if en.event.p.Character.id == p.Character.id {
			err = nil
		}
	}
	return err
}

// Although it's called a stack, it mostly behaves like a linked list.
// Cards such as Dice Shard, Soul Heart, and Book of Belial can
// manipulate certain events placed anywhere on the stack.
// Need to be able to add and remove events from the stack without
// significant overhead.
type eventStack struct {
	head      *eventNode
	idCounter uint // auto-incremented counter to assign deckNode ids from 0 to 2^64 - 1
	size      uint // the current size of the stack
}

// Add or subtract a value from a diceroll on the event stack
// n int8: The positive or negative value to apply to the diceroll
// rollNode *eventNode: The deckNode containing the dice roll event.
func (es *eventStack) addToDiceRoll(n int8, rollNode *eventNode) {
	if oldEvent, ok := rollNode.event.e.(diceRollEvent); ok {
		x := int8(oldEvent.n) + n
		if x <= 0 {
			x = 1
		} else if x > 6 {
			x = 6
		}
		rollNode.event.e = diceRollEvent{n: uint8(x)}
	} else {
		panic("not a node that contains a dice roll.")
	}
}

// Although we have deletion, fizzling should be called by any card
// effect that calls for a cancellation of some event (butter bean, no, holy card).
// This method will check if the deckNode has already been deleted, and if
// it has not, will delete the deckNode.
func (es *eventStack) fizzle(en *eventNode) error {
	_, err := es.search(en.id)
	if err == nil {
		en.event.e = fizzledEvent{}
	}
	return err
}

func (es eventStack) isEmpty() bool {
	var b bool
	if es.head == nil {
		b = true
	}
	return b
}

func (es *eventStack) peek() *eventNode {
	var top *eventNode
	if es.head != nil {
		top = es.head.top
	}
	return top
}

// Instead of returning the deckNode, just return the event
// inside the deckNode.
func (es *eventStack) peekEvent() (event, error) {
	var e event
	var err = errors.New("no items in stack")
	if es.head != nil {
		e, err = es.head.top.event, nil
	}
	return e, err
}

func (es *eventStack) pop() *eventNode {
	var top *eventNode
	if es.head != nil {
		top = es.head.top
		if top.next == nil { // head of deckNode
			es.head = nil
		} else {
			es.head.top = top.next
		}
		top.next = nil
		es.size -= 1
	}
	return top
}

func (es *eventStack) push(event event) {
	newNode := &eventNode{id: es.idCounter, event: event}
	if es.head == nil {
		es.head = newNode
	} else {
		newNode.next = es.head.top
	}
	es.head.top = newNode
	es.idCounter += 1
	es.size += 1
}

// Prevent damage on a damage deckNode.
// Return an error if the deckNode is not a damage deckNode or if the deckNode is not on the event stack
func (es *eventStack) preventDamage(i uint8, damageNode *eventNode) error {
	var err = errors.New("not a damage node")
	if oldEvent, ok := damageNode.event.e.(damageEvent); ok {
		x := oldEvent.n - i
		if x == 0 {
			err = es.fizzle(damageNode)
		} else {
			damageNode.event = event{p: damageNode.event.p, e: damageEvent{n: x}}
			err = nil
		}
	}
	return err
}

func (es *eventStack) search(id uint) (*eventNode, error) {
	curr, err := es.head.top, errors.New("node not found")
	for curr != nil {
		if curr.id == id {
			err = nil
			break
		} else if curr.id < id { // know the deckNode does not exist
			curr = nil
		} else {
			curr = curr.next
		}
	}
	return curr, err
}
