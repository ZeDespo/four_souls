/*
Every function in this file will return the "deck" of each of the value types:
Character, Starting Items, lArea, mArea, and tArea.
Each of these functions will have their deck build sectioned off as the following:
- The first section of cards are from the base game.
- The second section of cards are from the first expansion pack (Kickstarter)
- The third section of cards are from the second expansion pack (Retail).
*/

package four_souls

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// The base structure for all cards.
// Every card in the game will have the following attributes.
type baseCard struct {
	name   string // the card's name
	effect string // the card's text
	id     uint16 // unique identifier
}

// Cards that will be each player's avatar
type characterCard struct {
	baseCard
	baseHealth         uint8 // Max amount of health a character has
	baseAttack         uint8 // baseAttack power
	tapped             bool  // Can use play from hand quick cardEffect
	hp                 uint8 // Health remaining for turn. If == 0, you're dead
	ap                 uint8 // Attack points for the turn.
	diceModifier       uint8 // Dice modifier for non-attack rolls
	attackDiceModifier uint8 // Dice modifier for attack rolls
}

// A generic function that details both the activation requirements and resolved effects
// of some card in the game, with the exception to loot cards played from the hand
//
// EX: Mr Boom Activation Req: Choose a target.
// Resolve By: Pushing a damage event
//
// EX2: Battery Bum. Activation: Pay 4 cents. Resolve: Recharge an Item
//
// EX3: Cambion Conception Activation: Damage dealt to player. Resolve: Add a counter.
//
// p *player: The activating player
// b *Board: The board
// target card: The card that invoked the function
// Return: the loot card effect to call on the stack, whether the card needs a dice roll, an activation error.
type activator func(p *player, b *Board, c card) (cardEffect, bool, error)

// A generic card effect to be pushed onto the event stack
// and resolve some card.
type cardEffect func(roll uint8)

type continuousActivator func(p *player, b *Board, c card, isLeaving bool)

type eventActivator func(p *player, b *Board, c card, en *eventNode) (cardEffect, bool, error)

// Identical to generic activator; however, a loot card's effect can change at the
// resolving step due to Blank Card (Double the effect of the next loot card you play).
// Therefore, loot cards must resolve on the stack with a special function type
type lootActivator func(p *player, b *Board) (lootCardEffect, bool, error)

type lootCard struct {
	baseCard
	eternal bool // can be an eternal trinket thanks to cards that can make a card eternal
	trinket bool
	f       lootActivator // The function assigned to the loot card for active and paid effects
	ef      eventActivator
	cf      continuousActivator
}

// A loot card effect to be pushed onto the stack and resolve some
// loot card.
//
// blankCard bool: If the player activated Blank Card, then the effects
// of most loot cards are doubled. Exceptions including all trinkets,
// Mega Battery, The Devil tarot card, etc.
type lootCardEffect func(roll uint8, blankCard bool)

// If starting health is 0, it's not a monster card (random happening or curse)
type monsterCard struct {
	baseCard
	baseHealth uint8 // Base health of a monster / boss as shown on the card. Will be 0 if card is a curse or bonus
	baseRoll   uint8 // The base die roll needed to damage the monster. If 0, special non-attackable monster
	baseAttack uint8 // The base numAttacks power of the monster as shown on the card
	hp         uint8 // Health Points counter for the turn
	ap         uint8 // Attack points counter for the turn
	roll       uint8 // roll needed to hit for the turn.
	inBattle   bool  // Shows if the monster is in an active battle state. True once player successfully declares numAttacks
	isBoss     bool  // Switch for if special conditions are met
	isCurse    bool
	f          activator // This function will be our "on deathPenalty" effect for monsters like Big Spider and Death
	ef         eventActivator
	rf         rewardGiver
}

// Reward for eliminating a monster
type reward struct {
	cents        uint8 // number of cents gained
	loot         uint8 // number of loot cards to add to hand
	treasure     uint8 // number of treasure cards to draw
	souls        uint8 // the value of the souls (1 or 2)
	rollRequired bool  // for rewards that go "n x loot cards" where n is the result of a die roll.
}

// return true if dice roll needed, else false
type rewardGiver func(b *Board) (cardEffect, bool) // Active Player always gains rewards

type treasureCard struct {
	baseCard
	eternal  bool                // Starting itemCard quality. Immune to theft and destruction.
	passive  bool                // Is the itemCard passive
	paid     bool                // Do you need to pay a cost to activate the itemCard (not once per turn)
	active   bool                // Do you need to activate the card
	tapped   bool                // Is the card activated?
	counters int8                // How many counters there are on the card.
	f        activator           // The function for the active / paid effects
	ef       eventActivator      // The function for event based passive effects
	cf       continuousActivator // The function for continuous effects
}

// Represents every card in the game
type card interface {
	getId() uint16
	getName() string
	header() string
	showCard(int) string
}

// Only Characters and some Active Items can be activated.
// These cards can only be used once before they need to
// recharge at the start of the player's next turn.
type activeCards interface {
	card
	activate(p *player, b *Board) error
	recharge()
}

type itemCard interface {
	card
	getCounters() int8
	getContinuousPassive() continuousActivator
	getEventPassive() eventActivator
	isEternal() bool
	isPassive() bool
}

// For character and monster cards
// These cards are the avatars for the player and
// the game's engine. They can be buffed, weakened,
// and succumb to deathPenalty.
type combatTarget interface {
	card
	decreaseAP(n uint8)
	decreaseBaseAttack(n uint8)
	decreaseBaseHealth(n uint8)
	decreaseHP(n uint8)
	increaseAP(n uint8)
	increaseBaseAttack(n uint8)
	increaseBaseHealth(n uint8)
	increaseHP(n uint8)
	isDead() bool
	heal(n uint8)
}

// Only passive treasure cards and trinket loot cards can inherit this interface!
type passiveItem interface {
	itemCard
	trigger(p *player, b *Board, en *eventNode) []event
}

// Activate a character card and allow a player to play a loot card from the hand
func (cc *characterCard) activate(p *player, b *Board) error {
	var err = errors.New("character card already tapped")
	e := activateEvent{c: cc}
	if !cc.tapped {
		err = nil
		e.f = func(roll uint8) {
			l := len(p.Hand)
			if l > 0 {
				showLootCards(p.Hand, p.Character.name, 0)
				fmt.Print("Play which card?")
				ans := readInput(0, l-1)
				err = p.Hand[ans].activate(p, b)
			}
		}
		b.eventStack.push(event{p: p, e: e})
	}
	return err
}

// Activate a monster's "on death" effect, a bonus card's
// on draw effect, or the "give curse" helper for drawn curse cards.
func (mc monsterCard) activate(p *player, b *Board) error {
	var err = errors.New("not an on death monster effect, bonus card, or curse")
	if mc.f != nil {
		e := triggeredEffectEvent{c: mc}
		var f cardEffect
		var rollRequired bool
		if f, rollRequired, err = mc.f(p, b, mc); err == nil {
			e.f = f
			b.eventStack.push(event{p: p, e: e})
			if rollRequired {
				b.rollDiceAndPush()
			}
		}
	}
	return err
}

// Activate a tappable active / paid item's effect.
// Active item effects can only be used once per charge.
// Paid item effects can be used as the player can pay the cost (cents, damage, etc).
func (tc *treasureCard) activate(p *player, b *Board) error {
	var err = errors.New("not an active or paid item")
	if (tc.active || tc.paid) && !tc.tapped {
		e := activateEvent{c: tc}
		var f cardEffect
		var specialCondition bool
		if f, specialCondition, err = tc.f(p, b, tc); err == nil {
			tc.tapped = tc.active // If solely a paid item will default to false
			if specialCondition && tc.id == guppysPaw {
				defer b.eventStack.push(event{p: p, e: damageEvent{target: p, n: 1}})
			} else if specialCondition && (tc.id == theBone || tc.id == techX) { // specialCondition = paid event used
				tc.tapped = false
			} else if specialCondition {
				defer b.rollDiceAndPush()
			}
			e.f = f
			b.eventStack.push(event{p: p, e: e})
		}
	}
	return err
}

// Play a loot card from the hand and invoke it's activator.
// If played card is a trinket, then it will add the trinket to the game
// If not, activate its effect, then discard the card
func (lc lootCard) activate(p *player, b *Board) error {
	var i uint8
	var err error
	if i, err = p.getHandCardIndexById(lc.id); err == nil {
		e := lootCardEvent{l: lc}
		if lc.trinket {
			e.f = func(roll uint8, blankCard bool) {}
			p.addCardToBoard(lc)
		} else {
			var f lootCardEffect
			var specialCondition bool
			f, specialCondition, err = lc.f(p, b)
			if err == nil {
				defer b.discard(p.popHandCard(i))
				e.f = f
				b.eventStack.push(event{p: p, e: e})
				if specialCondition && lc.id != temperance {
					b.rollDiceAndPush()
				} else if !specialCondition && lc.id == temperance {
					b.damagePlayerToPlayer(p, p, 1)
				} else if specialCondition && lc.id == temperance {
					b.damagePlayerToPlayer(p, p, 2)
				}
			}
		}
	}
	return err
}

// Trigger the cards without activating their effects.
func (tc *treasureCard) deathPenalty() {
	tc.tapped = true
}

func (p *player) decreaseAP(n uint8) {
	p.Character.ap = subtractUint8(p.Character.ap, n)
}

func (mc *monsterCard) decreaseAP(n uint8) {
	mc.ap = subtractUint8(mc.ap, n)
}

func (p *player) decreaseBaseAttack(n uint8) {
	p.Character.baseAttack = subtractUint8(p.Character.baseAttack, n)
}

func (mc *monsterCard) decreaseBaseAttack(n uint8) {
	mc.baseAttack = subtractUint8(mc.baseAttack, n)
}

func (p *player) decreaseBaseHealth(n uint8) {
	p.Character.baseHealth = subtractUint8(p.Character.baseHealth, n)
}

func (mc *monsterCard) decreaseBaseHealth(n uint8) {
	mc.baseHealth = subtractUint8(mc.baseHealth, n)
}

func (p *player) decreaseHP(n uint8) {
	p.Character.hp = subtractUint8(p.Character.hp, n)
}

func (mc *monsterCard) decreaseHP(n uint8) {
	mc.hp = subtractUint8(mc.hp, n)
}

func (lc lootCard) getContinuousPassive() continuousActivator {
	return lc.cf
}

func (mc monsterCard) getContinuousPassive() continuousActivator {
	return nil
}

func (tc treasureCard) getContinuousPassive() continuousActivator {
	return tc.cf
}

func (lc lootCard) getCounters() int8 {
	return 0
}

func (tc treasureCard) getCounters() int8 {
	return tc.counters
}

func (lc lootCard) getEventPassive() eventActivator {
	return lc.ef
}

func (mc monsterCard) getEventPassive() eventActivator {
	return mc.ef
}

func (tc treasureCard) getEventPassive() eventActivator {
	return tc.ef
}

func (p player) getId() uint16 {
	return p.Character.id
}

func (cc characterCard) getId() uint16 {
	return cc.id
}

func (lc lootCard) getId() uint16 {
	return lc.id
}

func (mc monsterCard) getId() uint16 {
	return mc.id
}

