package four_souls

import (
	"errors"
	"fmt"
)

// Bomb / Gold Bomb loot card helper.
// Check if the targeted monster is still on the field.
// If it is, push a 1 damage event to the event stack.
// m *mArea: The monster area of the board
// p *player: The player that activated Bomb!
// es *eventStack: The event stack
// mId uint16: The id of the monster card.
// n: The amount of damage done.
func (m *mArea) bombHelper(p *player, b *Board, monster *monsterCard, n uint8) lootCardEffect {
	return func(roll uint8, blankCard bool) {
		if blankCard {
			n *= 2
		}
		b.damagePlayerToMonster(p, monster, n, 0)
	}
}

// Bomb / Gold Bomb loot card helper.
// Check if the character is not dead.
// If it is not, push a 1 damage event to the event stack.
// p *player: The target of Bomb!
// ap *player: The player that activated Bomb!
// es *eventStack: The event stack
// n: The amount of damage done.
func (p *player) bombHelper(ap *player, b *Board, n uint8) lootCardEffect {
	return func(roll uint8, blankCard bool) {
		if blankCard {
			n *= 2
		}
		b.damagePlayerToPlayer(ap, p, n)
	}
}

// Helper for the Bum-Bo passive value
// Whenever Bum-Bo is on the field, the player who owns it puts counters
// on the value whenever they gain cents. This helper determines which effects
// are currently switched on

