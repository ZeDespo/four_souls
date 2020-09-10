package four_souls

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"
)

// The main type that the game revolves around. Holds all major variables in one struct
type Board struct {
	players    []player   // All the players for the game.
	loot       *lArea     // Loot deck and discard pile
	monster    *mArea     // Monster deck, discard pile, and zones
	treasure   *tArea     // Treasure deck, discard pile, and zones
	eventStack eventStack // stack that will hold most state changing events, except monster deaths and rewards.
	api        uint8      // Active Player Index: the index of the active player in players
}

type actionReaction struct {
	msg   string
	value uint8
}

// The area of the board designate for loot cards
type lArea struct {
	deck, discardPile deck
	activeEffects     map[uint16]struct{}
}

// The area of the board designated for battle / monster cards and their zones.
type mArea struct {
	deck, discardPile deck
	zones             []activeSlot // active monsters will be on top of the stack. Overlayed monsters beneath them
	theMidasTouch     map[*player]struct{}
}

// Type representing the player's board: their character, all items they control, money, souls, and their hand
type player struct {
	Character         characterCard
	ActiveItems       []treasureCard
	PassiveItems      []passiveItem
	Pennies           int8
	Souls             []card // Character value (The Lost), tArea Cards, mArea Cards, Lost Soul loot value
	Curses            []monsterCard
	Hand              []lootCard
	baseNumLootPlayed int8
	baseNumPurchases  int8
	baseNumAttacks    int8
	numLootPlayed     int8 // The number of times we could play a loot card.
	numPurchases      int8 // The number of time we could numPurchases an itemCard
	numAttacks        int8 // The number of times a new monster can be attacked.
	inBattle          bool // Is the unit in combat or not?
	forceAttack       bool
	forceAttackTarget map[int8]uint8 // Key = the attack target. Value = the number of attacks forced
	forceEnd          bool           // Death, effects like Holy Card and The Beginning, can force an end to a turn.
	activeEffects     map[uint16]struct{}
}

// The area of the board designated for the shop / treasure cards
type tArea struct {
	deck, discardPile deck
	zones             []treasureCard
	activeEffects     map[uint16]struct{}
	crystalBallGuess  map[*player]uint8 // The guess of someone who used the Crystal Ball .
}

// Add a loot card (trinket), treasure card (active / passive) or a monster card (curse)
// on the player's board.
// For active and passive items, this method will sort the items to enable
// binary search of items
func (p *player) addCardToBoard(c card) {
	switch c.(type) {
	case treasureCard:
		t := c.(treasureCard)
		if t.passive {
			p.PassiveItems = p.addPassiveItem(&t, p.PassiveItems)
		} else {
			p.ActiveItems = p.addNonPassiveItem(t, p.ActiveItems)
		}
	case *treasureCard:
		t := c.(*treasureCard)
		if t.passive {
			p.PassiveItems = p.addPassiveItem(t, p.PassiveItems)
		} else {
			p.ActiveItems = p.addNonPassiveItem(*t, p.ActiveItems)
		}
	case lootCard: // A trinket derived from the loot value deck.
		lc, _ := c.(lootCard)
		if lc.trinket == false {
			panic("not a value that can be added to the board")
		} else {
			p.addPassiveItem(lc, p.PassiveItems)
		}
	case monsterCard:
		curses := map[uint16]struct{}{curseOfAmnesia: {}, curseOfGreed: {}, curseOfLoss: {}, curseOfPain: {},
			curseOfTheBlind: {}, curseOfFatigue: {}, curseOfTinyHands: {}, curseOfBloodLust: {}, curseOfImpulse: {}}
		mc := c.(monsterCard)
		if _, ok := curses[mc.id]; ok {
			p.Curses = append(p.Curses, mc)
		}
	default:
		panic("not a valid card to add to the board")
	}
}

func (b *Board) addMonsterToZone(i uint8) {
	ap, m := &b.players[b.api], b.monster
	card := m.draw()
	if card.isBonusCard() {
		if f, special, err := card.f(ap, b, card); err == nil && f != nil {
			b.eventStack.push(event{p: ap, e: triggeredEffectEvent{c: card, f: f}})
			if special {
				b.rollDiceAndPush()
			}
		}
	} else {
		card.resetStats()
		m.zones[i].push(card)
	}
}