func (tc treasureCard) getId() uint16 {
	return tc.id
}

func (p player) getName() string {
	return p.Character.name
}

func (cc characterCard) getName() string {
	return cc.name
}

func (lc lootCard) getName() string {
	return lc.name
}

func (mc monsterCard) getName() string {
	return mc.name
}

func (tc treasureCard) getName() string {
	return tc.name
}

func (p *player) increaseAP(n uint8) {
	p.Character.ap += n
}

func (mc *monsterCard) increaseAP(n uint8) {
	mc.ap += n
}

func (p *player) increaseBaseAttack(n uint8) {
	p.Character.baseAttack += n
}

func (mc *monsterCard) increaseBaseAttack(n uint8) {
	mc.baseAttack += n
}

func (p *player) increaseBaseHealth(n uint8) {
	p.Character.baseHealth += n
}

func (mc *monsterCard) increaseBaseHealth(n uint8) {
	mc.baseHealth += n
}

func (p *player) increaseHP(n uint8) {
	p.Character.hp += n
}

func (mc *monsterCard) increaseHP(n uint8) {
	mc.hp += n
}

func (mc monsterCard) isBonusCard() bool {
	var isBonusCard bool
	if mc.baseHealth == 0 {
		isBonusCard = true
	}
	return isBonusCard
}

func (p player) isDead() bool {
	var isDead bool
	if p.Character.hp <= 0 {
		isDead = true
	}
	return isDead
}

func (mc monsterCard) isDead() bool {
	var isDead bool
	if mc.hp <= 0 {
		isDead = true
	}
	return isDead
}

func (lc lootCard) isEternal() bool {
	return lc.eternal
}

func (tc treasureCard) isEternal() bool {
	return tc.eternal
}

func (lc lootCard) isPassive() bool {
	var isPassive bool
	if lc.trinket {
		isPassive = true
	}
	return isPassive
}

func (tc treasureCard) isPassive() bool {
	return tc.passive
}

func (mc *monsterCard) heal(n uint8) {
	if mc.hp < mc.baseHealth {
		toHeal := mc.hp + n
		if toHeal > mc.baseHealth {
			mc.hp = mc.baseHealth
		} else {
			mc.hp = toHeal
		}
	}
}

func (p *player) heal(n uint8) {
	if p.Character.hp < p.Character.baseHealth {
		toHeal := p.Character.hp + n
		if toHeal > p.Character.baseHealth {
			p.Character.hp = p.Character.baseHealth
		} else {
			p.Character.hp = toHeal
		}
	}
}

func (tc *treasureCard) loseCounters(n int8) {
	x := int8(tc.counters) - n
	if x < 0 {
		x = 0
	}
}

func (p *player) modifyDiceRoll(n int8) {}

func (mc *monsterCard) modifyDiceRoll(n int8) {
	modifyDiceRoll(&mc.roll, n)
}

func (cc *characterCard) recharge() {
	cc.tapped = false
}

func (tc *treasureCard) recharge() {
	if tc.active {
		tc.tapped = false
	}
}

func (mc *monsterCard) trigger(p *player, b *Board, en *eventNode) []event {
	events := make([]event, 0, 2)
	if mc.ef != nil {
		events = executeEventFunction(p, b, mc, en, mc.ef)
	}
	return events
}

func (lc lootCard) trigger(p *player, b *Board, en *eventNode) []event {
	events := make([]event, 0, 2)
	if lc.isPassive() && lc.ef != nil {
		events = executeEventFunction(p, b, lc, en, lc.ef)
	}
	return events
}

func (tc *treasureCard) trigger(p *player, b *Board, en *eventNode) []event {
	events := make([]event, 0, 2)
	if tc.passive && tc.ef != nil {
		events = executeEventFunction(p, b, tc, en, tc.ef)
	}
	return events
}

func executeEventFunction(p *player, b *Board, c card, en *eventNode, ef eventActivator) []event {
	events := make([]event, 1, 2)
	if f, rollRequired, err := ef(p, b, c, en); err == nil && f != nil {
		events[0] = event{p: p, e: triggeredEffectEvent{c: c, f: f}}
		if rollRequired {
			e, _ := b.rollDice()
			events = append(events, event{p: p, e: e})
		}
	}
	return events
}

// Initialize the characters and return only the number of characters that is required to play the game.
func getCharacters(numPlayers uint8, useExpansionOne bool, useExpansionTwo bool) []characterCard {
	charDeck := getCharacterCards(useExpansionOne, useExpansionTwo)
	var deck = make([]characterCard, 0, numPlayers)
	rand.Seed(time.Now().UnixNano())
	for uint8(len(deck)) < numPlayers {
		index := rand.Intn(len(charDeck))
		c := charDeck[index]
		deck = append(deck, c)
		charDeck = append(charDeck[:index], charDeck[index+1:]...)
	}
	return deck
}

// Initialize the starting items for all of the characters in the game
func getStartingItems() map[string]treasureCard {
	return getStartingCards()
}

// Helper function for creating a new game.
// Get all of the loot cards and shuffle them into a deck.
// param useExpansionOne bool: Include cards in the first expansion pass.
// param useExpansionTwo bool: Include cards in the second expansion pass.
// return deck: a linked list representing the deck.
func getLootDeck(useExpansionOne bool, useExpansionTwo bool) deck {
	var lootDeck deck = getLootCards(useExpansionOne, useExpansionTwo)
	lootDeck.shuffle()
	return lootDeck
}

// Get all monster cards, shuffle them into a deck.
// param useExpansionOne bool: Include cards in the first expansion pass.
// param useExpansionTwo bool: Include cards in the second expansion pass.
// return deck: a linked list representing the deck.
func getMonsterDeck(useExpansionOne bool, useExpansionTwo bool) deck {
	monsterDeck := getMonsterCards(useExpansionOne, useExpansionTwo)
	monsterDeck.shuffle()
	return monsterDeck
}

// Get all treasure cards, shuffle them into a deck.
// param useExpansionOne bool: Include cards in the first expansion pass.
// param useExpansionTwo bool: Include cards in the second expansion pass.
// return deck: a linked list representing the deck.
func getTreasureDeck(useExpansionOne bool, useExpansionTwo bool) deck {
	treasureDeck := getTreasureCards(useExpansionOne, useExpansionTwo)
	treasureDeck.shuffle()
	return treasureDeck
}

// Returns cards from the character deck as a slice.
func getCharacterCards(useExpansionOne bool, useExpansionTwo bool) []characterCard {
	var deck = []characterCard{
		{baseCard: baseCard{id: blueBaby, name: "Blue Baby", effect: characterEffect}, baseHealth: 2, baseAttack: 1, tapped: true},
		{baseCard: baseCard{id: cain, name: "Cain", effect: characterEffect}, baseHealth: 2, baseAttack: 1, tapped: true},
		{baseCard: baseCard{id: eden, name: "Eden", effect: characterEdenEffect}, baseHealth: 2, baseAttack: 1, tapped: true},
		{baseCard: baseCard{id: eve, name: "Eve", effect: characterEffect}, baseHealth: 2, baseAttack: 1, tapped: true},
		{baseCard: baseCard{id: isaac, name: "Isaac", effect: characterEffect}, baseHealth: 2, baseAttack: 1, tapped: true},
		{baseCard: baseCard{id: judas, name: "Judas", effect: characterEffect}, baseHealth: 2, baseAttack: 1, tapped: true},
		{baseCard: baseCard{id: lazarus, name: "Lazarus", effect: characterEffect}, baseHealth: 2, baseAttack: 1, tapped: true},
		{baseCard: baseCard{id: lilith, name: "Lilith", effect: characterEffect}, baseHealth: 2, baseAttack: 1, tapped: true},
		{baseCard: baseCard{id: maggy, name: "Maggy", effect: characterEffect}, baseHealth: 2, baseAttack: 1, tapped: true},
		{baseCard: baseCard{id: samson, name: "Samson", effect: characterEffect}, baseHealth: 2, baseAttack: 1, tapped: true},
		{baseCard: baseCard{id: theForgotten, name: "The Forgotten", effect: characterEffect}, baseHealth: 2, baseAttack: 1, tapped: true},
	}
	if useExpansionOne {
		deck = append(deck,
			characterCard{baseCard: baseCard{id: apollyon, name: "Apollyon", effect: characterEffect}, baseHealth: 2, baseAttack: 1, tapped: true},
			characterCard{baseCard: baseCard{id: azazel, name: "Azazel", effect: characterEffect}, baseHealth: 2, baseAttack: 1, tapped: true},
			characterCard{baseCard: baseCard{id: theKeeper, name: "The Keeper", effect: characterEffect}, baseHealth: 2, baseAttack: 1, tapped: true},
			characterCard{baseCard: baseCard{id: theLost, name: "The Lost", effect: characterEffect}, baseHealth: 1, baseAttack: 1, tapped: true},
		)
	}
	if useExpansionTwo {
		deck = append(deck,
			characterCard{baseCard: baseCard{id: bumboCharacter, name: "Bum-Bo", effect: characterEffect}, baseHealth: 2, baseAttack: 1, tapped: true},
			characterCard{baseCard: baseCard{id: darkJudas, name: "Dark Judas", effect: characterEffect}, baseHealth: 2, baseAttack: 1, tapped: true},
			characterCard{baseCard: baseCard{id: guppy, name: "Guppy", effect: characterEffect}, baseHealth: 2, baseAttack: 1, tapped: true},
			characterCard{baseCard: baseCard{id: whoreOfBabylon, name: "Whore of Babylon", effect: characterEffect}, baseHealth: 2, baseAttack: 1, tapped: true},
		)

	}
	return deck
}