// p *player: the player who owns the card
// bumboTC *treasureCard: the pointer to the Bum-Bo treasure card
// n int8: The number of counters to add to the card
func (p *player) bumboAddCounterHelper(bumboTC *treasureCard, n int8) {
	beforeCounters := bumboTC.counters
	bumboTC.counters += n
	afterCounters := bumboTC.counters
	if beforeCounters < 1 && afterCounters >= 1 && !p.inBattle { // Add two to first attack roll if not in battle
		p.activeEffects[bumbo] = struct{}{}
	}
	if beforeCounters < 10 && afterCounters >= 10 {
		p.increaseAP(1)
		p.increaseBaseAttack(1)
	}
	if beforeCounters < 25 && afterCounters >= 25 {
		p.numAttacks += 99
		p.baseNumAttacks += 99
	}
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
	err := errors.New("not a death to self event")
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

// Helper for Dagaz.
// Activate the curse destruction card by selecting which
// curse to destroy. Return a function that will destroy the selected
// curse once called upon in the event stack.
// Assumes the number of curses is greater than 0.
// p *player: The player that has a curse.
// m *mArea: The monster area of the board
// l int: the number of curses
func (p *player) dagazCurseHelper(m *mArea, l int) lootCardEffect {
	var i uint8
	if l > 1 {
		showMonsterCards(p.Curses, 0)
		i = uint8(readInput(0, l-1))
	}
	curseId := p.Curses[i].id
	return func(roll uint8, blankCard bool) {
		i, err := p.getCurseIndex(curseId)
		if err == nil {
			c := p.popCurse(i)
			m.discard(&c)
		}
	}
}

// Helper to any action / effect that damages a monster.
// search the zone to make sure the monster is in play, then push damage to the monster.
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

func (b *Board) damageMonsterToPlayer(m *monsterCard, target *player, n uint8, combatRoll uint8) {
	if !target.isDead() {
		b.eventStack.push(event{p: target, e: damageEvent{monster: m, target: target, n: n}, roll: combatRoll})
		b.preventDamageHelper(target, b.eventStack.peek())
	}
}

// p *player: the player that pushed the event to the stack, if applicable (monster attack)
// target *player: the subject of the damage
// n uint8: how much damage to inflict
func (b *Board) damagePlayerToPlayer(p, target *player, n uint8) {
	if !target.isDead() { // Not dead
		b.eventStack.push(event{p: p, e: damageEvent{target: target, n: n}})
		b.preventDamageHelper(p, b.eventStack.peek())
	}
}

func (b *Board) preventDamageHelper(p *player, damageNode *eventNode) {
	damagePrevention := [2]uint16{guppysHairball, theDeadCat}
	for _, id := range damagePrevention {
		var f cardEffect
		if _, ok := damageNode.event.e.(damageEvent); ok {
			if i, err := p.getItemIndex(id, true); err == nil {
				var c card = p.PassiveItems[i]
				if id == guppysHairball {
					f = guppysHairballChecker(&b.eventStack, damageNode)
					b.eventStack.push(event{p: p, e: triggeredEffectEvent{c: c, f: f}})
					b.rollDiceAndPush()
				} else if id == theDeadCat {
					var err error
					if f, err = theDeadCatChecker(p.PassiveItems[i].(*treasureCard), &b.eventStack, damageNode); err != nil {
						continue
					}
					b.eventStack.push(event{p: p, e: triggeredEffectEvent{c: c, f: f}})
				}
			}
		}
	}
}

// Helper for all active cards where you look for a generic dice roll
func (es eventStack) diceItem() *eventNode {
	var ans int
	var rollNode *eventNode
	nodes := es.getDiceRollEvents()
	l := len(nodes)
	if l > 0 {
		if l > 1 {
			showEvents(nodes)
			ans = readInput(0, l-1)
		}
		rollNode = nodes[ans]
	}
	return rollNode
}

func (p *player) discardHandChoiceHelper(la *lArea, n uint8) {
	var i uint8
	for i = 0; i < n; i++ {
		showLootCards(p.Hand, p.Character.name, 0)
		fmt.Println("Choose what to discard")
		la.discard(p.popHandCard(uint8(readInput(0, len(p.Hand)-1))))
	}
}

func (m *mArea) fillMonsterZone(ap *player, b *Board, i uint8) {
	card := m.draw()
	for card.isBonusCard() {
		if f, special, err := card.f(ap, b, card); err == nil {
			b.eventStack.push(event{p: ap, e: triggeredEffectEvent{c: card, f: f}})
			if special {
				b.rollDiceAndPush()
			}
		}
		card = m.draw()
	}
	m.zones[i].push(card)
}

// Helper for the "Baby/Daddy/Mama Haunt" card.
// Before paying penalties, give this card to another player.
// Choose the player, then give them the card.
func (b Board) hauntGiveAwayHelper(p *player, hauntCard itemCard) {
	others := b.getOtherPlayers(p, false)
	l := len(others)
	var i uint8
	if l > 1 {
		showPlayers(others, 0)
		fmt.Println(fmt.Sprintf("Which player should get %s?", hauntCard.getName()))
		i = uint8(readInput(0, l-1))
	}
	others[i].stealItem(hauntCard.getId(), hauntCard.isPassive(), p)
}

func (b *Board) incubus(p, p2 *player) cardEffect {
	return func(roll uint8) {
		fmt.Println("0) Do Nothing.")
		showLootCards(p2.Hand, p2.Character.name, 1)
		j := uint8(readInput(0, len(p2.Hand)))
		if j > 0 {
			j -= 1
			showLootCards(p.Hand, p.Character.name, 0)
			fmt.Println("Choose a value to give to your opponent.")
			i := uint8(readInput(0, len(p.Hand)-1))
			p2Card := p2.Hand[j]
			p2.Hand[j] = p.Hand[i]
			p.Hand[i] = p2Card
		}
	}
}

func (p *player) incubus(l *lArea) cardEffect {
	return func(roll uint8) {
		p.loot(l)
		showLootCards(p.Hand, p.Character.name, 0)
		fmt.Println("Place value on top of the loot deck.")
		ans := readInput(0, len(p.Hand)-1)
		l.placeInDeck(p.popHandCard(uint8(ans)), true)
	}
}

// Helper for judgement
// Iterate through all players in the game, and collect
// the number of souls each player has. Represent them in a
// map where key = number of souls and value = player(s) with
// that number of souls.
// Additionally, return the key that contains the player(s) with
// the most souls.
func (b *Board) judgementHelper() (map[uint8][]*player, uint8) {
	players := b.getPlayers(false)
	numPlayers := len(players)
	soulsMap := make(map[uint8][]*player)
	var max uint8
	for i := range players {
		l := uint8(len(b.players[i].Souls))
		if _, ok := soulsMap[l]; !ok {
			soulsMap[l] = make([]*player, 0, numPlayers)
		}
		soulsMap[l] = append(soulsMap[l], players[i])
		if l > max {
			max = l
		}
	}
	return soulsMap, max
}

func (b *Board) killMonster(p *player, mId uint16) {
	if i, err := b.monster.getActiveMonster(mId); err == nil {
		m := b.monster.zones[i].pop()
		m.resetStats()
		if m.f != nil {
			if f, _, err := m.f(p, b, m); err == nil { // on death trigger
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

// Method that pushes the death event to the stack for a player
// Here's the player's passive opportunity to prevent death and end his / her turn
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

func deathPlayerPrevention(id uint16, p *player, b *Board, en *eventNode) {
	var f cardEffect
	var err error
	var i uint8
	if _, ok := en.event.e.(deathOfCharacterEvent); ok { // No need to perform if event's already fizzled
		if i, err = p.getItemIndex(id, true); err == nil && (id == brokenAnkh || id == guppysCollar) {
			f = func(roll uint8) {
				if (id == brokenAnkh && roll == 6) || (id == guppysCollar && roll >= 1 && roll <= 3) {
					en.event.e = fizzledEvent{}
					if err := p.isActivePlayer(b); err == nil {
						b.forceEndOfTurn()
					}
				}
			}
			b.eventStack.push(event{p: p, e: triggeredEffectEvent{c: p.PassiveItems[i], f: f}})
			b.rollDiceAndPush()
		}
	}
}

// Helper for cains eye, golden horse Shoe, and Purple Heart
func (b *Board) peekTrinketHelper(id uint16) cardEffect {
	cardDeckMap := map[uint16]*deck{
		cainsEye: &b.loot.deck, purpleHeart: &b.monster.deck, goldenHorseShoe: &b.treasure.deck}
	var f cardEffect
	if d, ok := cardDeckMap[id]; ok {
		f = func(roll uint8) {
			if c, err := d.peek(); err == nil {
				c.showCard(0)
				fmt.Println("1) Place this card on the bottom of the deck.\n2) Place it back on top.")
				if readInput(1, 2) == 1 {
					c, _ = d.pop()
					b.placeInDeck(c, false)
				}
			}
		}
	} else {
		panic(fmt.Errorf("card with id %d should not have called this function", id))
	}
	return f
}

// Helper for the "Remote Detonator" treasure card
// Tally the votes of whose item will be destroyed
//
// Returns:
// map[uint16]uint8: key = card voted on's id; value = numbner of votes
// map[uint16]bool: key = card voted on's id; value = is it passive
// map[uint16]*player: key = card voted on's id; value = pointer to player that owns the card
func (b *Board) remoteDetonatorVoteHelper() (map[uint16]uint8, map[uint16]bool, map[uint16]*player) {
	itemVotes := make(map[uint16]uint8, len(b.players)) // key = value id; value = number of votes
	cardType := make(map[uint16]bool, len(b.players))   // key = value id: value = isPassive
	items, owners := b.getAllItems(false, nil)
	showItems(items, 0)
	for range b.getPlayers(false) {
		fmt.Println("Vote for the item to destroy.")
		ans := readInput(0, len(items)-1)
		id, isPassive := items[ans].getId(), items[ans].isPassive()
		if _, ok := itemVotes[id]; !ok {
			itemVotes[id] = 0
		} else {
			itemVotes[id] += 1
		}
		cardType[id] = isPassive
	}
	return itemVotes, cardType, owners
}

// Helper to "The Bone" to get off it's first paid effect of adding one to any dice roll
func (es *eventStack) theBoneFirstPaidHelper(tc *treasureCard) (cardEffect, error) {
	var f cardEffect
	var err error
	rolls := es.getDiceRollEvents()
	l := len(rolls)
	if l == 0 {
		err = errors.New("no dice roll events")
	} else {
		tc.loseCounters(1)
		var i uint8
		if l > 1 {
			showEvents(rolls)
			i = uint8(readInput(0, l))
		}
		f = func(roll uint8) { es.addToDiceRoll(1, rolls[i]) }
	}
	return f, err
}

// Helper for "The Bone" to get off it's second paid effect of damaging another monster or player by 1 damage
func (b *Board) theBoneSecondPaidHelper(ap *player, players []*player, monsters []*monsterCard) cardEffect {
	l := len(players)
	showPlayers(players, 0)
	showMonsterCards(monsters, l)
	ans := readInput(0, l-1)
	var f cardEffect = func(roll uint8) {
		e, de := event{}, damageEvent{n: 1}
		if ans < l {
			b.damagePlayerToPlayer(ap, players[ans], 1)
			e.p, de.target = players[ans], players[ans]
		} else {
			b.damagePlayerToMonster(ap, monsters[ans], 1, 0)
			e.p, de.target = ap, monsters[ans-l]
		}
		e.e = de
		b.eventStack.push(e)
	}
	return f
}

// To be called for all loot cards that prevent damage on the stack.
// Dagaz, Soul Heart, and the Hierophant are such examples.
// The blank card will double the amount of damage prevented.
// Assumes the length will always be greater than 0.
func (es *eventStack) preventDamageWithLootHelper(damageEvents []*eventNode, n uint8) lootCardEffect {
	var i uint8
	l := len(damageEvents)
	if l > 1 {
		showEvents(damageEvents)
		i = uint8(readInput(0, l-1))
	}
	return func(roll uint8, blankCard bool) {
		if blankCard {
			n *= 2
		}
		_ = es.preventDamage(n, damageEvents[i])
	}
}

func modifyAttack(p *player, n uint8, leavingField bool) {
	if !leavingField {
		p.increaseBaseAttack(n)
		p.increaseAP(n)
	} else {
		p.decreaseBaseAttack(n)
		if p.Character.ap > p.Character.baseAttack {
			p.decreaseAP(n)
		}
	}
}

func modifyHealth(p *player, n uint8, leavingField bool) {
	if !leavingField {
		p.increaseBaseHealth(n)
		p.increaseHP(n)
	} else {
		p.decreaseBaseHealth(n)
		if p.Character.hp > p.Character.baseHealth {
			p.decreaseHP(n)
		}
	}
}

// Helper function to ensure that any subtractions with uint8
// variables do not cause an underflow
func subtractUint8(a, b uint8) uint8 {
	var y uint8
	x := int8(a) - int8(b)
	if x < 0 {
		y = 0
	} else {
		y = uint8(x)
	}
	return y
}

func theMidasTouchHelper(m *mArea) {
	for p, _ := range m.theMidasTouch {
		p.gainCents(3)
	}
}