// Helper function for adding an item card to the player's board.
// This particular function will add an Active / Paid Item to the
// player's board in a sorted manner.
func (p *player) addNonPassiveItem(c treasureCard, dest []treasureCard) []treasureCard {
	i := sort.Search(len(dest), func(i int) bool { return dest[i].id >= c.id })
	dest = append(dest, treasureCard{})
	copy(dest[i+1:], dest[i:])
	dest[i] = c
	return dest
}

// Helper function for adding an item card to the player's board.
// This particular function will add a Passive Item to the
// player's board in a sorted manner.
func (p *player) addPassiveItem(c passiveItem, dest []passiveItem) []passiveItem {
	i := sort.Search(len(dest), func(i int) bool { return dest[i].getId() >= c.getId() })
	dest = append(dest, &treasureCard{})
	copy(dest[i+1:], dest[i:])
	dest[i] = c
	return dest
}

// Adds a card to the player's soul slice. If the player reached the number of souls to win,
// return true. Else, return false
func (p *player) addSoulToBoard(c card) bool {
	p.Souls = append(p.Souls, c)
	var victory bool
	if len(p.Souls) == 4 {
		victory = true
	}
	return victory
}

func (b *Board) battle(p *player, m *monsterCard, roll uint8) {
	if roll >= m.roll { // successful hit
		attack := p.Character.ap
		if checkActiveEffects(p.activeEffects, curvedHorn, true) {
			attack += 1
		}
		if checkActiveEffects(p.activeEffects, championBelt, true) {
			attack += 1
		}
		if checkActiveEffects(p.activeEffects, polydactyly, true) {
			attack += 1
		}
		if emptyVesselChecker(p, true) {
			attack += 1
		}
		b.damagePlayerToMonster(p, m, attack, 0)
	} else { // missed
		damage := m.ap
		if horfChecker(m.id, roll) {
			damage += 1
		} else if leaperChecker(m.id, roll) || momChecker(m.id, roll) {
			damage *= 2
		}
		b.damageMonsterToPlayer(m, p, damage, roll)
	}
}

func (p *player) beforePayingPenalties(b *Board) {
	for i := range p.Curses {
		b.discard(p.popCurse(uint8(i)))
	}
	for _, c := range p.getPassiveItems(false) {
		switch c.getId() {
		case babyHaunt:
			b.hauntGiveAwayHelper(p, c)
		case daddyHaunt:
			b.hauntGiveAwayHelper(p, c)
		case mamaHaunt:
			b.hauntGiveAwayHelper(p, c)
		}
	}
}

// Buy an itemCard from either the treasure zone or the top of the deck.
// There are several items that will influence the purchasing process
// Credit Card (Loot Item): Makes the cost of a single itemCard purchased 0
func (t *tArea) buyFromShop(p *player, idx uint8) error {
	var err = errors.New("not enough money to buy")
	tCard := t.zones[idx]
	var cost = steamySaleFunc(p)
	if _, ok := t.activeEffects[creditCard]; ok {
		cost = 0
		delete(t.activeEffects, creditCard)
	}
	if p.Pennies >= cost {
		p.loseCents(cost)
		p.addCardToBoard(&tCard)
		err = nil
	}
	t.zones[idx] = treasureCard{}
	return err
}

// Add an itemCard's id to the player's active effects map to
// instruct any card that requires an activation cost of n damage to a player
// that the cost was met successfully.
// Ex: damage(1) -> temperance(1): if character is alive, cost met successfully. Gain 4 cents.
// Ex: damage (1) -> guppysPaw(1): if character is alive, prevent up to 2 damage on the stack to a player.
// The param nexToResolve is the next deckNode to be popped off the stack following the damage.
func (p *player) checkDamageRequiredEffects(nextToResolve *eventNode) {
	if p.Character.hp > 0 {
		if nextToResolve != nil {
			switch nextToResolve.event.e.(type) {
			case lootCardEvent:
				e := nextToResolve.event.e.(lootCardEvent)
				if e.l.id == temperance {
					p.activeEffects[temperance] = struct{}{}
				}
			case activateEvent:
				e := nextToResolve.event.e.(activateEvent)
				if e.c.getId() == guppysPaw {
					p.activeEffects[guppysPaw] = struct{}{}
				}
			}
		}
	}
}

