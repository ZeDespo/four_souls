package four_souls

import "errors"

type eventHolder interface {
	eHolder()
}

type activateEvent struct {
	c activeCards
	f cardEffect
}

func (a activateEvent) eHolder() {}

// Damage a monster or character card
type damageEvent struct {
	monster *monsterCard // Will not be nil if event's pushed due to a missed attack
	target  combatTarget
	n       uint8
}

func (d damageEvent) eHolder() {}

// A player dies
type deathOfCharacterEvent struct{}

func (d deathOfCharacterEvent) eHolder() {}

// Declaring a target for an numAttacks
// Will always follow the intention to numAttacks
type declareAttackEvent struct {
	m *monsterCard
}

func (d declareAttackEvent) eHolder() {}

// Declare a target for purchase form the shop
// Will always follow the intention to purchase
type declarePurchaseEvent struct {
	i uint8
	t treasureCard
}

func (d declarePurchaseEvent) eHolder() {}

// The resulting dice roll
type diceRollEvent struct {
	n uint8
}

func (d diceRollEvent) eHolder() {}

// Signals the end of a turn
type endTurnEvent struct{}

func (e endTurnEvent) eHolder() {}

// This is not an actual event. If an event is cancelled by cards such as
// Butter Bean, Holy Card, and Guppy's Paw, that event will be replaced by this one
// Note that cards whose proper conditions are not met upon resolving will also fizzle out
// without being replaced with this event.
// No card or action in the game can respond to this event.
type fizzledEvent struct{}

func (d fizzledEvent) eHolder() {}

// Declaring an intention to numAttacks some monster
// This is the time to play effects that change the monster zone.
type intentionToAttackEvent struct {
	m *monsterCard
}

func (d intentionToAttackEvent) eHolder() {}

// Declaring an intention to numPurchases an itemCard
// This is the time to play effects that change the shop
type intentionToPurchaseEvent struct{}

func (d intentionToPurchaseEvent) eHolder() {}

// Playing a loot card from the hand
type lootCardEvent struct {
	l lootCard
	f lootCardEffect
}

func (d lootCardEvent) eHolder() {}

// When a monster dies, the player receives the rewards listed on the monster's card
// No card or action can respond to this event
type monsterRewardEvent struct {
	r cardEffect
}

func (m monsterRewardEvent) eHolder() {}

// Pay a cost to use an itemCard (must pay the cost in order to place the itemCard on the stack).
type paidItemEvent struct {
	t treasureCard
	f cardEffect
}

func (d paidItemEvent) eHolder() {}

type startOfTurnEvent struct{}

func (s startOfTurnEvent) eHolder() {}

// Any cardEffect that has triggered during play
// A passive itemCard, monster / monster death, or trinket
type triggeredEffectEvent struct {
	c card
	f cardEffect
}

func (d triggeredEffectEvent) eHolder() {}

type event struct {
	p    *player     // The ORIGINAL player that pushed the effect on the stack
	e    eventHolder // Can be any of the events listed above.
	roll uint8       // Holds the dice roll value for the event
}

// TODO centralize all pushes to the stack and implement passive effect resolvement!
//func (b *Board) pushEventToStack(p *player, eh eventHolder) error {
//	event := event{p: p, e: eh}
//	switch eh.(type) {
//	case intentionToAttackEvent:
//		m := eh.(intentionToAttackEvent)
//
//	}
//}

// pop an event off of the event stack
// and alter the game's state based off the
// action pushed.

// TODO: search for monster's id before resolving monster damage. If in play, inflict damage. Else, don't.
func (b *Board) resolve() error {
	var err error = errors.New("no deckNode on stack")
	var node *eventNode
	es := &b.eventStack
	node = es.pop()
	if node != nil {
		p, ev, roll := node.event.p, node.event.e, node.event.roll
		switch ev.(type) {
		case activateEvent:
			ev.(activateEvent).f(roll)
		case damageEvent:
			e := ev.(damageEvent)
			if _, ok := e.target.(*monsterCard); ok {
				if _, monster := b.monster.getActiveMonster(e.target.getId()); monster != nil && !e.target.isDead() {
					monster.decreaseHP(e.n)
					if monster.isDead() {
						b.killMonster(p, monster)
					}
				}
			} else { // character value
				if !e.target.isDead() {
					if !dryBabyFunc(p) { // if not dry baby, proceed with normal calculation
						e.target.decreaseHP(e.n)
					}
					if !e.target.isDead() {
						p.damageRequiredEffects(node.next)
					} else {
						b.eventStack.push(event{p: p, e: deathOfCharacterEvent{}})
					}
				}
			}
		case deathOfCharacterEvent:
			p.death(b)
		case declareAttackEvent:
			e := ev.(declareAttackEvent)
			b.battle(p, e.m, roll)
		case declarePurchaseEvent:
			e := ev.(declarePurchaseEvent)
			err = b.treasure.buyFromShop(p, e.i)
		case diceRollEvent:
			e := ev.(diceRollEvent)
			b.eventStack.peek().event.roll = e.n // safe to do this. dice rolls are not isolated events
			if len(b.treasure.crystalBallGuess) > 0 {
				b.treasure.checkCrystalBall(e.n, b.loot)
			}
		case intentionToAttackEvent:
			p.inBattle = true
			ev.(intentionToAttackEvent).m.inBattle = true
		case lootCardEvent:
			e := ev.(lootCardEvent)
			_, usedBC := p.activeEffects[blankCard]
			e.f(roll, usedBC)
			if usedBC {
				delete(p.activeEffects, blankCard)
			}
		case paidItemEvent:
			e := ev.(paidItemEvent)
			e.f(roll)
		case triggeredEffectEvent:
			e := ev.(triggeredEffectEvent)
			e.f(roll)
		}
		//triggeredEvents := b.eventDependentPassiveActivation(p, deckNode)
	}
	return err
}

func (t *tArea) checkCrystalBall(roll uint8, l *lArea) {
	for player, guess := range t.crystalBallGuess {
		if guess == roll {
			player.loot(l)
			player.loot(l)
			player.loot(l)
		}
	}
	t.crystalBallGuess = make(map[*player]uint8, 3)
}