// Returns cards for the loot deck as a map
// There are many cards that appear more than once. It makes sense to have these cards represented in a
// hash table rather than repeat code with a slice literal.
func getLootCards(useExpansionOne bool, useExpansionTwo bool) deck {
	var n uint8 = 105
	if useExpansionOne {
		n += 21
	}
	if useExpansionTwo {
		n += 31
	}
	d := make(deck, 0, n)
	d.append([]card{
		lootCard{baseCard: baseCard{name: "A Penny!", effect: "", id: aPenny}, f: aPennyFunc},
		lootCard{baseCard: baseCard{name: "A Penny!", effect: "", id: aPenny}, f: aPennyFunc},
		lootCard{baseCard: baseCard{name: "A Penny!", effect: "", id: aPenny}, f: aPennyFunc},
		lootCard{baseCard: baseCard{name: "A Penny!", effect: "", id: aPenny}, f: aPennyFunc},
		lootCard{baseCard: baseCard{name: "A Penny!", effect: "", id: aPenny}, f: aPennyFunc},
		lootCard{baseCard: baseCard{name: "A Penny!", effect: "", id: aPenny}, f: aPennyFunc},
		lootCard{baseCard: baseCard{name: "2 Cents!", effect: "", id: twoCents}, f: twoCentsFunc},
		lootCard{baseCard: baseCard{name: "2 Cents!", effect: "", id: twoCents}, f: twoCentsFunc},
		lootCard{baseCard: baseCard{name: "2 Cents!", effect: "", id: twoCents}, f: twoCentsFunc},
		lootCard{baseCard: baseCard{name: "2 Cents!", effect: "", id: twoCents}, f: twoCentsFunc},
		lootCard{baseCard: baseCard{name: "2 Cents!", effect: "", id: twoCents}, f: twoCentsFunc},
		lootCard{baseCard: baseCard{name: "2 Cents!", effect: "", id: twoCents}, f: twoCentsFunc},
		lootCard{baseCard: baseCard{name: "2 Cents!", effect: "", id: twoCents}, f: twoCentsFunc},
		lootCard{baseCard: baseCard{name: "2 Cents!", effect: "", id: twoCents}, f: twoCentsFunc},
		lootCard{baseCard: baseCard{name: "2 Cents!", effect: "", id: twoCents}, f: twoCentsFunc},
		lootCard{baseCard: baseCard{name: "2 Cents!", effect: "", id: twoCents}, f: twoCentsFunc},
		lootCard{baseCard: baseCard{name: "2 Cents!", effect: "", id: twoCents}, f: twoCentsFunc},
		lootCard{baseCard: baseCard{name: "2 Cents!", effect: "", id: twoCents}, f: twoCentsFunc},
		lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: threeCentsFunc},
		lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: threeCentsFunc},
		lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: threeCentsFunc},
		lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: threeCentsFunc},
		lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: threeCentsFunc},
		lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: threeCentsFunc},
		lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: threeCentsFunc},
		lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: threeCentsFunc},
		lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: threeCentsFunc},
		lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: threeCentsFunc},
		lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: threeCentsFunc},
		lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: threeCentsFunc},
		lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: threeCentsFunc},
		lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: threeCentsFunc},
		lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: threeCentsFunc},
		lootCard{baseCard: baseCard{name: "4 Cents!", effect: "", id: fourCents}, f: fourCentsFunc},
		lootCard{baseCard: baseCard{name: "4 Cents!", effect: "", id: fourCents}, f: fourCentsFunc},
		lootCard{baseCard: baseCard{name: "4 Cents!", effect: "", id: fourCents}, f: fourCentsFunc},
		lootCard{baseCard: baseCard{name: "4 Cents!", effect: "", id: fourCents}, f: fourCentsFunc},
		lootCard{baseCard: baseCard{name: "4 Cents!", effect: "", id: fourCents}, f: fourCentsFunc},
		lootCard{baseCard: baseCard{name: "4 Cents!", effect: "", id: fourCents}, f: fourCentsFunc},
		lootCard{baseCard: baseCard{name: "4 Cents!", effect: "", id: fourCents}, f: fourCentsFunc},
		lootCard{baseCard: baseCard{name: "4 Cents!", effect: "", id: fourCents}, f: fourCentsFunc},
		lootCard{baseCard: baseCard{name: "4 Cents!", effect: "", id: fourCents}, f: fourCentsFunc},
		lootCard{baseCard: baseCard{name: "4 Cents!", effect: "", id: fourCents}, f: fourCentsFunc},
		lootCard{baseCard: baseCard{name: "A Nickel!", effect: "", id: aNickel}, f: aNickelFunc},
		lootCard{baseCard: baseCard{name: "A Nickel!", effect: "", id: aNickel}, f: aNickelFunc},
		lootCard{baseCard: baseCard{name: "A Nickel!", effect: "", id: aNickel}, f: aNickelFunc},
		lootCard{baseCard: baseCard{name: "A Nickel!", effect: "", id: aNickel}, f: aNickelFunc},
		lootCard{baseCard: baseCard{name: "A Nickel!", effect: "", id: aNickel}, f: aNickelFunc},
		lootCard{baseCard: baseCard{name: "A Dime!!", effect: "", id: aDime}, f: aDimeFunc},
		lootCard{baseCard: baseCard{name: "Blank Rune", effect: "", id: blankRune}, f: blankRuneFunc},
		lootCard{baseCard: baseCard{name: "Bomb!", effect: "", id: bomb}, f: bombFunc},
		lootCard{baseCard: baseCard{name: "Bomb!", effect: "", id: bomb}, f: bombFunc},
		lootCard{baseCard: baseCard{name: "Bomb!", effect: "", id: bomb}, f: bombFunc},
		lootCard{baseCard: baseCard{name: "Bomb!", effect: "", id: bomb}, f: bombFunc},
		lootCard{baseCard: baseCard{name: "Butter Bean!", effect: "", id: butterBean}, f: butterBeanFunc},
		lootCard{baseCard: baseCard{name: "Butter Bean!", effect: "", id: butterBean}, f: butterBeanFunc},
		lootCard{baseCard: baseCard{name: "Butter Bean!", effect: "", id: butterBean}, f: butterBeanFunc},
		lootCard{baseCard: baseCard{name: "Dagaz", effect: "", id: dagaz}, f: dagazFunc},
		lootCard{baseCard: baseCard{name: "Dice Shard", effect: "", id: diceShard}, f: diceShardFunc},
		lootCard{baseCard: baseCard{name: "Dice Shard", effect: "", id: diceShard}, f: diceShardFunc},
		lootCard{baseCard: baseCard{name: "Dice Shard", effect: "", id: diceShard}, f: diceShardFunc},
		lootCard{baseCard: baseCard{name: "Ehwaz", effect: "", id: ehwaz}, f: ehwazFunc},
		lootCard{baseCard: baseCard{name: "Gold Bomb!!", effect: "", id: goldBomb}, f: goldBombFunc},
		lootCard{baseCard: baseCard{name: "Lil Battery", effect: "", id: lilBattery}, f: lilBatteryFunc},
		lootCard{baseCard: baseCard{name: "Lil Battery", effect: "", id: lilBattery}, f: lilBatteryFunc},
		lootCard{baseCard: baseCard{name: "Lil Battery", effect: "", id: lilBattery}, f: lilBatteryFunc},
		lootCard{baseCard: baseCard{name: "Lil Battery", effect: "", id: lilBattery}, f: lilBatteryFunc},
		lootCard{baseCard: baseCard{name: "Lost Soul", effect: "", id: lostSoul}, f: lostSoulFunc},
		lootCard{baseCard: baseCard{name: "Mega Battery", effect: "", id: megaBattery}, f: megaBatteryFunc},
		lootCard{baseCard: baseCard{name: "Pills! (Blue)", effect: "", id: pillsBlue}, f: pillsBlueFunc},
		lootCard{baseCard: baseCard{name: "Pills! (Red)", effect: "", id: pillsRed}, f: pillsRedFunc},
		lootCard{baseCard: baseCard{name: "Pills! (Yellow)", effect: "", id: pillsYellow}, f: pillsYellowFunc},
		lootCard{baseCard: baseCard{name: "Soul Heart", effect: "", id: soulHeart}, f: soulHeartFunc},
		lootCard{baseCard: baseCard{name: "Soul Heart", effect: "", id: soulHeart}, f: soulHeartFunc},
		lootCard{baseCard: baseCard{name: "0. The Fool", effect: "", id: theFool}, f: theFoolFunc},
		lootCard{baseCard: baseCard{name: "I. The Magician", effect: "", id: theMagician}, f: theMagicianFunc},
		lootCard{baseCard: baseCard{name: "II. The High Priestess", effect: "", id: theHighPriestess}, f: theHighPriestessFunc},
		lootCard{baseCard: baseCard{name: "III. The Empress", effect: "", id: theEmpress}, f: theEmpressFunc},
		lootCard{baseCard: baseCard{name: "IV. The Emperor", effect: "", id: theEmperor}, f: theEmperorFunc},
		lootCard{baseCard: baseCard{name: "V. The Hierophant", effect: "", id: theHierophant}, f: theHierophantFunc},
		lootCard{baseCard: baseCard{name: "VI. The Lovers", effect: "", id: theLovers}, f: theLoversFunc},
		lootCard{baseCard: baseCard{name: "VII. The Chariot", effect: "", id: theChariot}, f: theChariotFunc},
		lootCard{baseCard: baseCard{name: "VIII. Justice", effect: "", id: justice}, f: justiceFunc},
		lootCard{baseCard: baseCard{name: "IX. The Hermit", effect: "", id: theHermit}, f: theHermitFunc},
		lootCard{baseCard: baseCard{name: "X. Wheel of Fortune", effect: "", id: wheelOfFortune}, f: wheelOfFortuneFunc},
		lootCard{baseCard: baseCard{name: "XI. Strength", effect: "", id: strength}, f: strengthFunc},
		lootCard{baseCard: baseCard{name: "XII. The Hanged Man", effect: "", id: theHangedMan}, f: theHangedManFunc},
		lootCard{baseCard: baseCard{name: "XIII. Death", effect: "", id: deathLoot}, f: deathTarotCardFunc},
		lootCard{baseCard: baseCard{name: "XIV. The Tower", effect: "", id: theTower}, f: theTowerFunc},
		lootCard{baseCard: baseCard{name: "XV. The Devil", effect: "", id: theDevil}, f: theDevilFunc},
		lootCard{baseCard: baseCard{name: "XVI. Temperance", effect: "", id: temperance}, f: temperanceFunc},
		lootCard{baseCard: baseCard{name: "XVII. The Stars", effect: "", id: theStars}, f: theStarsFunc},
		lootCard{baseCard: baseCard{name: "XVIII. The Moon", effect: "", id: theMoon}, f: theMoonFunc},
		lootCard{baseCard: baseCard{name: "XIX. The Sun", effect: "", id: theSun}, f: theSunFunc},
		lootCard{baseCard: baseCard{name: "XX. Judgement", effect: "", id: judgement}, f: judgementFunc},
		lootCard{baseCard: baseCard{name: "XXI. The World", effect: "", id: theWorld}, f: theWorldFunc},
		lootCard{baseCard: baseCard{name: "Bloody Penny", effect: "", id: bloodyPenny}, trinket: true},
		lootCard{baseCard: baseCard{name: "Broken Ankh", effect: "", id: brokenAnkh}, trinket: true},
		lootCard{baseCard: baseCard{name: "Cain's Eye", effect: "", id: cainsEye}, trinket: true},
		lootCard{baseCard: baseCard{name: "Counterfeit Penny", effect: "", id: counterfeitPenny}, trinket: true},
		lootCard{baseCard: baseCard{name: "Curved Horn", effect: "", id: curvedHorn}, trinket: true},
		lootCard{baseCard: baseCard{name: "Golden Horseshoe", effect: "", id: goldenHorseShoe}, trinket: true},
		lootCard{baseCard: baseCard{name: "Guppy's Hairball", effect: "", id: guppysHairball}, trinket: true},
		lootCard{baseCard: baseCard{name: "Purple Heart", effect: "", id: purpleHeart}, trinket: true},
		lootCard{baseCard: baseCard{name: "Swallowed Penny", effect: "", id: swallowedPenny}, trinket: true},
	}...)
	if useExpansionOne {
		d.append([]card{
			lootCard{baseCard: baseCard{name: "A Penny!", effect: "", id: aPenny}, f: nil},
			lootCard{baseCard: baseCard{name: "A Penny!", effect: "", id: aPenny}, f: nil},
			lootCard{baseCard: baseCard{name: "2 Cents!", effect: "", id: twoCents}, f: nil},
			lootCard{baseCard: baseCard{name: "2 Cents!", effect: "", id: twoCents}, f: nil},
			lootCard{baseCard: baseCard{name: "2 Cents!", effect: "", id: twoCents}, f: nil},
			lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: nil},
			lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: nil},
			lootCard{baseCard: baseCard{name: "4 Cents!", effect: "", id: fourCents}, f: nil},
			lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: nil},
			lootCard{baseCard: baseCard{name: "A Sack", effect: "", id: aSack}, f: aSackFunc},
			lootCard{baseCard: baseCard{name: "Bomb", effect: "", id: bomb}, f: nil},
			lootCard{baseCard: baseCard{name: "Charged Penny", effect: "", id: chargedPenny}, f: nil},
			lootCard{baseCard: baseCard{name: "Credit Card", effect: "", id: creditCard}, f: nil},
			lootCard{baseCard: baseCard{name: "Holy Card", effect: "", id: holyCard}, f: nil},
			lootCard{baseCard: baseCard{name: "Jera", effect: "", id: jera}, f: nil},
			lootCard{baseCard: baseCard{name: "Joker", effect: "", id: joker}, f: nil},
			lootCard{baseCard: baseCard{name: "Pills! (Purple)", effect: "", id: pillsPurple}, f: nil},
			lootCard{baseCard: baseCard{name: "Soul Heart", effect: "", id: soulHeart}, f: nil},
			lootCard{baseCard: baseCard{name: "Two of Diamonds", effect: "", id: twoOfDiamonds}, f: nil},
			lootCard{baseCard: baseCard{name: "Cancer", effect: "", id: cancer}, trinket: true, f: nil},
			lootCard{baseCard: baseCard{name: "Pink Eye", effect: "", id: pinkEye}, trinket: true, f: nil},
		}...)
	}
	if useExpansionTwo {
		d.append([]card{
			lootCard{baseCard: baseCard{name: "A Penny!", effect: "", id: aPenny}, f: nil},
			lootCard{baseCard: baseCard{name: "A Penny!", effect: "", id: aPenny}, f: nil},
			lootCard{baseCard: baseCard{name: "A Penny!", effect: "", id: aPenny}, f: nil},
			lootCard{baseCard: baseCard{name: "2 Cents!", effect: "", id: twoCents}, f: nil},
			lootCard{baseCard: baseCard{name: "2 Cents!", effect: "", id: twoCents}, f: nil},
			lootCard{baseCard: baseCard{name: "2 Cents!", effect: "", id: twoCents}, f: nil},
			lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: nil},
			lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: nil},
			lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: nil},
			lootCard{baseCard: baseCard{name: "3 Cents!", effect: "", id: threeCents}, f: nil},
			lootCard{baseCard: baseCard{name: "4 Cents!", effect: "", id: fourCents}, f: nil},
			lootCard{baseCard: baseCard{name: "4 Cents!", effect: "", id: fourCents}, f: nil},
			lootCard{baseCard: baseCard{name: "A Nickel!", effect: "", id: aNickel}, f: nil},
			lootCard{baseCard: baseCard{name: "A Nickel!", effect: "", id: aNickel}, f: nil},
			lootCard{baseCard: baseCard{name: "Ansuz", effect: "", id: ansuz}, f: nil},
			lootCard{baseCard: baseCard{name: "Black Rune", effect: "", id: blackRune}, f: nil},
			lootCard{baseCard: baseCard{name: "Bomb!", effect: "", id: bomb}, f: nil},
			lootCard{baseCard: baseCard{name: "Butter Bean!", effect: "", id: butterBean}, f: nil},
			lootCard{baseCard: baseCard{name: "Dice Shard", effect: "", id: diceShard}, f: nil},
			lootCard{baseCard: baseCard{name: "Get Out of Jail Card", effect: "", id: getOutOfJail}, f: nil},
			lootCard{baseCard: baseCard{name: "Gold Key", effect: "", id: goldKey}, f: nil},
			lootCard{baseCard: baseCard{name: "Lil Battery", effect: "", id: lilBattery}, f: nil},
			lootCard{baseCard: baseCard{name: "Perthro", effect: "", id: perthro}, f: nil},
			lootCard{baseCard: baseCard{name: "Pills! (Black)", effect: "", id: pillsBlack}, f: nil},
			lootCard{baseCard: baseCard{name: "Pills! (Spots)", effect: "", id: pillsSpots}, f: nil},
			lootCard{baseCard: baseCard{name: "Pills! (White)", effect: "", id: pillsWhite}, f: nil},
			lootCard{baseCard: baseCard{name: "? Card", effect: "", id: questionMarkCard}, f: nil},
			lootCard{baseCard: baseCard{name: "AAA Battery", effect: "", id: aaaBattery}, trinket: true, f: nil},
			lootCard{baseCard: baseCard{name: "Poker Chip", effect: "", id: pokerChip}, trinket: true, f: nil},
			lootCard{baseCard: baseCard{name: "Tape Worm", effect: "", id: tapeWorm}, trinket: true, f: nil},
			lootCard{baseCard: baseCard{name: "The Left Hand", effect: "", id: theLeftHand}, trinket: true, f: nil},
		}...)
	}
	return d
}