// This method should only be called after all events on the
// event stack have resolved
// Check the field for the following things:
// 0) Check for a victory
// 1) If there are unfilled shop zones, fill them up with items from the
// top of the treasure deck
// 2) If there are unfilled monster zones, draw cards until the zone is filled.
func (b *Board) checkTheField() []player {
	if b.eventStack.size == 0 {
		victors := checkVictory(b.players)
		if len(victors) > 0 {
			return victors
		}
		for i := range b.treasure.zones {
			if b.treasure.zones[i].id == 0 { // No treasure value here
				b.treasure.zones[i] = b.treasure.draw()
			}
		}
		for i, _ := range b.monster.zones {
			for len(b.monster.zones[i]) == 0 {
				b.addMonsterToZone(uint8(i))
			}
		}
	}
	return []player{}
}

// A monster has inflicted damage to a player via a missed attack / other effect.
// Push the damage to the stack appropriately
func (b *Board) damageMonsterToPlayer(m *monsterCard, target *player, n uint8, combatRoll uint8) {
	if !target.isDead() {
		b.eventStack.push(event{p: target, e: damageEvent{monster: m, target: target, n: n}, roll: combatRoll})
		b.preventDamageHelper(target, b.eventStack.peek())
	}
}

// Helper to any action / effect that damages a monster.
// Search the zone to make sure the monster is in play, then push damage to the monster.
func (b *Board) damagePlayerToMonster(p *player, target *monsterCard, n uint8, combatRoll uint8) {
	if !target.isDead() {
		if (target.id == carrionQueen && combatRoll != 6) || (target.id == pin && combatRoll == 6) {
			n = 0
		}
		b.eventStack.push(event{p: p, e: damageEvent{target: target, n: n}, roll: combatRoll})
		if target.id == theDukeOfFlies {
			b.eventStack.push(event{p: p, e: triggeredEffectEvent{c: target, f: theDukeOfFliesEvent(b.eventStack.peek())}})
			b.rollDiceAndPush()
		}
	}
}

// Some player's action / card effect damaged another player.
// p *player: the player that pushed the event to the stack, if applicable (monster attack)
// target *player: the subject of the damage
// n uint8: how much damage to inflict
func (b *Board) damagePlayerToPlayer(p, target *player, n uint8) {
	if !target.isDead() { // Not dead
		b.eventStack.push(event{p: p, e: damageEvent{target: target, n: n}})
		b.preventDamageHelper(p, b.eventStack.peek())
	}
}

// The player died and no card effect was able to stop it.
// Resolve effects that occur before paying penalties, then pay the deathPenalty penalty
// 1) Discard one card
// 2) Lose 1 cent
// 3) Destroy one item
// 4) Deactivate all items and character card
func (p *player) deathPenalty(b *Board) {
	p.beforePayingPenalties(b)
	shadowActivated := shadowFunc(p, b)
	if !shadowActivated { // The shadow is not in play. Resume deathPenalty normally
		showLootCards(p.Hand, p.Character.name, 0)
		fmt.Println("Discard one card.")
		b.discard(p.popHandCard(uint8(readInput(0, len(p.Hand)-1))))
		p.loseCents(1)
	}
	for _, c := range p.getActiveItems(true) {
		if c.active {
			c.tapped = true
		}
	}
}

func deathPlayerPrevention(id uint16, p *player, b *Board, en *eventNode) {
	var f cardEffect
	var err error
	var i uint8
	if _, ok := en.event.e.(deathOfCharacterEvent); ok { // No need to perform if event's already fizzled
		if i, err = p.getItemIndex(id, true); err == nil && (id == brokenAnkh || id == guppysCollar) {
			f = func(roll uint8) {
				if (id == brokenAnkh && roll == 6) || (id == guppysCollar && roll >= 1 && roll <= 3) {
					en.event.e = fizzledEvent{}
					if ok := p.isActivePlayer(b); ok {
						b.forceEndOfTurn()
					}
				}
			}
			b.eventStack.push(event{p: p, e: triggeredEffectEvent{c: p.PassiveItems[i], f: f}})
			b.rollDiceAndPush()
		}
	}
}

func (p *player) destroyItem(b *Board, ic itemCard) {
	id, isPassive := ic.getId(), ic.isPassive()
	if i, err := p.getItemIndex(id, isPassive); err == nil {
		p.destroyItemByIndex(b, i, isPassive)
	}
}

func (p *player) destroyItemByIndex(b *Board, i uint8, isPassive bool) {
	ic := p.popItemByIndex(i, isPassive)
	if f := ic.getContinuousPassive(); f != nil {
		f(p, b, ic, true)
		if ic.getId() != theChest { // Does not go in the discard pile
			b.discard(ic)
		}
	}
}

