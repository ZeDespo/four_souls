package four_souls

import (
	"errors"
	"fmt"
)

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
type declarePurchaseEvent struct{}

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

// Any cardEffect that has tapped during play
// A passive itemCard, monster / monster deathPenalty, or trinket
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

func (b *Board) checkActiveMonsterPassives(en *eventNode) []event {
	events := make([]event, 0)
	for _, am := range b.monster.getActiveMonsters() {
		events = append(events, am.trigger(en.event.p, b, en)...)
	}
	return events
}

func (b *Board) checkCursePassives(en *eventNode) []event {
	events := make([]event, 0)
	p := en.event.p
	for _, curse := range p.Curses {
		events = append(events, curse.trigger(p, b, en)...)
	}
	return events
}

func (b *Board) checkPlayerPassives(en *eventNode, startOrEndOfTurn bool) []event {
	events := make([]event, 0)
	if !startOrEndOfTurn {
		for _, p := range b.getPlayers(false) {
			for i := range p.getPassiveItems(true) {
				events = append(events, p.PassiveItems[i].trigger(p, b, en)...)
			}
		}
	} else {
		p := b.getActivePlayer()
		for i := range p.getPassiveItems(true) {
			events = append(events, p.PassiveItems[i].trigger(p, b, en)...)
		}
	}
	return events
}

func (b *Board) resolveNextEvent() error {
	err := errors.New("no eventNode on stack")
	es := &b.eventStack
	triggeredEvents := make([]event, 0)
	node := es.pop()
	if node != nil {
		p, ev, roll := node.event.p, node.event.e, node.event.roll
		switch ev.(type) {
		case activateEvent: // Regardless of Treasure card or character
			eh := ev.(activateEvent)
			eh.f(roll)
			if _, ok := eh.c.(*treasureCard); ok {
				if i, err := p.getItemIndex(shinyRock, true); err == nil {
					triggeredEvents = append(triggeredEvents, p.PassiveItems[i].trigger(p, b, node)...)
				}
			}
		case damageEvent:
			e := ev.(damageEvent)
			if _, ok := e.target.(*monsterCard); ok {
				if _, monster := b.monster.getActiveMonster(e.target.getId()); monster != nil && !e.target.isDead() {
					monster.decreaseHP(e.n)
					if monster.isDead() {
						b.killMonster(p, monster.id)
					}
				}
			} else { // character value
				if !e.target.isDead() {
					if !dryBabyFunc(p) { // if not dry baby, proceed with normal calculation
						e.target.decreaseHP(e.n)
					}
					if !e.target.isDead() {
						p.checkDamageRequiredEffects(node.next)
					} else {
						b.killPlayer(p)
					}
				}
			}
			triggeredEvents = append(triggeredEvents, b.checkActiveMonsterPassives(node)...)
			triggeredEvents = append(triggeredEvents, b.checkPlayerPassives(node, false)...)
		case deathOfCharacterEvent:
			p.deathPenalty(b)
			triggeredEvents = append(triggeredEvents, b.checkPlayerPassives(node, false)...)
		case declareAttackEvent:
			b.battle(p, ev.(declareAttackEvent).m, roll)
			triggeredEvents = append(triggeredEvents, b.checkActiveMonsterPassives(node)...)
			triggeredEvents = append(triggeredEvents, b.checkPlayerPassives(node, false)...)
		case declarePurchaseEvent:
			showTreasureCards(b.treasure.zones, "shop", 0)
			err = b.treasure.buyFromShop(p, uint8(readInput(0, len(b.treasure.zones)-1)))
		case diceRollEvent:
			e := ev.(diceRollEvent)
			b.eventStack.peek().event.roll = e.n // safe to do this. dice rolls are not isolated events
			if len(b.treasure.crystalBallGuess) > 0 {
				b.treasure.checkCrystalBall(e.n, b.loot)
			}
			triggeredEvents = append(triggeredEvents, b.checkActiveMonsterPassives(node)...)
			triggeredEvents = append(triggeredEvents, b.checkPlayerPassives(node, false)...)
		case endTurnEvent:
			for _, p := range b.getPlayers(true) {
				p.resetStats(p.isActivePlayer(b))
			}
			triggeredEvents = append(triggeredEvents, b.checkPlayerPassives(node, true)...)
			triggeredEvents = append(triggeredEvents, b.checkCursePassives(node)...)
			if !checkActiveEffects(p.activeEffects, theSun, true) { // Player gains extra turn
				b.api = (b.api + 1) % uint8(len(b.players))
			} else {
				if checkActiveEffects(b.players[b.api].activeEffects, famine, true) { // Skip that player's turn
					b.api = (b.api + 1) % uint8(len(b.players))
				}
			}
		case intentionToAttackEvent:
			e := ev.(intentionToAttackEvent)
			p.numAttacks -= 1
			if e.m != nil { // Attack an actual monster
				p.inBattle, e.m.inBattle = true, true
				triggeredEvents = append(triggeredEvents, b.checkActiveMonsterPassives(node)...)
				if i, err := p.getItemIndex(babyHaunt, true); err == nil {
					triggeredEvents = append(triggeredEvents, p.PassiveItems[i].trigger(p, b, node)...)
				}
			} else { // Attack the monster deck. May or may not be a monster
				m := b.monster.draw()
				m.showCard(0)
				if !m.isBonusCard() {
					p.inBattle, m.inBattle = true, true
					showMonsterCards(b.monster.getActiveMonsters(), 0)
					fmt.Print("Overlay over which monster?")
					b.monster.zones[readInput(0, len(b.monster.zones))].push(m)
				} else {
					err = m.activate(&b.players[b.api], b)
				}
			}
		case intentionToPurchaseEvent:
			triggeredEvents = append(triggeredEvents, event{p: p, e: declarePurchaseEvent{}})
		case lootCardEvent:
			e := ev.(lootCardEvent)
			_, usedBC := p.activeEffects[blankCard]
			e.f(roll, usedBC)
			if usedBC {
				delete(p.activeEffects, blankCard)
			}
		case monsterRewardEvent:
			e := ev.(monsterRewardEvent)
			e.r(roll)
		case paidItemEvent:
			e := ev.(paidItemEvent)
			e.f(roll)
		case startOfTurnEvent:
			p.Character.tapped = false
			for _, ai := range p.getActiveItems(true) {
				ai.recharge()
			}
			p.loot(b.loot)
			firstTimeIds := map[uint16]struct{}{curvedHorn: {}, bumbo: {}, championBelt: {}, theHabit: {}, polydactyly: {}}
			for _, c := range p.PassiveItems {
				id := c.getId()
				if _, ok := firstTimeIds[id]; ok {
					if (id == bumbo && c.getCounters() > 0) || id != bumbo {
						p.activeEffects[id] = struct{}{}
					}
				}
			}
			triggeredEvents = append(triggeredEvents, b.checkPlayerPassives(node, true)...)
			triggeredEvents = append(triggeredEvents, b.checkCursePassives(node)...)
		case triggeredEffectEvent:
			e := ev.(triggeredEffectEvent)
			e.f(roll)
		}
		var i, max uint8 = 0, uint8(len(triggeredEvents))
		for i = 0; i < max; i++ {
			b.eventStack.push(triggeredEvents[i])
			actionReactionChecker(triggeredEvents[i].p, b)
		}
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