// Returns all the monster cards as a slice.
func getMonsterCards(useExpansionOne bool, useExpansionTwo bool) deck {
	var n uint8 = 106
	if useExpansionOne {
		n += 20
	}
	if useExpansionTwo {
		n += 30
	}
	deck := make(deck, 0, n)
	deck.append([]card{
		monsterCard{baseCard: baseCard{name: "Big Spider", id: bigSpider}, baseHealth: 3, baseRoll: 4, baseAttack: 1, f: bigSpiderDeath, rf: bigSpiderReward},
		monsterCard{baseCard: baseCard{name: "Black Bony", id: blackBony}, baseHealth: 3, baseRoll: 4, baseAttack: 1, f: blackBonyDeath, rf: blackBonyReward},
		monsterCard{baseCard: baseCard{name: "Boom Fly", id: boomFly}, baseHealth: 1, baseRoll: 4, baseAttack: 1, f: boomFlyDeath, rf: boomFlyReward},
		monsterCard{baseCard: baseCard{name: "Clotty", id: clotty}, baseHealth: 2, baseRoll: 3, baseAttack: 1, rf: clottyReward},
		monsterCard{baseCard: baseCard{name: "Cod Worm", id: codWorm}, baseHealth: 1, baseRoll: 5, rf: codWormReward},
		monsterCard{baseCard: baseCard{name: "Conjoined Fatty", id: conjoinedFatty}, baseHealth: 4, baseRoll: 3, baseAttack: 2, rf: conjoinedFattyReward},
		monsterCard{baseCard: baseCard{name: "Dank Globin", id: dankGlobin}, baseHealth: 2, baseRoll: 4, baseAttack: 2, f: dankGlobinDeath, rf: dankGlobinReward},
		monsterCard{baseCard: baseCard{name: "Dinga", id: dinga}, baseHealth: 3, baseRoll: 3, baseAttack: 1, rf: dingaReward},
		monsterCard{baseCard: baseCard{name: "Dip", id: dip}, baseHealth: 1, baseRoll: 4, baseAttack: 1, rf: dipReward},
		monsterCard{baseCard: baseCard{name: "Dople", id: dople}, baseHealth: 2, baseRoll: 4, baseAttack: 2, ef: dopleEvent, rf: dopleReward},
		monsterCard{baseCard: baseCard{name: "Evil Twin", id: evilTwin}, baseHealth: 3, baseRoll: 5, baseAttack: 2, ef: evilTwinEvent, rf: evilTwinReward},
		monsterCard{baseCard: baseCard{name: "Fat Bat", id: fatBat}, baseHealth: 3, baseRoll: 5, baseAttack: 1, rf: fatBatReward},
		monsterCard{baseCard: baseCard{name: "Fatty", id: fatty}, baseHealth: 4, baseRoll: 2, baseAttack: 1, rf: fattyReward},
		monsterCard{baseCard: baseCard{name: "Fly", id: fly}, baseHealth: 1, baseRoll: 2, baseAttack: 1, rf: flyReward},
		monsterCard{baseCard: baseCard{name: "Greedling", id: greedling}, baseHealth: 2, baseRoll: 5, baseAttack: 1, f: greedlingDeath, rf: greedlingReward},
		monsterCard{baseCard: baseCard{name: "Hanger", id: hanger}, baseHealth: 2, baseRoll: 4, baseAttack: 2, f: hangerDeath, rf: hangerReward},
		monsterCard{baseCard: baseCard{name: "Hopper", id: hopper}, baseHealth: 2, baseRoll: 3, baseAttack: 1, ef: hopperEvent, rf: hopperReward},
		monsterCard{baseCard: baseCard{name: "Horf", id: horf}, baseHealth: 1, baseRoll: 4, baseAttack: 1, rf: horfReward},
		monsterCard{baseCard: baseCard{name: "Keeper Head", id: keeperHead}, baseHealth: 2, baseRoll: 4, baseAttack: 1, ef: keeperHeadEvent, rf: keeperReward},
		monsterCard{baseCard: baseCard{name: "Leaper", id: leaper}, baseHealth: 2, baseRoll: 4, baseAttack: 1, rf: leaperReward},
		monsterCard{baseCard: baseCard{name: "Leech", id: leech}, baseHealth: 1, baseRoll: 4, baseAttack: 2, rf: leechReward},
		monsterCard{baseCard: baseCard{name: "Mom's Dead Hand", id: momsDeadHand}, baseHealth: 2, baseRoll: 5, baseAttack: 1, f: momsDeadHandDeath, rf: momsDeadHandReward},
		monsterCard{baseCard: baseCard{name: "Mom's Eye", id: momsEye}, baseHealth: 1, baseRoll: 4, baseAttack: 2, f: momsEyeDeath, rf: momsEyeReward},
		monsterCard{baseCard: baseCard{name: "Mom's Hand", id: momsHand}, baseHealth: 2, baseRoll: 4, baseAttack: 1, ef: momsHandEvent, rf: momsHandReward},
		monsterCard{baseCard: baseCard{name: "Mulliboom", id: mulliboom}, baseHealth: 1, baseRoll: 2, baseAttack: 4, f: mulliboomDeath, rf: mulliboomReward},
		monsterCard{baseCard: baseCard{name: "Mulligan", id: mulligan}, baseHealth: 1, baseRoll: 3, baseAttack: 1, f: mulliganDeath, rf: mulliganReward},
		monsterCard{baseCard: baseCard{name: "Pale Fatty", id: paleFatty}, baseHealth: 4, baseRoll: 3, baseAttack: 1, rf: paleFattyReward},
		monsterCard{baseCard: baseCard{name: "Pooter", id: pooter}, baseHealth: 2, baseRoll: 3, baseAttack: 1, rf: pooterReward},
		monsterCard{baseCard: baseCard{name: "Portal", id: portal}, baseHealth: 2, baseRoll: 4, baseAttack: 1, f: portalDeath, rf: portalReward},
		monsterCard{baseCard: baseCard{name: "Psy Horf", id: psyHorf}, baseHealth: 1, baseRoll: 5, baseAttack: 1, f: psyHorfDeath, rf: psyHorfReward},
		monsterCard{baseCard: baseCard{name: "Rage Creep", id: rageCreep}, baseHealth: 1, baseRoll: 5, baseAttack: 1, ef: rageCreepEvent, rf: rageCreepReward},
		monsterCard{baseCard: baseCard{name: "Red Host", id: redHost}, baseHealth: 2, baseRoll: 3, baseAttack: 2, rf: redHostReward},
		monsterCard{baseCard: baseCard{name: "Ring of Flies", id: ringOfFlies}, baseHealth: 3, baseRoll: 3, baseAttack: 1, ef: ringOfFliesEvent, rf: ringOfFliesReward},
		monsterCard{baseCard: baseCard{name: "Spider", id: spider}, baseHealth: 1, baseRoll: 4, baseAttack: 1, rf: spiderReward},
		monsterCard{baseCard: baseCard{name: "Squirt", id: squirt}, baseHealth: 2, baseRoll: 3, baseAttack: 1, rf: squirtReward},
		monsterCard{baseCard: baseCard{name: "Stoney", id: stoney}, baseHealth: 3, ef: stoneyEvent, rf: stoneyReward},
		monsterCard{baseCard: baseCard{name: "Swarm of Flies", id: swarmOfFlies}, baseHealth: 5, baseRoll: 2, baseAttack: 1, ef: swarmOfFliesEvent, rf: swarmOfFliesReward},
		monsterCard{baseCard: baseCard{name: "Trite", id: trite}, baseHealth: 1, baseRoll: 5, baseAttack: 1, rf: triteReward},
		monsterCard{baseCard: baseCard{name: "Wizoob", id: wizoob}, baseHealth: 3, baseRoll: 5, baseAttack: 1, f: wizoobDeath, rf: wizoobReward},
		monsterCard{baseCard: baseCard{name: "Cursed Fatty", id: cursedFatty}, baseHealth: 4, baseRoll: 2, baseAttack: 1, ef: cursedFattyEvent, rf: cursedFattyReward},
		monsterCard{baseCard: baseCard{name: "Cursed Gaper", id: cursedGaper}, baseHealth: 2, baseRoll: 4, baseAttack: 1, ef: cursedGaperEvent, rf: cursedGaperReward},
		monsterCard{baseCard: baseCard{name: "Cursed Horf", id: cursedHorf}, baseHealth: 1, baseRoll: 4, baseAttack: 1, ef: cursedHorfEvent, rf: cursedHorfReward},
		monsterCard{baseCard: baseCard{name: "Cursed Keeper Head", id: cursedKeeperHead}, baseHealth: 2, baseRoll: 4, baseAttack: 1, ef: cursedKeeperHeadEvent, rf: cursedKeeperHeadReward},
		monsterCard{baseCard: baseCard{name: "Cursed Mom's Hand", id: cursedMomsHand}, baseHealth: 2, baseRoll: 4, baseAttack: 1, ef: cursedMomsHandEvent, rf: cursedMomsHandReward},
		monsterCard{baseCard: baseCard{name: "Cursed Psy Horf", id: cursedPsyHorf}, baseHealth: 1, baseRoll: 5, baseAttack: 1, ef: cursedPsyHorfEvent, rf: cursedPsyHorfReward},
		monsterCard{baseCard: baseCard{name: "Holy Dinga", id: holyDinga}, baseHealth: 3, baseRoll: 3, baseAttack: 1, ef: holyDingaEvent, rf: holyDingaReward},
		monsterCard{baseCard: baseCard{name: "Holy Dip", id: holyDip}, baseHealth: 1, baseRoll: 4, baseAttack: 1, ef: holyDipEvent, rf: holyDipReward},
		monsterCard{baseCard: baseCard{name: "Holy Keeper Head", id: holyKeeperHead}, baseHealth: 2, baseRoll: 4, baseAttack: 1, ef: holyKeeperHeadEvent, rf: holyKeeperHeadReward},
		monsterCard{baseCard: baseCard{name: "Holy Mom's Eye", id: holyMomsEye}, baseHealth: 1, baseRoll: 4, baseAttack: 2, ef: holyMomsEyeEvent, rf: holyMomsEyeReward},
		monsterCard{baseCard: baseCard{name: "Holy Squirt", id: holySquirt}, baseHealth: 2, baseRoll: 3, baseAttack: 1, ef: holySquirtEvent, rf: holySquirtReward},
		monsterCard{baseCard: baseCard{name: "Carrion Queen", id: carrionQueen}, baseHealth: 3, baseRoll: 4, baseAttack: 1, isBoss: true, rf: carrionQueenReward},
		monsterCard{baseCard: baseCard{name: "Chub", id: chub}, baseHealth: 4, baseRoll: 3, baseAttack: 1, isBoss: true, ef: chubEvent, rf: chubReward},
		monsterCard{baseCard: baseCard{name: "Conquest", id: conquest}, baseHealth: 2, baseRoll: 3, baseAttack: 1, isBoss: true, f: conquestDeath, rf: conquestReward},
		monsterCard{baseCard: baseCard{name: "Daddy Long Legs", id: daddyLongLegsMonster}, baseHealth: 4, baseRoll: 4, baseAttack: 1, isBoss: true, ef: daddyLongLegsEvent, rf: daddyLongLegsReward},
		monsterCard{baseCard: baseCard{name: "Dark One", id: darkOne}, baseHealth: 3, baseRoll: 4, baseAttack: 1, isBoss: true, ef: darkOneEvent, rf: darkOneReward},
		monsterCard{baseCard: baseCard{name: "Death", id: deathMonster}, baseHealth: 3, baseRoll: 4, baseAttack: 2, isBoss: true, f: deathMonsterDeath, rf: deathMonsterReward},
		monsterCard{baseCard: baseCard{name: "Delirium", id: delirium}, baseHealth: 5, baseRoll: 4, baseAttack: 3, ef: deliriumEvent, rf: deliriumReward},
		monsterCard{baseCard: baseCard{name: "Envy", id: envy}, baseHealth: 2, baseRoll: 5, baseAttack: 1, isBoss: true, f: envyDeath, rf: envyReward},
		monsterCard{baseCard: baseCard{name: "Famine", id: famine}, baseHealth: 2, baseRoll: 3, baseAttack: 1, isBoss: true, f: famineDeath, rf: famineReward},
		monsterCard{baseCard: baseCard{name: "Gemini", id: gemini}, baseHealth: 3, baseRoll: 4, baseAttack: 1, isBoss: true, ef: geminiEvent, rf: geminiReward},
		monsterCard{baseCard: baseCard{name: "Gluttony", id: gluttony}, baseHealth: 4, baseRoll: 3, baseAttack: 1, isBoss: true, ef: gluttonyEvent, rf: gluttonyReward},
		monsterCard{baseCard: baseCard{name: "Greed", id: greedMonster}, baseHealth: 3, baseRoll: 4, baseAttack: 1, isBoss: true, ef: greedMonsterEvent, rf: greedMonsterReward},
		monsterCard{baseCard: baseCard{name: "Gurdy JR.", id: gurdyJr}, baseHealth: 2, baseRoll: 5, baseAttack: 1, isBoss: true, ef: gurdyJrEvent, rf: gurdyJrReward},
		monsterCard{baseCard: baseCard{name: "Gurdy", id: gurdy}, baseHealth: 5, baseRoll: 4, baseAttack: 1, isBoss: true, rf: gurdyReward},
		monsterCard{baseCard: baseCard{name: "Larry JR.", id: larryJr}, baseHealth: 4, baseRoll: 3, baseAttack: 1, isBoss: true, ef: larryJrEvent, rf: larryJrReward},
		monsterCard{baseCard: baseCard{name: "Little Horn", id: littleHorn}, baseHealth: 2, baseRoll: 6, baseAttack: 1, isBoss: true, rf: littleHornReward},
		monsterCard{baseCard: baseCard{name: "Lust", id: lust}, baseHealth: 2, baseRoll: 4, baseAttack: 1, isBoss: true, ef: lustEvent, rf: lustReward},
		monsterCard{baseCard: baseCard{name: "Mask of Infamy", id: maskOfInfamy}, baseHealth: 4, baseRoll: 4, baseAttack: 1, isBoss: true, ef: maskOfInfamyEvent, rf: maskOfInfamyReward},
		monsterCard{baseCard: baseCard{name: "Mega Fatty", id: megaFatty}, baseHealth: 3, baseRoll: 3, baseAttack: 1, isBoss: true, ef: megaFattyEvent, rf: megaFattyReward},
		monsterCard{baseCard: baseCard{name: "Monstro", id: monstro}, baseHealth: 4, baseRoll: 4, baseAttack: 1, isBoss: true, rf: monstroReward},
		monsterCard{baseCard: baseCard{name: "Peep", id: peep}, baseHealth: 3, baseRoll: 4, baseAttack: 1, isBoss: true, f: thePeepDeath, rf: thePeepReward},
		monsterCard{baseCard: baseCard{name: "Pestilence", id: pestilence}, baseHealth: 4, baseRoll: 4, baseAttack: 1, isBoss: true, f: pestilenceDeath, rf: pestilenceReward},
		monsterCard{baseCard: baseCard{name: "Pin", id: pin}, baseHealth: 2, baseRoll: 2, baseAttack: 1, isBoss: true, rf: pinReward},
		monsterCard{baseCard: baseCard{name: "Pride", id: pride}, baseHealth: 2, baseRoll: 4, baseAttack: 1, isBoss: true, ef: prideEvent, rf: prideReward},
		monsterCard{baseCard: baseCard{name: "Ragman", id: ragman}, baseHealth: 2, baseRoll: 3, baseAttack: 2, isBoss: true, f: ragmanDeath, rf: ragmanReward},
		monsterCard{baseCard: baseCard{name: "Scolex", id: scolex}, baseHealth: 3, baseRoll: 5, baseAttack: 1, isBoss: true, ef: scolexEvent, rf: scolexReward},
		monsterCard{baseCard: baseCard{name: "Sloth", id: sloth}, baseHealth: 3, baseRoll: 4, baseAttack: 1, isBoss: true, f: slothDeath, rf: slothReward},
		monsterCard{baseCard: baseCard{name: "The Bloat", id: theBloat}, baseHealth: 4, baseRoll: 4, baseAttack: 2, isBoss: true, ef: theBloatEvent, rf: theBloatReward},
		monsterCard{baseCard: baseCard{name: "The Duke Of Flies", id: theDukeOfFlies}, baseHealth: 4, baseRoll: 3, baseAttack: 1, isBoss: true, rf: theDukeOfFliesReward},
		monsterCard{baseCard: baseCard{name: "The Haunt", id: theHaunt}, baseHealth: 3, baseRoll: 4, baseAttack: 1, isBoss: true, ef: theHauntEvent, rf: theHauntReward},
		monsterCard{baseCard: baseCard{name: "War", id: war}, baseHealth: 3, baseRoll: 3, baseAttack: 1, isBoss: true, ef: warEvent, rf: warReward},
		monsterCard{baseCard: baseCard{name: "Wrath", id: wrath}, baseHealth: 3, baseRoll: 3, baseAttack: 1, isBoss: true, f: wrathDeath, rf: wrathReward},
		monsterCard{baseCard: baseCard{name: "Mom!", id: mom}, baseHealth: 5, baseRoll: 4, baseAttack: 2, isBoss: true, f: momDeath, rf: momReward},
		monsterCard{baseCard: baseCard{name: "Satan!", id: satan}, baseHealth: 6, baseRoll: 4, baseAttack: 2, isBoss: true, ef: satanEvent, rf: satanReward},
		monsterCard{baseCard: baseCard{name: "The Lamb", id: theLamb}, baseHealth: 6, baseRoll: 3, baseAttack: 6, f: theLambDeath, rf: theLambReward},
		monsterCard{baseCard: baseCard{name: "Ambush!", id: ambush}, f: ambushFunc},
		monsterCard{baseCard: baseCard{name: "Chest", id: chest}, f: chestFunc},
		monsterCard{baseCard: baseCard{name: "Chest", id: chest}, f: chestFunc},
		monsterCard{baseCard: baseCard{name: "Cursed Chest", id: cursedChest}, f: cursedChestFunc},
		monsterCard{baseCard: baseCard{name: "Dark Chest", id: darkChest}, f: darkChestFunc},
		monsterCard{baseCard: baseCard{name: "Dark Chest", id: darkChest}, f: darkChestFunc},
		monsterCard{baseCard: baseCard{name: "Devil Deal", id: devilDeal}, f: devilDealFunc},
		monsterCard{baseCard: baseCard{name: "Gold Chest", id: goldChest}, f: goldChestFunc},
		monsterCard{baseCard: baseCard{name: "Greed!", id: greedHappening}, f: greedBonusFunc},
		monsterCard{baseCard: baseCard{name: "I Can See Forever!", id: iCanSeeForever}, f: iCanSeeForeverFunc},
		monsterCard{baseCard: baseCard{name: "Troll Bombs", id: trollBombs}, f: trollBombsFunc},
		monsterCard{baseCard: baseCard{name: "Mega Troll Bomb!", id: megaTrollBomb}, f: megaTrollBombFunc},
		monsterCard{baseCard: baseCard{name: "Secret Room!", id: secretRoom}, f: secretRoomFunc},
		monsterCard{baseCard: baseCard{name: "Shop Upgrade!", id: shopUpgrade}, f: shopUpgradeFunc},
		monsterCard{baseCard: baseCard{name: "We Need To Go Deeper!", id: weNeedToGoDeeper}, f: weNeedToGoDeeperFunc},
		monsterCard{baseCard: baseCard{name: "XL Floor!", id: xlFloor}, f: xlFloorFunc},
		monsterCard{baseCard: baseCard{name: "Curse of Amnesia", id: curseOfAmnesia}, f: giveCurseHelper, ef: curseOfAmnesiaEvent},
		monsterCard{baseCard: baseCard{name: "Curse of Greed", id: curseOfGreed}, f: giveCurseHelper, ef: curseOfGreedEvent},
		monsterCard{baseCard: baseCard{name: "Curse of Loss", id: curseOfLoss}, f: giveCurseHelper},
		monsterCard{baseCard: baseCard{name: "Curse of Pain", id: curseOfPain}, f: giveCurseHelper, ef: curseOfPainEvent},
		monsterCard{baseCard: baseCard{name: "Curse of the Blind", id: curseOfTheBlind}, f: giveCurseHelper, ef: curseOfTheBlindEvent},
	}...)
	if useExpansionOne == true {
		deck.append([]card{
			monsterCard{baseCard: baseCard{name: "Begotten", id: begotten}, baseHealth: 3, baseRoll: 4, baseAttack: 1},
			monsterCard{baseCard: baseCard{name: "Boil", id: boil}, baseHealth: 2, baseRoll: 4, baseAttack: 1},
			monsterCard{baseCard: baseCard{name: "Charger", id: charger}, baseHealth: 1, baseRoll: 5, baseAttack: 1},
			monsterCard{baseCard: baseCard{name: "Death's Head", id: deathsHead}, baseHealth: 2},
			monsterCard{baseCard: baseCard{name: "Gaper", id: gaper}, baseHealth: 2, baseRoll: 4, baseAttack: 1},
			monsterCard{baseCard: baseCard{name: "Imp", id: imp}, baseHealth: 3, baseRoll: 5, baseAttack: 1},
			monsterCard{baseCard: baseCard{name: "Knight", id: knight}, baseHealth: 2, baseRoll: 6, baseAttack: 1},
			monsterCard{baseCard: baseCard{name: "Parabite", id: parabite}, baseHealth: 2, baseRoll: 3, baseAttack: 1},
			monsterCard{baseCard: baseCard{name: "Ragling", id: ragling}, baseHealth: 2, baseRoll: 3, baseAttack: 1},
			monsterCard{baseCard: baseCard{name: "Round Worm", id: roundWorm}, baseHealth: 1, baseRoll: 5, baseAttack: 1},
			monsterCard{baseCard: baseCard{name: "Fistula", id: fistula}, baseHealth: 4, baseRoll: 2, baseAttack: 1, isBoss: true},
			monsterCard{baseCard: baseCard{name: "Gurglings", id: gurglings}, baseHealth: 4, baseRoll: 5, baseAttack: 1, isBoss: true},
			monsterCard{baseCard: baseCard{name: "Polycephalus", id: polycephalus}, baseHealth: 3, baseRoll: 3, baseAttack: 1, isBoss: true},
			monsterCard{baseCard: baseCard{name: "Steven", id: steven}, baseHealth: 4, baseRoll: 2, baseAttack: 1, isBoss: true},
			monsterCard{baseCard: baseCard{name: "The Cage", id: theCage}, baseHealth: 8, baseRoll: 3, baseAttack: 1, isBoss: true},
			monsterCard{baseCard: baseCard{name: "!HUSH!", id: hush}, baseHealth: 8, baseRoll: 3, baseAttack: 1, isBoss: true},
			monsterCard{baseCard: baseCard{name: "I Am Error!", id: iAmError}},
			monsterCard{baseCard: baseCard{name: "Trap Door!", id: trapDoor}},
			monsterCard{baseCard: baseCard{name: "Curse of Fatigue", id: curseOfFatigue}},
			monsterCard{baseCard: baseCard{name: "Curse of Tiny Hands", id: curseOfTinyHands}},
		}...)
	}
	if useExpansionTwo == true {
		deck.append([]card{
			monsterCard{baseCard: baseCard{name: "Bony", id: bony}, baseHealth: 2, baseRoll: 3, baseAttack: 1},
			monsterCard{baseCard: baseCard{name: "Brain", id: brain}, baseHealth: 2, baseRoll: 3, baseAttack: 1},
			monsterCard{baseCard: baseCard{name: "Flaming Hopper", id: flaminHopper}, baseHealth: 1, baseRoll: 4, baseAttack: 2},
			monsterCard{baseCard: baseCard{name: "Globin", id: globin}, baseHealth: 4, baseRoll: 4, baseAttack: 1},
			monsterCard{baseCard: baseCard{name: "Nerve Ending", id: nerveEnding}, baseHealth: 4, baseRoll: 2, baseAttack: 1},
			monsterCard{baseCard: baseCard{name: "Roundy", id: roundy}, baseHealth: 3, baseRoll: 4, baseAttack: 2},
			monsterCard{baseCard: baseCard{name: "Sucker", id: sucker}, baseHealth: 1, baseRoll: 3, baseAttack: 1},
			monsterCard{baseCard: baseCard{name: "Swarmer", id: swarmer}, baseHealth: 4, baseRoll: 3, baseAttack: 2},
			monsterCard{baseCard: baseCard{name: "Tumor", id: tumor}, baseHealth: 3, baseRoll: 4, baseAttack: 1},
			monsterCard{baseCard: baseCard{name: "Cursed Globin", id: cursedGlobin}, baseHealth: 3, baseRoll: 4, baseAttack: 1},
			monsterCard{baseCard: baseCard{name: "Cursed Tumor", id: cursedTumor}, baseHealth: 3, baseRoll: 4, baseAttack: 1},
			monsterCard{baseCard: baseCard{name: "Holy Bony", id: holyBony}, baseHealth: 1, baseRoll: 3, baseAttack: 1},
			monsterCard{baseCard: baseCard{name: "Holy Mulligan", id: holyMulligan}, baseHealth: 1, baseRoll: 3, baseAttack: 1},
			monsterCard{baseCard: baseCard{name: "Blastocyst", id: blastocyst}, baseHealth: 5, baseRoll: 4, baseAttack: 1, isBoss: true},
			monsterCard{baseCard: baseCard{name: "Dingle", id: dingle}, baseHealth: 3, baseRoll: 3, baseAttack: 1, isBoss: true},
			monsterCard{baseCard: baseCard{name: "Headless Horseman", id: headlessHorseman}, baseHealth: 5, baseRoll: 4, baseAttack: 1, isBoss: true},
			monsterCard{baseCard: baseCard{name: "Krampus", id: krampus}, baseHealth: 4, baseRoll: 4, baseAttack: 2, isBoss: true},
			monsterCard{baseCard: baseCard{name: "Monstro II", id: monstroII}, baseHealth: 5, baseRoll: 4, baseAttack: 1, isBoss: true},
			monsterCard{baseCard: baseCard{name: "The Fallen", id: theFallen}, baseHealth: 4, baseRoll: 5, baseAttack: 2, isBoss: true},
			monsterCard{baseCard: baseCard{name: "Widow", id: widow}, baseHealth: 3, baseRoll: 4, baseAttack: 1, isBoss: true},
			monsterCard{baseCard: baseCard{name: "Isaac!", id: isaacMonster}, baseHealth: 7, baseRoll: 3, baseAttack: 1, isBoss: true},
			monsterCard{baseCard: baseCard{name: "Mom's Heart!", id: momsHeart}, baseHealth: 8, baseRoll: 4, baseAttack: 2, isBoss: true},
			monsterCard{baseCard: baseCard{name: "Angel Room", id: angelRoom}},
			monsterCard{baseCard: baseCard{name: "Boss Rush!", id: bossRush}},
			monsterCard{baseCard: baseCard{name: "Head Trauma", id: headTrauma}},
			monsterCard{baseCard: baseCard{name: "Holy Chest", id: holyChest}},
			monsterCard{baseCard: baseCard{name: "Spiked Chest", id: spikedChest}},
			monsterCard{baseCard: baseCard{name: "Troll Bombs", id: trollBombs}},
			monsterCard{baseCard: baseCard{name: "Curse of Blood Lust", id: curseOfBloodLust}},
			monsterCard{baseCard: baseCard{name: "Curse of Impulse", id: curseOfImpulse}},
		}...)
	}
	return deck
}