func (b *Board) discard(c card) {
	switch c.(type) {
	case lootCard:
		b.loot.discard(c.(lootCard))
	case monsterCard:
		m := c.(monsterCard)
		b.monster.discard(&m)
	case *monsterCard:
		m := c.(*monsterCard)
		b.monster.discard(m)
	case treasureCard:
		card := c.(treasureCard)
		b.treasure.discard(&card)
	case *treasureCard:
		b.treasure.discard(c.(*treasureCard))
	}
}

func (l *lArea) discard(lc lootCard) {
	l.discardPile = append(l.discardPile, lc)
}

func (m *mArea) discard(mc *monsterCard) {
	mc.resetStats()
	m.discardPile = append(m.discardPile, *mc)
}

func (t *tArea) discard(tc *treasureCard) {
	tc.counters = 0
	tc.tapped = false
	t.discardPile = append(t.discardPile, *tc)
}

func (l *lArea) draw() lootCard {
	card, err := l.deck.pop()
	check(err)
	return card.(lootCard)
}

func (m *mArea) draw() monsterCard {
	card, err := m.deck.pop()
	check(err)
	return card.(monsterCard)
}

func (t *tArea) draw() treasureCard {
	card, err := t.deck.pop()
	check(err)
	return card.(treasureCard)
}

// Resolve end of turn passive effects, be rid of any "until end of turn" effects
// that would otherwise not be resolved by the reset method.
func (b *Board) endPhase() {
	for _, p := range b.players {
		checkActiveEffects(p.activeEffects, twoOfClubs, true)
		if checkActiveEffects(p.activeEffects, diplopia, true) {
			j, err := p.getItemIndex(diplopia, true)
			if err == nil {
				tc := treasureCard{baseCard: baseCard{name: "Diplopia", id: diplopia}, active: true, f: diplopiaFunc}
				p.popPassiveItem(j)
				p.addCardToBoard(&tc)
			}
		}
	}
}

func (p *player) forceAnAttack(addAttack bool) {
	p.forceAttack = true
	if addAttack {
		p.numAttacks += 1
	}
}

// Force the end of the active player's turn, and pop
// any unresolved actions off the event stack.
// Cards like The Fool and Holy Card will call this method.
func (b *Board) forceEndOfTurn() {
	b.players[b.api].forceEnd = true
	for b.eventStack.head != nil {
		_ = b.eventStack.pop()
	}
}

func (p *player) gainCents(n int8) {
	if i, err := p.getItemIndex(bumbo, true); err == nil { // Do not gain cents! Put counters instead
		p.bumboAddCounterHelper(p.PassiveItems[i].(*treasureCard), n)
	} else {
		p.Pennies += n
		counterfeitPennyChecker(p)
	}
}

func (p player) isActivePlayer(b *Board) bool {
	var ok bool
	if p.Character.id == b.players[b.api].Character.id {
		ok = true
	}
	return ok
}

func (b *Board) killMonster(p *player, mId uint16) {
	if i, err := b.monster.getActiveMonster(mId); err == nil {
		m := b.monster.zones[i].pop()
		m.resetStats()
		if m.f != nil {
			if f, _, err := m.f(p, b, m); err == nil { // on deathPenalty trigger
				b.eventStack.push(event{p: p, e: triggeredEffectEvent{c: m, f: f}})
				if mId == ragman || mId == wrath {
					b.rollDiceAndPush()
				}
			}
		}
		if m.isBoss {
			b.players[b.api].addSoulToBoard(m)
			if mId == theHaunt {
				checkActiveEffects(b.players[b.api].activeEffects, theHaunt, true)
			}
		} else if m.isBoss && mId == delirium {
			deliriumDeathHandler(m, b.monster)
		} else {
			b.discard(m)
		}
		theMidasTouchHelper(b.monster)
		rf, rollRequired := m.rf(b)
		b.eventStack.push(event{p: p, e: monsterRewardEvent{r: rf}})
		if rollRequired {
			b.rollDiceAndPush()
		}
		b.killMonster(p, stoney)
		b.killMonster(p, deathsHead)
	}
}