// Get the starting cards for every character, EXCEPT for Eden, who has a special starting itemCard condition.
func getStartingCards() map[string]treasureCard {
	items := make(map[string]treasureCard)
	items["Blue Baby"] = treasureCard{baseCard: baseCard{name: "Forever Alone", effect: foreverAloneDesc, id: foreverAlone}, eternal: true, active: true, f: foreverAloneFunc}
	items["Cain"] = treasureCard{baseCard: baseCard{name: "Sleight Of Hand", effect: sleightOfHandDesc, id: sleightOfHand}, eternal: true, active: true, f: sleightOfHandFunc}
	items["Eve"] = treasureCard{baseCard: baseCard{name: "The Curse", effect: theCurseDesc, id: theCurse}, eternal: true, active: true, f: theCurseFunc}
	items["Isaac"] = treasureCard{baseCard: baseCard{name: "The D6", effect: theD6Desc, id: theD6}, eternal: true, active: true, f: theD6Func}
	items["Judas"] = treasureCard{baseCard: baseCard{name: "Book of Belial", effect: bookOfBelialDesc, id: bookOfBelial}, eternal: true, active: true, f: bookOfBelialFunc}
	items["Lazarus"] = treasureCard{baseCard: baseCard{name: "Lazarus' Rags", effect: lazarusRagsDesc, id: lazarusRags}, eternal: true, passive: true}
	items["Lilith"] = treasureCard{baseCard: baseCard{name: "Incubus", effect: incubusDesc, id: incubus}, eternal: true, active: true, f: incubusFunc}
	items["Maggy"] = treasureCard{baseCard: baseCard{name: "Yum Heart", effect: yumHeartDesc, id: yumHeart}, eternal: true, active: true, f: yumHeartFunc}
	items["Samson"] = treasureCard{baseCard: baseCard{name: "Blood Lust", effect: bloodLustDesc, id: bloodLust}, eternal: true, active: true, f: bloodLustFunc}
	items["The Forgotten"] = treasureCard{baseCard: baseCard{name: "The Bone", effect: theBoneDesc, id: theBone}, eternal: true, active: true, f: theBoneFunc}
	items["Apollyon"] = treasureCard{baseCard: baseCard{name: "Void", effect: voidDesc, id: void}, eternal: true, active: true}
	items["Azazel"] = treasureCard{baseCard: baseCard{name: "Lord of the Pit", effect: lordOfThePitDesc, id: lordOfThePit}, eternal: true, active: true}
	items["The Keeper"] = treasureCard{baseCard: baseCard{name: "Wooden Nickel", effect: woodenNickelDesc, id: woodenNickel}, eternal: true, active: true}
	items["The Lost"] = treasureCard{baseCard: baseCard{name: "Holy Mantle", effect: holyMantleDesc, id: theHolyMantle}, eternal: true, active: true}
	items["Dark Judas"] = treasureCard{baseCard: baseCard{name: "Dark Arts", effect: darkArtsDesc, id: darkArts}, eternal: true, passive: true}
	items["Guppy"] = treasureCard{baseCard: baseCard{name: "Infestation", effect: infestationDesc, id: infestation}, eternal: true, active: true}
	items["Whore of Babylon"] = treasureCard{baseCard: baseCard{name: "Gimpy", effect: gimpyDesc, id: gimpy}, eternal: true, passive: true}
	items["Bum-Bo"] = treasureCard{baseCard: baseCard{name: "Bag-O-Trash", effect: bagOTrashDesc, id: bagOTrash}, eternal: true, active: true}
	return items
}