// Method that pushes the deathPenalty event to the stack for a player
// Here's the player's passive opportunity to prevent deathPenalty and end his / her turn
func (b *Board) killPlayer(target *player) {
	if !target.isDead() {
		b.eventStack.push(event{p: target, e: deathOfCharacterEvent{}})
		deathNode := b.eventStack.peek()
		deathPrevention := [2]uint16{brokenAnkh, guppysCollar}
		for _, id := range deathPrevention {
			deathPlayerPrevention(id, target, b, deathNode)
		}
	}
}

func (p *player) loot(l *lArea) {
	if checkActiveEffects(l.activeEffects, compost, true) { // Draw from top of discard pile.
		if dC, err := l.discardPile.pop(); err == nil {
			p.Hand = append(p.Hand, dC.(lootCard))
		}
	} else { // Standard draw
		p.Hand = append(p.Hand, l.draw())

	}
	if checkActiveEffects(p.activeEffects, twoOfClubs, false) { // Draw an additional card while active.
		p.Hand = append(p.Hand, l.draw())
	}
}

func (p *player) loseCents(n int8) {
	x := p.Pennies - n
	if x < 0 {
		p.Pennies = 0
	} else {
		p.Pennies = x
	}
}

// return: Whether the player made an action or decided to pass
func (p *player) makeChoice(b *Board) bool {
	didSomething := true
	actions := p.getPlayerActions(p.isActivePlayer(b), b.eventStack.isEmpty())
	for i, a := range actions {
		fmt.Println(i, ") ", a.msg)
	}
	fmt.Println("What would", p.Character.name, "like to do?")
	switch actions[readInput(0, len(actions))].value {
	case playLootCard:
		showLootCards(p.Hand, p.Character.name, 0)
		fmt.Println("Play which card?")
		handCard := p.Hand[readInput(0, len(p.Hand)-1)]
		err := handCard.activate(p, b)
		if err != nil {
			fmt.Println(fmt.Sprintf("Could not activate %s:\n%s.", handCard.name, err))
		}
	case buyItem:
		b.eventStack.push(event{p: p, e: intentionToPurchaseEvent{}})
	case attackMonster:
		if p.inBattle && p.isActivePlayer(b) {
			monsters := b.monster.getActiveMonsters()
			l := len(monsters)
			showMonsterCards(monsters, 0)
			fmt.Println(fmt.Sprintf("%d) Monster Deck\nWhich target to attack?", l))
			i := readInput(0, l)
			var m *monsterCard
			if i < l {
				m = monsters[i]
			} else {
				m = nil // Attack the deck instead
			}
			b.eventStack.push(event{p: p, e: intentionToAttackEvent{m: m}})
		} else {
			b.eventStack.push(event{p: p, e: declareAttackEvent{m: b.monster.getActiveMonsterInBattle()}})
			b.rollDiceAndPush()
		}
	case activateCharacter:
		err := p.Character.activate(p, b)
		if err != nil {
			fmt.Println(err)
		}
	case activateItem:
		items := p.getUsableActiveItems()
		l := len(items)
		if l > 0 {
			showTreasureCards(items, p.Character.name, 0)
			fmt.Println("Which card to activate?")
			err := items[readInput(0, l-1)].activate(p, b)
			if err != nil {
				fmt.Println(err)
			}
		}
	case doNothing:
		didSomething = false
	}
	return didSomething
}

func (b *Board) placeInDeck(c card, onTop bool) {
	switch c.(type) {
	case lootCard:
		b.loot.placeInDeck(c.(lootCard), onTop)
	case monsterCard:
		b.monster.placeInDeck(c.(monsterCard), onTop)
	case treasureCard:
		b.treasure.placeInDeck(c.(treasureCard), onTop)
	}
}

func (l *lArea) placeInDeck(lc lootCard, onTop bool) {
	if onTop {
		l.deck.prepend(lc)
	} else {
		l.deck.append(lc)
	}
}

func (m *mArea) placeInDeck(mc monsterCard, onTop bool) {
	if onTop {
		m.deck.prepend(mc)
	} else {
		m.deck.append(mc)
	}
}

func (t *tArea) placeInDeck(tc treasureCard, onTop bool) {
	if onTop {
		t.deck.prepend(tc)
	} else {
		t.deck.append(tc)
	}
}

func (p *player) popActiveItem(idx uint8) itemCard {
	length := len(p.ActiveItems)
	var card itemCard = &p.ActiveItems[idx]
	if length == 0 {
		panic("No active items to destroy!")
	} else if length == 1 && idx == 0 { // deleting only or last element in slice
		p.ActiveItems = p.ActiveItems[:idx]
	} else { // Must preserve order of these slices so middle deletion doesn't screw up order of elements proceeding it
		copy(p.ActiveItems[idx:], p.ActiveItems[idx+1:])
		p.ActiveItems[length-1] = treasureCard{}
		p.ActiveItems = p.ActiveItems[:length-1]
	}
	return card
}

func (l *lArea) popCardFromDiscardPile(idx uint8) lootCard {
	card, err := l.discardPile.popByIndex(idx)
	check(err)
	return card.(lootCard)
}

func (m *mArea) popCardFromDiscardPile(idx uint8) monsterCard {
	card, err := m.discardPile.popByIndex(idx)
	check(err)
	return card.(monsterCard)
}

func (t *tArea) popCardFromDiscardPile(idx uint8) treasureCard {
	card, err := t.discardPile.popByIndex(idx)
	check(err)
	return card.(treasureCard)
}

func (p *player) popCurse(idx uint8) monsterCard {
	length := len(p.Curses)
	card := p.Curses[idx]
	if length == 0 {
		panic("No curses to destroy!")
	} else if length == 1 && idx == 0 { // deleting only or last element in slice
		p.Curses = p.Curses[:idx]
	} else { // Must preserve order of these slices so middle deletion doesn't screw up order of elements proceeding it
		copy(p.Curses[idx:], p.Curses[idx+1:])
		p.Curses[length-1] = monsterCard{}
		p.Curses = p.Curses[:length-1]
	}
	return card
}

func (p *player) popHandCard(idx uint8) lootCard {
	length := len(p.Hand)
	card := p.Hand[idx]
	if length == 1 && idx == 0 { // deleting only or last element in slice
		p.Hand = p.Hand[:idx]
	} else { // Must preserve order of these slices so middle deletion doesn't screw up order of elements proceeding it
		copy(p.Hand[idx:], p.Hand[idx+1:])
		p.Hand[length-1] = lootCard{}
		p.Hand = p.Hand[:length-1]
	}
	return card
}

func (p *player) popItem(ic itemCard) (itemCard, error) {
	var c itemCard
	var i uint8
	var err error
	if i, err = p.getItemIndex(ic.getId(), ic.isPassive()); err == nil {
		c = p.popItemByIndex(i, ic.isPassive())
	}
	return c, err
}

func (p *player) popItemByIndex(idx uint8, isPassive bool) itemCard {
	var c itemCard
	if !isPassive {
		c = p.popActiveItem(idx)
	} else {
		c = p.popPassiveItem(idx)
	}
	return c
}

func (p *player) popPassiveItem(idx uint8) passiveItem {
	length := len(p.PassiveItems)
	card := p.PassiveItems[idx]
	if length == 0 {
		panic("No passive Items to destroy!")
	} else if length == 1 && idx == 0 { // deleting only or last element in slice
		p.PassiveItems = p.PassiveItems[:idx]
	} else { // Must preserve order of these slices so middle deletion doesn't screw up order of elements proceeding it
		copy(p.PassiveItems[idx:], p.PassiveItems[idx+1:])
		p.PassiveItems[length-1] = nil
		p.PassiveItems = p.PassiveItems[:length-1]
	}
	return card
}

func (p *player) popSoul(idx uint8) card {
	length := len(p.Souls)
	card := p.Souls[idx]
	if length == 1 && idx == 0 { // deleting only or last element in slice
		p.Souls = p.Souls[:idx]
	} else if length > 1 { // Must preserve order of these slices so middle deletion doesn't screw up order of elements proceeding it
		copy(p.Souls[idx:], p.Souls[idx+1:])
		p.Souls[length-1] = nil
		p.Souls = p.Souls[:length-1]
	}
	return card
}

func (p *player) rechargeActiveItemById(tId uint16) {
	if i, err := p.getItemIndex(tId, false); err == nil {
		p.ActiveItems[i].recharge()
	}
}

func (mc *monsterCard) resetStats() {
	mc.hp, mc.ap, mc.roll, mc.inBattle = mc.baseHealth, mc.baseAttack, mc.baseRoll, false
}

// Set the player's and character's values back to their base values,
// free of any buffs or nerfs to any stats.
func (p *player) resetStats(isActivePlayer bool) {
	p.forceEnd = false
	p.numAttacks, p.numPurchases, p.numLootPlayed = p.baseNumAttacks, p.baseNumPurchases, p.baseNumLootPlayed
	c := &p.Character
	c.tapped = false
	c.hp, c.ap = c.baseHealth, c.baseAttack
	delete(p.activeEffects, theEmpress)
	if isActivePlayer {
		delete(p.activeEffects, larryJr)
	}
}