// Get the treasure cards for the treasure deck as a slice.
func getTreasureCards(useExpansionOne bool, useExpansionTwo bool) deck {
	var n uint8 = 105
	if useExpansionOne {
		n += 20
	}
	if useExpansionTwo {
		n += 31
	}
	deck := make(deck, 0, n)
	deck.append([]card{
		treasureCard{baseCard: baseCard{name: "Blank Card", id: blankCard}, active: true, f: blankCardFunc},
		treasureCard{baseCard: baseCard{name: "Book of Sin", id: bookOfSin}, active: true, f: bookOfSinFunc},
		treasureCard{baseCard: baseCard{name: "Boomerang", id: boomerang}, active: true, f: boomerangeFunc},
		treasureCard{baseCard: baseCard{name: "Box!", id: box}, active: true, f: boxFunc},
		treasureCard{baseCard: baseCard{name: "Bum Friend", id: bumFriend}, active: true, f: bumFriendFunc},
		treasureCard{baseCard: baseCard{name: "Chaos", id: chaos}, active: true, f: chaosFunc},
		treasureCard{baseCard: baseCard{name: "Chaos Card", id: chaosCard}, active: true, f: chaosCardFunc},
		treasureCard{baseCard: baseCard{name: "Compost", id: compost}, active: true, f: compostFunc},
		treasureCard{baseCard: baseCard{name: "Crystal Ball", id: crystalBall}, active: true, f: crystalBallFunc},
		treasureCard{baseCard: baseCard{name: "Decoy", id: decoy}, active: true, f: decoyFunc},
		treasureCard{baseCard: baseCard{name: "Diplopia", id: diplopia}, active: true, f: diplopiaFunc},
		treasureCard{baseCard: baseCard{name: "Flush!", id: flush}, active: true, f: flushFunc},
		treasureCard{baseCard: baseCard{name: "Glass Cannon", id: glassCannon}, active: true, f: glassCannonFunc},
		treasureCard{baseCard: baseCard{name: "Godhead", id: godhead}, active: true, f: godheadFunc},
		treasureCard{baseCard: baseCard{name: "Guppy's Head", id: guppysHead}, active: true, f: guppysHeadFunc},
		treasureCard{baseCard: baseCard{name: "Guppy's Paw", id: guppysPaw}, active: true, f: guppysPawFunc},
		treasureCard{baseCard: baseCard{name: "Host Hat", id: hostHat}, active: true, f: hostHatFunc},
		treasureCard{baseCard: baseCard{name: "Jawbone", id: jawbone}, active: true, f: jawboneFunc},
		treasureCard{baseCard: baseCard{name: "Lucky Foot", id: luckyFoot}, active: true, f: luckyFootFunc},
		treasureCard{baseCard: baseCard{name: "Mini Mush", id: miniMush}, active: true, f: miniMushFunc},
		treasureCard{baseCard: baseCard{name: "Modeling Clay", id: modelingClay}, active: true, f: modelingClayFunc},
		treasureCard{baseCard: baseCard{name: "Mom's Bra", id: momsBra}, active: true, f: momsBraFunc},
		treasureCard{baseCard: baseCard{name: "Mom's Shovel", id: momsShovel}, active: true, tapped: true, f: momsShovelFunc},
		treasureCard{baseCard: baseCard{name: "Monster Manual", id: monsterManual}, active: true, f: monsterManualFunc},
		treasureCard{baseCard: baseCard{name: "Mr. Boom", id: mrBoom}, active: true, f: mrBoomFunc},
		treasureCard{baseCard: baseCard{name: "Mystery Sack", id: mysterySack}, active: true, f: mysterySackFunc},
		treasureCard{baseCard: baseCard{name: "No!", id: no}, active: true, f: noFunc},
		treasureCard{baseCard: baseCard{name: "Pandora's Box", id: pandorasBox}, active: true, f: pandorasBoxFunc},
		treasureCard{baseCard: baseCard{name: "Placebo", id: placebo}, active: true, f: placeboFunc},
		treasureCard{baseCard: baseCard{name: "Potato Peeler", id: potatoPeeler}, active: true, f: potatoPeelerFunc},
		treasureCard{baseCard: baseCard{name: "Razor Blade", id: razorBlade}, active: true, f: razorBladeFunc},
		treasureCard{baseCard: baseCard{name: "Remote Detonator", id: remoteDetonator}, active: true, f: remoteDetonatorFunc},
		treasureCard{baseCard: baseCard{name: "Sack Head", id: sackHead}, active: true, f: sackHeadFunc},
		treasureCard{baseCard: baseCard{name: "Sack of Pennies", id: sackOfPennies}, active: true, f: sackOfPenniesFunc},
		treasureCard{baseCard: baseCard{name: "Spoon Bender", id: spoonBender}, active: true, f: spoonBenderFunc},
		treasureCard{baseCard: baseCard{name: "The Battery", id: theBattery}, active: true, f: theBatteryFunc},
		treasureCard{baseCard: baseCard{name: "The D4", id: theD4}, active: true, f: theD4Func},
		treasureCard{baseCard: baseCard{name: "The D20", id: theD20}, active: true, f: theD20Func},
		treasureCard{baseCard: baseCard{name: "The D100", id: theD100}, active: true, f: theD100Func},
		treasureCard{baseCard: baseCard{name: "The Shovel", id: theShovel}, active: true, f: theShovelFunc},
		treasureCard{baseCard: baseCard{name: "Two of Clubs", id: twoOfClubs}, active: true, f: twoOfClubsFunc},
		treasureCard{baseCard: baseCard{name: "Battery Bum", id: batteryBum}, paid: true, f: batteryBumFunc},
		treasureCard{baseCard: baseCard{name: "Contract From Below", id: contractFromBelow}, paid: true, f: contractFromBelowFunc},
		treasureCard{baseCard: baseCard{name: "Donation Machine", id: donationMachine}, paid: true, f: donationMachineFunc},
		treasureCard{baseCard: baseCard{name: "Golden Razor Blade", id: goldenRazorBlade}, paid: true, f: goldenRazorBladeFunc},
		treasureCard{baseCard: baseCard{name: "Pay To Play", id: payToPlay}, paid: true, f: payToPlayFunc},
		treasureCard{baseCard: baseCard{name: "Portable Slot Machine", id: portableSlotMachine}, paid: true, f: portableSlotMachineFunc},
		treasureCard{baseCard: baseCard{name: "Smelter", id: smelter}, paid: true, f: smelterFunc},
		treasureCard{baseCard: baseCard{name: "The Poop", id: thePoop}, paid: true, f: thePoopFunc},
		treasureCard{baseCard: baseCard{name: "Tech X", id: techX}, active: true, paid: true, f: techXFunc},
		treasureCard{baseCard: baseCard{name: "Baby Haunt", id: babyHaunt}, passive: true, ef: babyHauntFunc},
		treasureCard{baseCard: baseCard{name: "Belly Button", id: bellyButton}, passive: true, ef: bellyButtonFuncEvent, cf: bellyButtonFuncConstant},
		treasureCard{baseCard: baseCard{name: "Bob's Brain", id: bobsBrain}, passive: true, ef: bobsBrainFunc},
		treasureCard{baseCard: baseCard{name: "Breakfast", id: breakfast}, passive: true, cf: breakfastFunc},
		treasureCard{baseCard: baseCard{name: "Brimstone", id: brimstone}, passive: true, ef: brimstoneFuncEvent, cf: brimstoneFuncConstant},
		treasureCard{baseCard: baseCard{name: "Bum-Bo!", id: bumbo}, passive: true, cf: bumboFunc},
		treasureCard{baseCard: baseCard{name: "Cambion Conception", id: cambionConception}, passive: true, ef: cambionConceptionFunc},
		treasureCard{baseCard: baseCard{name: "Champion Belt", id: championBelt}, passive: true, cf: championBeltFunc},
		treasureCard{baseCard: baseCard{name: "Charged Baby", id: chargedBaby}, passive: true, ef: chargedBabyFunc},
		treasureCard{baseCard: baseCard{name: "Cheese Grater", id: cheeseGrater}, passive: true, ef: cheeseGraterFunc},
		treasureCard{baseCard: baseCard{name: "Curse of the Tower", id: curseOfTheTower}, passive: true, ef: curseOfTheTowerFunc},
		treasureCard{baseCard: baseCard{name: "Dad's Lost Coin", id: dadsLostCoint}, passive: true, ef: dadsLostCoinFunc},
		treasureCard{baseCard: baseCard{name: "Daddy Haunt", id: daddyHaunt}, passive: true, ef: daddyHauntFunc},
		treasureCard{baseCard: baseCard{name: "Dark Bum", id: darkBum}, passive: true, ef: darkBumFunc},
		treasureCard{baseCard: baseCard{name: "Dead Bird", id: deadBird}, passive: true, ef: deadBirdFunc},
		treasureCard{baseCard: baseCard{name: "Dinner", id: dinner}, passive: true, cf: dinnerFunc},
		treasureCard{baseCard: baseCard{name: "Dry Baby", id: dryBaby}, passive: true}, // No function need be attached
		treasureCard{baseCard: baseCard{name: "Eden's Blessing", id: edensBlessing}, passive: true, ef: edensBlessingFunc},
		treasureCard{baseCard: baseCard{name: "Empty Vessel", id: emptyVessel}, passive: true}, // No function need be attached
		treasureCard{baseCard: baseCard{name: "Eye of Greed", id: eyeOfGreed}, passive: true, ef: eyeOfGreedFunc},
		treasureCard{baseCard: baseCard{name: "Fanny Pack", id: fannyPack}, passive: true, ef: fannyPackFunc},
		treasureCard{baseCard: baseCard{name: "Finger", id: finger}, passive: true, ef: fingerFunc},
		treasureCard{baseCard: baseCard{name: "Greed's Gullet", id: greedsGullet}, passive: true, ef: greedsGulletFunc},
		treasureCard{baseCard: baseCard{name: "Goat Head", id: goatHead}, passive: true, ef: goatHeadFunc},
		treasureCard{baseCard: baseCard{name: "Guppy's Collar", id: guppysCollar}, passive: true, ef: guppysCollarFunc},
		treasureCard{baseCard: baseCard{name: "Ipecac", id: ipecac}, passive: true, ef: ipecacFuncEvent, cf: ipecacFuncConstant},
		treasureCard{baseCard: baseCard{name: "Meat!", id: meat}, passive: true}, // No function need be attached
		treasureCard{baseCard: baseCard{name: "Mom's Box", id: momsBox}, passive: true, ef: momsBoxFunc},
		treasureCard{baseCard: baseCard{name: "Mom's Coin Purse", id: momsCoinPurse}, passive: true, ef: momsPursesFunc},
		treasureCard{baseCard: baseCard{name: "Mom's Purse", id: momsPurse}, passive: true, ef: momsPursesFunc},
		treasureCard{baseCard: baseCard{name: "Mom's Razor", id: momsRazor}, passive: true, ef: momsRazorFunc},
		treasureCard{baseCard: baseCard{name: "Monstro's Tooth", id: monstrosTooth}, passive: true, ef: monstrosToothFunc},
		treasureCard{baseCard: baseCard{name: "Polydactyly", id: polydactyly}, passive: true, cf: polydactylyFunc},
		treasureCard{baseCard: baseCard{name: "Restock", id: restock}, passive: true, ef: restockFunc},
		treasureCard{baseCard: baseCard{name: "Sacred Heart", id: sacredHeart}, passive: true, ef: sacredHeartFunc},
		treasureCard{baseCard: baseCard{name: "Shadow", id: shadow}, passive: true}, // No function need be attached
		treasureCard{baseCard: baseCard{name: "Shiny Rock", id: shinyRock}, passive: true, ef: shinyRockFunc},
		treasureCard{baseCard: baseCard{name: "Spider Mod", id: spiderMod}, passive: true, ef: spiderModFunc},
		treasureCard{baseCard: baseCard{name: "Starter Deck", id: starterDeck}, passive: true, ef: starterDeckFunc},
		treasureCard{baseCard: baseCard{name: "Steamy Sale!", id: steamySale}, passive: true}, // No function need be attached
		treasureCard{baseCard: baseCard{name: "Suicide King", id: suicideKing}, passive: true, ef: suicideKingFunc},
		treasureCard{baseCard: baseCard{name: "Synthoil", id: synthoil}, passive: true}, // No function need be attached
		treasureCard{baseCard: baseCard{name: "Tarot Cloth", id: tarotCloth}, passive: true, ef: tarotClothFunc},
		treasureCard{baseCard: baseCard{name: "There's Options", id: theresOptions}, passive: true, cf: theresOptionsFunc},
		treasureCard{baseCard: baseCard{name: "The Blue Map", id: theBlueMap}, passive: true, ef: theBlueMapFunc},
		treasureCard{baseCard: baseCard{name: "The Chest", id: theChest}, passive: true, cf: theChestFunc},
		treasureCard{baseCard: baseCard{name: "The Compass", id: theCompass}, passive: true, ef: theCompassFunc},
		treasureCard{baseCard: baseCard{name: "The D10", id: theD10}, passive: true, ef: theD10Func},
		treasureCard{baseCard: baseCard{name: "The Dead Cat", id: theDeadCat}, passive: true, cf: theDeadCatFuncConstant},
		treasureCard{baseCard: baseCard{name: "The Habit", id: theHabit}, passive: true, ef: theHabitFuncEvent, cf: theHabitFuncConstant},
		treasureCard{baseCard: baseCard{name: "The Map", id: theMap}, passive: true, ef: theMapFunc},
		treasureCard{baseCard: baseCard{name: "The Midas Touch", id: theMidasTouch}, passive: true}, // no function need be assigned
		treasureCard{baseCard: baseCard{name: "The Polaroid", id: thePolaroid}, passive: true, ef: thePolaroidFunc},
		treasureCard{baseCard: baseCard{name: "The Relic", id: theRelic}, passive: true, ef: theRelicFunc},
		treasureCard{baseCard: baseCard{name: "Trinity Shield", id: trinityShield}, passive: true},
	}...)
	if useExpansionOne == true {
		deck.append([]card{
			treasureCard{baseCard: baseCard{name: "Crooked Penny", id: crookedPenny}, active: true},
			treasureCard{baseCard: baseCard{name: "Fruitcake", id: fruitCake}, active: true},
			treasureCard{baseCard: baseCard{name: "I Can't Believe it's Not Butter Bean", id: iCantBelieveItsNotButterBean}, active: true},
			treasureCard{baseCard: baseCard{name: "Lemon Mishap", id: lemonMishap}, active: true},
			treasureCard{baseCard: baseCard{name: "Library Card", id: libraryCard}, active: true},
			treasureCard{baseCard: baseCard{name: "Ouija Board", id: ouijaBoard}, active: true},
			treasureCard{baseCard: baseCard{name: "Plan C", id: planC}, active: true},
			treasureCard{baseCard: baseCard{name: "The Bible", id: theBible}, active: true},
			treasureCard{baseCard: baseCard{name: "The Butter Bean", id: theButterBean}, active: true},
			treasureCard{baseCard: baseCard{name: "Dad's Key", id: dadsKey}, paid: true},
			treasureCard{baseCard: baseCard{name: "Succubus", id: succubus}, paid: true},
			treasureCard{baseCard: baseCard{name: "9 Volt", id: nineVolt}, passive: true},
			treasureCard{baseCard: baseCard{name: "Guppy's Tail", id: guppysTail}, passive: true},
			treasureCard{baseCard: baseCard{name: "Infamy", id: infamy}, passive: true},
			treasureCard{baseCard: baseCard{name: "Mom's Knife", id: momsKnife}, passive: true},
			treasureCard{baseCard: baseCard{name: "More Options", id: moreOptions}, passive: true},
			treasureCard{baseCard: baseCard{name: "Placenta", id: placenta}, passive: true},
			treasureCard{baseCard: baseCard{name: "Skeleton Key", id: skeletonKey}, passive: true},
			treasureCard{baseCard: baseCard{name: "Soy Milk", id: soyMilk}, passive: true},
			treasureCard{baseCard: baseCard{name: "The Missing Page", id: theMissingPage}, passive: true},
		}...)
	}
	if useExpansionTwo == true {
		deck.append([]card{
			treasureCard{baseCard: baseCard{name: "20/20", id: twentyTwenty}, active: true},
			treasureCard{baseCard: baseCard{name: "Black Candle", id: blackCandle}, active: true},
			treasureCard{baseCard: baseCard{name: "Distant Admiration", id: distantAdmiration}, active: true},
			treasureCard{baseCard: baseCard{name: "Divorce Papers", id: divorcePapers}, active: true},
			treasureCard{baseCard: baseCard{name: "Forget Me Now", id: forgetMeNow}, active: true},
			treasureCard{baseCard: baseCard{name: "Head of Krampus", id: headOfKrampus}, active: true},
			treasureCard{baseCard: baseCard{name: "Infestation", id: infestation}, active: true},
			treasureCard{baseCard: baseCard{name: "Libra", id: libra}, active: true},
			treasureCard{baseCard: baseCard{name: "Mutant Spider", id: mutantSpider}, active: true},
			treasureCard{baseCard: baseCard{name: "Rainbow Baby", id: rainbowBaby}, active: true},
			treasureCard{baseCard: baseCard{name: "Red Candle", id: redCandle}, active: true},
			treasureCard{baseCard: baseCard{name: "Smart Fly", id: smartFly}, active: true, f: smartFlyFunc},
			treasureCard{baseCard: baseCard{name: "Athame", id: athame}, paid: true},
			treasureCard{baseCard: baseCard{name: "1-Up", id: oneUp}, passive: true},
			treasureCard{baseCard: baseCard{name: "Abaddon", id: abaddon}, passive: true},
			treasureCard{baseCard: baseCard{name: "Cursed Eye", id: cursedEye}, passive: true},
			treasureCard{baseCard: baseCard{name: "Daddy Long Legs", id: daddyLongLegsTreasure}, passive: true},
			treasureCard{baseCard: baseCard{name: "Euthanasia", id: euthanasia}, passive: true},
			treasureCard{baseCard: baseCard{name: "Game Breaking Bug!", id: gameBreakingBug}, passive: true},
			treasureCard{baseCard: baseCard{name: "Guppy's Eye", id: guppysEye}, passive: true},
			treasureCard{baseCard: baseCard{name: "Head of the Keeper", id: headOfTheKeeper}, passive: true},
			treasureCard{baseCard: baseCard{name: "Hourglass", id: hourGlass}, passive: true},
			treasureCard{baseCard: baseCard{name: "Lard", id: lard}, passive: true},
			treasureCard{baseCard: baseCard{name: "Magnet", id: magnet}, passive: true},
			treasureCard{baseCard: baseCard{name: "Mama Haunt", id: mamaHaunt}, passive: true},
			treasureCard{baseCard: baseCard{name: "Mom's Eye Shadow", id: momsEyeShaow}, passive: true},
			treasureCard{baseCard: baseCard{name: "P.H.D", id: phd}, passive: true},
			treasureCard{baseCard: baseCard{name: "Polyphemus", id: polyphemus}, passive: true},
			treasureCard{baseCard: baseCard{name: "Rubber Cement", id: rubberCement}, passive: true},
			treasureCard{baseCard: baseCard{name: "Telepathy For Dummies", id: telepathyForDummies}, passive: true},
			treasureCard{baseCard: baseCard{name: "The Wiz", id: theWiz}, passive: true},
		}...)
	}
	return deck
}