func (b *Board) rollDice() (diceRollEvent, *player) {
	if node := b.eventStack.peek(); node != nil {
		nextEvent := node.next.event.e
		var p *player = node.event.p
		roll := uint8(rand.Intn(6) + 1)
		if checkActiveEffects(p.activeEffects, theEmpress, false) {
			modifyDiceRoll(&roll, 1)
		}
		if checkActiveEffects(p.activeEffects, theHaunt, false) {
			modifyDiceRoll(&roll, -1)
		}
		if _, ok := nextEvent.(declareAttackEvent); ok {
			if checkActiveEffects(p.activeEffects, bumbo, true) {
				modifyDiceRoll(&roll, 2)
			}
			if emptyVesselChecker(p, false) {
				modifyDiceRoll(&roll, 1)
			}
			if _, err := p.getItemIndex(meat, true); err == nil {
				modifyDiceRoll(&roll, 1)
			}
			if _, err := p.getItemIndex(synthoil, true); err == nil {
				modifyDiceRoll(&roll, 1)
			}
		}
		return diceRollEvent{n: roll}, p
	} else {
		panic("dice rolls do not happen in isolation!")
	}
}

// Player does not need to be explicitly passed to this receiver.
// Dice rolls do not occur in an isolated state. They always follow some action.
func (b *Board) rollDiceAndPush() {
	roll, p := b.rollDice()
	event := event{p: p, e: roll}
	b.eventStack.push(event)
}

func (t *tArea) stealFromShop(id uint16) (treasureCard, error) {
	var tc treasureCard
	var err = errors.New("itemCard does not exist in the shop")
	for i, c := range t.zones {
		if c.id == id {
			tc, err = t.zones[i], nil
			t.zones[i] = treasureCard{baseCard: baseCard{id: 0}}
			break
		}
	}
	return tc, err
}

func (p *player) stealItem(id uint16, isPassive bool, p2 *player) {
	j, err := p2.getItemIndex(id, isPassive)
	if err == nil {
		p.addCardToBoard(p2.popItemByIndex(j, isPassive))
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Start a new game by doing the following:
// 1) Set up the decks and place them on the board
// 2) Initialize each player's characters
// 3) Give the players three loot cards for their hand
// 4) Place two monsters on the board's monster zone. All other cards will go on the bottom of the deck.
// 5) Place two treasure items on the board's treasure zone.
func NewGame(numPlayers uint8, useKickStarterExpansion bool, useFourSoulsExpansion bool) Board {
	rand.Seed(time.Now().UnixNano())
	lootDeck := getLootDeck(useKickStarterExpansion, useFourSoulsExpansion)
	monsterDeck := getMonsterDeck(useKickStarterExpansion, useFourSoulsExpansion)
	treasureDeck := getTreasureDeck(useKickStarterExpansion, useFourSoulsExpansion)
	board := Board{
		loot: &lArea{deck: lootDeck, discardPile: make(deck, 0, lootDeck.len())},
		monster: &mArea{deck: monsterDeck, discardPile: make(deck, 0, monsterDeck.len()),
			zones: make([]activeSlot, 2, 6)},
		treasure: &tArea{deck: treasureDeck, discardPile: make(deck, 0, treasureDeck.len()),
			zones: make([]treasureCard, 2, 4), crystalBallGuess: make(map[*player]uint8, 3)},
	}
	players := setPlayerBoards(numPlayers, useKickStarterExpansion, useFourSoulsExpansion)
	for i := range players {
		var j uint8
		for j = 0; j < 3; j++ {
			players[i].loot(board.loot)
		}
	}
	board.players = players
	var i uint8
	for i < 2 {
		m := board.monster.draw()
		if m.isBonusCard() {
			board.monster.placeInDeck(m, false)
		} else {
			m.resetStats()
			board.monster.zones[i].push(m)
			i += 1
		}
	}
	for i = 0; i < 2; i++ {
		board.treasure.zones[i] = board.treasure.draw()
	}
	return board
}

func (b *Board) attackMonsterDeck() {
	//ap := &b.players[b.api]
	//m := b.monster.draw()
	//fmt.Println("Drew ", m.name)
	//panic()
	//if !m.isBonusCard() {
	//
	//}

}

// Create a single player game that will test the game's mechanics in a single player state.
// This is to test the functionality of the game as a whole rather than the effects of cards.
// There's only one modification for the game's rules: if the player dies three times, the game ends.
// All of the outputs here will take place on the command line only.
func (b *Board) DebugGame() {
	numPlayers := uint8(len(b.players))
	victors := make([]player, 0, numPlayers)
	for len(victors) == 0 {
		ap := &b.players[b.api]
		if checkActiveEffects(ap.activeEffects, famine, true) { // skip the player's next turn
			b.api = (b.api + 1) % numPlayers
			continue
		}
		if ap.forceAttack && !ap.inBattle {
			if checkActiveEffects(ap.activeEffects, portal, true) {
				b.attackMonsterDeck()
			}
		}
		for target, numTimes := range ap.forceAttackTarget {
			numTimes -= 1
			if numTimes == 0 {
				delete(ap.forceAttackTarget, target)
			} else {
				ap.forceAttackTarget[target] = numTimes
			}
			if target == forceAttackDeck {
				b.attackMonsterDeck()
			} else {
				// TODO
			}
		}
		actions := ap.getPlayerActions(true, true)
		for i, a := range actions {
			fmt.Println(i, ") ", a.msg)
		}
		fmt.Println("What would you like to do?")
		switch actions[readInput(0, len(actions))].value {
		case playLootCard:
			fmt.Println()
		case buyItem:
			fmt.Println()
		}
		chainCards(ap, b)
		b.api = (b.api + 1) % numPlayers
	}
}

// This function is originally called either when the active player
// pushes an event to the event stack, or when some other card's effect
// pushes a new event to the stack in the middle of resolving.
// Check to see if other players have a response to some action.
// initiatingPlayer *player: the player whose pushed action is responsible
// for calling this function
// b *Board: the board struct
func actionReactionChecker(initiatingPlayer *player, b *Board) {
	keepGoing := true
	checkOrder := append(b.getOtherPlayers(initiatingPlayer, false), initiatingPlayer)
	ap := &b.players[b.api]
	for keepGoing {
		keepGoing = false
		for _, p := range checkOrder {
			if p.isActivePlayer(b) || (!p.isActivePlayer(b) && !trinityShieldFunc(ap)) {
				keepGoing = keepGoing || p.makeChoice(b)
			}
		}
	}
}

func checkActiveEffects(activeEffects map[uint16]struct{}, key uint16, deleteIfExists bool) bool {
	var activeEffect bool
	if _, ok := activeEffects[key]; ok {
		activeEffect = true
		if deleteIfExists { // Usually true for "just once" effects (Cancer, Pestilence, Credit Card)
			delete(activeEffects, key)
		}
	}
	return activeEffect
}

func checkVictory(players []player) []player {
	victors := make([]player, 0, len(players))
	twoSoulCards := map[uint16]struct{}{mom: {}, satan: {}, theLamb: {}, hush: {}, isaacMonster: {}, momsHeart: {}}
	for _, p := range players {
		var numSoulsToWin uint8 = 4
		if curseOfLossChecker(p) {
			numSoulsToWin += 1
		}
		var souls uint8
		for _, s := range p.Souls {
			if _, ok := twoSoulCards[s.getId()]; ok {
				souls += 2
			} else {
				souls += 1
			}
		}
		if souls == numSoulsToWin {
			victors = append(victors, p)
		}
	}
	return victors
}

// Set up the players by distributing characters and their respective
// starting items (with exception to Eden that gets a choice between the
// top three cards in the treasure deck).
func setPlayerBoards(numPlayers uint8, useKickstarterExpansion bool, useFourSoulsExpansion bool) []player {
	var characterDeck = getCharacters(numPlayers, useKickstarterExpansion, useFourSoulsExpansion)
	var startingItems = getStartingItems()
	var players = make([]player, numPlayers)
	var i uint8
	for i = 0; i < numPlayers; i++ {
		c := characterDeck[i]
		if c.name == "Eden" {
			panic("not yet implemented")
		}
		player := player{Character: c, Pennies: 3, Hand: make([]lootCard, 0, 10)}
		player.addCardToBoard(startingItems[c.name])
		if player.Character.name == "The Lost" {
			_ = player.addSoulToBoard(player.Character) // Impossible for a victory here. No need to check.
		}

		players[i] = player
	}
	return players
}
