/*
This file will host all the activated and resolved effects of every treasure value
in the game. These cards are all activated from the player's hand.

There are three treasure value types:
1) Active: Cards that can be tapped (turned to the side) during any player's turn at any time provided
activation conditions have been properly met. Can only be used again once recharged thanks to an item
or waiting till the beginning of the player's next turn.
2) Paid: Cards that do not have to be tapped for activation, but require some kind of cost to trigger their
effects (pay cents, destroy items, discard a value)
3) Passive: Cards that are not directly activated and continuously alter the game state.

Broadly speaking there are two types of passive items:
1) Event-based: Require some external event to occur prior to triggering its effect (dice roll, damage, start / end of turn)
	- These effects will trigger upon some successfully resolved event
	- Ex: board's roll variable being set, damage / deathPenalty events prior to inflicting damage / deathPenalty, start / end of turn
2) Constant: Provides a constant game changing effect until the value leaves play
	- These can be cards that alter the game's stats or change a normal behavior to something else
	- Not all of the cards in the game will have an assigned / defined effect below
		- Cards like Charged Penny, Dry Baby, Empty Vessel, that have conditions to their constant effects
			will have their effects effects checked elsewhere

Only event-based passive effects can be pushed to the stack and are not directly activated by
the player. All passive value effect functions in this file are either event based passives,
or are a hybrid of event based and constant passives. All other passive cards that do NOT push
an event to the stack will be in a separate file.

*/

package four_souls

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// Event dependent passive card
// All Monsters you attack gain +1 Dice Roll
// When you die, before paying penalties, give this card to another Player
//
// The latter part of this effect is handled upon player deathPenalty.
func babyHauntFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	var intention intentionToAttackEvent
	if intention, err = en.checkIntentionToAttack(); err == nil && p.getId() == en.event.p.getId() {
		f = func(roll uint8) { intention.m.modifyDiceRoll(1) }
	}
	return f, false, err
}

// Paid Item
// Pay 4 Cents: Recharge an Item.
func batteryBumFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	if p.Pennies < 4 {
		return nil, false, errors.New("not enough cents to pay")
	}
	p.loseCents(4)
	items := p.getTappedActiveItems()
	l := len(items)
	if l == 0 {
		return nil, false, errors.New("no items have been tapped")
	}
	var i uint8
	if l > 1 {
		showTreasureCards(items, "self", 0)
		i = uint8(readInput(0, l-1))
	}
	return func(roll uint8) { p.rechargeActiveItemById(items[i].id) }, false, nil
}

// Hybrid passive item
// You may play an additional loot card on your turn
// Each time you take damage, you may recharge your Character card. <- This one
func bellyButtonFuncEvent(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if _, err = en.checkDamageToPlayer(p.Character.id); err == nil && p.Character.tapped {
		fmt.Println("1) Recharge character value\n2) Do not.")
		if uint8(readInput(1, 2)) == 1 {
			f = func(roll uint8) { p.Character.recharge() }
		}
	} else {
		err = errors.New("conditions not met")
	}
	return f, false, err
}

// Hybrid passive item
// You may play an additional loot card on your turn <-
// Each time you take damage, you may recharge your Character card.
func bellyButtonFuncConstant(p *player, b *Board, c card, leavingField bool) {
	if !leavingField {
		p.numLootPlayed += 1
		p.baseNumLootPlayed += 1
	} else {
		if p.numLootPlayed > 0 {
			p.numLootPlayed -= 1
		}
		p.baseNumLootPlayed -= 1
	}
}

// Active Item
// Double the effect of the next Loot Card you play.
func blankCardFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	return func(roll uint8) { p.activeEffects[blankCard] = struct{}{} }, false, nil
}

// Starting Item (Samson)
// Active Item
// Add + 1 Attack to a Player or Monster till the end of the turn.
func bloodLustFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	players := b.getPlayers(true)
	l := len(players)
	monsters := b.monster.getActiveMonsters()
	showPlayers(players, 0)
	showMonsterCards(monsters, l)
	ans := readInput(0, l+len(monsters)-1)
	var c combatTarget
	if ans < l {
		c = players[ans]
	} else {
		c = monsters[ans-l]
	}
	var f cardEffect = func(roll uint8) {
		c.increaseAP(1)
	}
	return f, false, nil
}

// Event Based Passive item (Declare Attack)
// When you start an attack, roll:
// 1-2: Deal 1 damage to an active monster
// 3-4: Deal 1 damage to a Player
// 5-6: Deal 1 damage to yourself
func bobsBrainFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err = errors.New("not a start to an attack")
	if _, err := en.checkDeclareAttack(p); err == nil {
		err, f = nil, func(roll uint8) {
			if roll == 1 || roll == 2 {
				monsters := b.monster.getActiveMonsters()
				showMonsterCards(monsters, 0)
				ans := uint8(readInput(0, len(monsters)-1))
				b.damagePlayerToMonster(p, monsters[ans], 1, 0)
			} else if roll == 3 || roll == 4 {
				players := b.getPlayers(true)
				l := len(players)
				var ans uint8
				if l > 1 {
					showPlayers(players, 0)
					ans = uint8(readInput(0, l-1))
				}
				b.damagePlayerToPlayer(p, players[ans], 1)
			} else {
				b.damagePlayerToPlayer(p, p, 1)
			}
		}
	}
	return f, true, err
}

// Starting Item (Judas)
// Active Item
// Add or subtractUint8 1 from any Dice roll.
func bookOfBelialFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	var ans uint8
	var n int8 = 1
	rolls := b.eventStack.getDiceRollEvents()
	l := len(rolls)
	if l == 0 {
		return nil, false, errors.New("no dice rolls")
	} else if l > 1 {
		showEvents(rolls)
		ans = uint8(readInput(0, len(rolls)-1))
	}
	node := rolls[ans]
	fmt.Println("1) Add 1 to the roll.\n2) Subtract 1 from the roll.")
	ans = uint8(readInput(1, 2))
	if ans == 2 {
		n = -1
	}
	var f cardEffect = func(roll uint8) { b.eventStack.addToDiceRoll(n, node) }
	return f, false, nil
}

// Active Item
// Roll:
// 1 - 2: Gain 1 cent.
// 3 - 4: Loot 1.
// 5 - 6: Gain +1 HP till the end of the turn.
func bookOfSinFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	loot := b.loot
	var f cardEffect = func(roll uint8) {
		if roll == 1 || roll == 2 {
			p.gainCents(1)
		} else if roll == 3 || roll == 4 {
			p.loot(loot)
		} else if roll == 5 || roll == 6 {
			p.increaseHP(1)
		}
	}
	return f, true, nil
}

// Active Item
// Steal a loot card at random from a player.
func boomerangeFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	players := b.getOtherPlayers(p, false)
	l := len(players)
	var i uint8
	if l > 1 {
		showPlayers(players, 0)
		i = uint8(readInput(0, l-1))
	}
	p2 := players[i]
	var f cardEffect = func(roll uint8) {
		if len(p2.Hand) > 0 {
			c := p2.popHandCard(uint8(rand.Intn(len(p2.Hand))))
			p.Hand = append(p.Hand, c)
		}
	}
	return f, false, nil
}

// Active Item
// Destroy this, You can play as many additional additional loot cards
// as you want till the end of turn.
func boxFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	searchId := box
	if tCard.getId() == placebo {
		searchId = placebo
	}
	i, err := p.getItemIndex(searchId, false)
	if err != nil {
		return nil, false, err
	}
	b.treasure.discard(&p.ActiveItems[i])
	return func(roll uint8) { p.numLootPlayed = 255 }, false, nil
}

// Constant Passive Item
// +1 HP
func breakfastFunc(p *player, b *Board, c card, leavingField bool) {
	modifyHealth(p, 1, leavingField)
}

// Hybrid Passive Item
// +1 Attack
// Each Time you deal damage to a monster, also deal 1 damage to another player
func brimstoneFuncEvent(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	_, err := en.checkDamageToMonster()
	if err == nil && en.event.p.Character.id == p.Character.id {
		others := b.getOtherPlayers(p, true)
		l := len(others)
		if l > 0 {
			var i uint8
			if l > 1 {
				showPlayers(others, 0)
				fmt.Println("Choose who to inflict damage to")
				i = uint8(readInput(0, l-1))
			}
			err, f = nil, func(roll uint8) { b.damagePlayerToPlayer(p, others[i], 1) }
		}
	}
	return f, false, err
}

// Hybrid Passive Item
// +1 Attack
// Each time you deal damage to a monster, also deal 1 damage to another player
func brimstoneFuncConstant(p *player, b *Board, tCard card, isLeaving bool) {
	modifyAttack(p, 1, isLeaving)
}

// Constant Passive Item
// If you would gain cents, instead put that many counters on this.
// 1+: Add +2 to your first attack roll each turn.
// 10+: Gain +1 Attack
// 25+: You may attack any number of times this turn.
func bumboFunc(p *player, b *Board, tCard card, isLeaving bool) {
	if isLeaving {
		t := tCard.(treasureCard)
		if t.counters >= 25 {
			p.numAttacks -= 99
			p.baseNumAttacks -= 99
		}
		if t.counters >= 10 {
			modifyAttack(p, 1, isLeaving)
		}
	}
}

// Active Item
// Loot 1: Put a loot card from your hand on top of the deck.
func bumFriendFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	var f cardEffect = func(roll uint8) {
		p.loot(b.loot)
		showLootCards(p.Hand, "self", 0)
		fmt.Println("Which to place on top of deck?")
		ans := readInput(0, len(p.Hand)-1)
		c := p.popHandCard(uint8(ans))
		b.loot.placeInDeck(c, true)
	}
	return f, false, nil
}

// Event based Passive Item
// Each time you take damage, put a counter on this.
// Whenever this has 6 counters on it, remove them and gain +1 treasure
func cambionConceptionFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if _, err = en.checkDamageToPlayer(p.Character.id); err == nil {
		t := tCard.(*treasureCard)
		f = func(roll uint8) {
			t.counters += 1
			if t.counters == 6 {
				t.counters = 0
				p.addCardToBoard(b.treasure.draw())
			}
		}
	}
	return f, false, err
}

// Constant Passive item
// Gain +1 Attack for the first attack roll of your turn
// You may attack an additional time
func championBeltFunc(p *player, b *Board, tCard card, isLeaving bool) {
	if !isLeaving {
		p.numAttacks += 1
		p.baseNumAttacks += 1
	} else {
		p.numAttacks -= 1
		p.baseNumAttacks -= 1
	}
}

// Active Item
// Each player gives all of their loot cards to the player to the left.
func chaosFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	var f cardEffect = func(roll uint8) {
		orderedPlayers := b.getPlayers(false)
		l := len(orderedPlayers)
		hands := make([][]lootCard, l)
		for i, p := range orderedPlayers {
			hands[i] = p.Hand
		}
		temp := hands[0]
		copy(hands[0:], hands[1:])
		hands[l-1] = temp
		for i, h := range hands {
			orderedPlayers[i].Hand = h
		}
	}
	return f, false, nil
}
func (b *Board) chaos() {

}

// Active Item
// Destroy this: Destroy any Monster, Player, Item, or Soul Card.
func chaosCardFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	searchId := chaosCard
	if tCard.getId() == placebo {
		searchId = placebo
	}
	i, err := p.getItemIndex(searchId, false)
	if err != nil {
		return nil, false, err
	}
	b.treasure.discard(&p.ActiveItems[i])
	var f cardEffect = func(roll uint8) {
		fmt.Println("1) Kill a monster / player.\n2) Destroy an treasure or soul value.")
		ans := readInput(1, 2)
		if ans == 1 {
			monsters, characters := b.monster.getActiveMonsters(), b.getCharacters(true)
			l := len(monsters)
			showMonsterCards(monsters, 0)
			showCharacterCards(b.getPlayers(false), l)
			i := uint8(readInput(0, l-1))
			if i < uint8(l) {
				b.killMonster(p, monsters[i].id)
			} else {
				target, _ := b.getPlayerFromCharacterId(characters[i-uint8(l)].id)
				b.killPlayer(target)
			}
		} else {
			for _, p2 := range b.getPlayers(false) {
				items := p2.getAllItems(false)
				l := len(items)
				fmt.Println("0) Continue to next player.")
				showItems(items, 0)
				showSouls(p2.Souls, p2.Character.name, l)
				ans := readInput(0, l+len(p2.Souls))
				var c card
				if ans > 0 && ans < l {
					id, isPassive := items[ans-1].getId(), items[ans-1].isPassive()
					j, _ := p2.getItemIndex(id, isPassive)
					c = p2.popItemByIndex(j, isPassive)
				} else {
					c = p2.popSoul(uint8(ans - l - 1))
				}
				b.discard(c)
				break
			}
		}
	}
	return f, false, nil
}

// Event based passive item
// When anyone rolls a 2, you may recharge an item.
func chargedBabyFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(2); err == nil {
		tappedItems := p.getTappedActiveItems()
		l := len(tappedItems)
		if l > 0 {
			fmt.Println("1) Recharge an item?\n2) Do nothing.")
			ans := uint8(readInput(1, 2))
			if ans == 1 {
				var i uint8
				if l > 1 {
					showTreasureCards(tappedItems, "self", 0)
					i = uint8(readInput(0, l-1))
				}
				f = func(roll uint8) { p.rechargeActiveItemById(tappedItems[i].id) }
			}
		}
	}
	return f, false, err
}

// Event based passive item
// When anyone rolls a 6, reveal the top card of any deck to all players.
// You may discard it or put it back on top.
func cheeseGraterFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(6); err == nil {
		f = func(roll uint8) {
			fmt.Print("1) Loot Deck\n2)Monster Deck\n3) Treasure Deck")
			ans := readInput(1, 3)
			var c card
			if ans == 1 {
				c = b.loot.draw()
			} else if ans == 2 {
				c = b.monster.draw()
			} else {
				c = b.treasure.draw()
			}
			c.showCard(0)
			fmt.Println("1) Discard this value?\n2) Place back on top.")
			ans = readInput(1, 2)
			if ans == 1 {
				b.discard(c)
			} else {
				b.placeInDeck(c, true)
			}
		}
	}
	return f, false, err
}

// Active Item
// The next time a player would loot, they loot from the top
// of the loot deck's discard pile instead.
func compostFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	return func(roll uint8) { b.loot.activeEffects[compost] = struct{}{} }, false, nil
}

// Paid Item
// Destroy 2 items you own:
// Steal an Item from any Player.
func contractFromBelowFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	items := p.getAllItems(false)
	l := len(items)
	if l < 2 {
		return nil, false, errors.New("not enough items to destroy")
	}
	var i uint8
	showItems(items, 0)
	fmt.Println("Pick two cards to destroy")
	toDestroy := make(map[uint8]struct{}, 2)
	for i < 2 {
		ans := uint8(readInput(0, l-1))
		if _, ok := toDestroy[ans]; !ok {
			toDestroy[ans] = struct{}{}
			i += 1
		} else {
			fmt.Println("Already chose this one.")
		}
	}
	for k := range toDestroy {
		id, isPassive := items[k].getId(), items[k].isPassive()
		idx, _ := p.getItemIndex(id, isPassive)
		b.discard(p.popItemByIndex(idx, isPassive))
	}
	var owners map[uint16]*player
	items, owners = b.getAllItems(false, p)
	showItems(items, 0)
	fmt.Println("Which to steal?")
	ans := uint8(readInput(0, len(items)-1))
	return func(roll uint8) {
		id, isPassive := items[ans].getId(), items[ans].isPassive()
		p.stealItem(id, isPassive, owners[id])
	}, false, nil
}

// Event based passive item
// When you take damage, roll:
// 1-3: All other Players take 1 damage.
// 4-6: Deal 1 damage to an active monster
func curseOfTheTowerFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if _, err = en.checkDamageToPlayer(p.Character.id); err == nil {
		f = func(roll uint8) {
			if roll >= 1 && roll <= 3 {
				for _, o := range b.getOtherPlayers(p, true) {
					b.damagePlayerToPlayer(p, o, 1)
				}
			} else {
				monsters := b.monster.getActiveMonsters()
				showMonsterCards(monsters, 0)
				fmt.Println("Who to damage?")
				ans := readInput(0, len(monsters)-1)
				b.damagePlayerToMonster(p, monsters[ans], 1, 0)
			}
		}
	}
	return f, true, err
}

// Active Item
// Before a dice roll is rolled, say a number.
// If the next dice result is the number said, loot 3.
func crystalBallFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	fmt.Println("Guess a dice roll:")
	ans := uint8(readInput(1, 6))
	return func(roll uint8) { b.treasure.crystalBallGuess[p] = ans }, false, nil
}

// Event based passive item
// Each time you take damage, take an additional 1 damage.
// When you die, before paying penalties, give this card to another player
func daddyHauntFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if _, err = en.checkDamageToPlayer(p.Character.id); err == nil && !p.isDead() {
		f = func(roll uint8) { p.decreaseHP(1) }
	}
	return f, false, err
}

// Event based passive item
// When anyone rolls a 1, you may force a player to reroll it.
func dadsLostCoinFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(1); err == nil {
		fmt.Println("1) Reroll the roll of 1\n2) Do nothing")
		if uint8(readInput(1, 2)) == 1 {
			f = func(roll uint8) { b.rollDiceAndPush() }
		}
	}
	return f, false, err
}

// Event based passive item
// At the start of your turn, roll:
// 1-2: Gain 3 cents. 3-4: Loot 1. 5-6: Take 1 damage.
func darkBumFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkStartOfTurn(p); err == nil {
		f = func(roll uint8) {
			if roll == 1 || roll == 2 {
				p.gainCents(3)
			} else if roll == 3 || roll == 4 {
				p.loot(b.loot)
			} else {
				b.damagePlayerToPlayer(p, p, 1)
			}
		}
	}
	return f, true, err
}

// Event based passive item
// When anyone rolls a 3, you may look at that player's hand and steal a loot card.
func deadBirdFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(3); err == nil {
		fmt.Println("1) Steal a value from their hand\n2) Do nothing.")
		if readInput(1, 2) == 1 {
			f = func(roll uint8) {
				target := en.event.p
				hand := target.Hand
				l := len(hand)
				var i uint8
				if l > 0 {
					showLootCards(hand, target.Character.name, 0)
					if l > 1 {
						i = uint8(readInput(0, l-1))
					}
					p.Hand = append(p.Hand, target.popHandCard(i))
				}
			}
		}
	}
	return f, false, err
}

// Active Item
// Swap this *treasureCard with any non-eternal *treasureCard a player controls.
func decoyFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	var i uint8
	others := b.getOtherPlayers(p, false)
	l := len(others)
	if l > 1 {
		showPlayers(others, 0)
		i = uint8(readInput(0, l-1))
	}
	p2 := others[i]
	al := len(p2.ActiveItems)
	showTreasureCards(p2.ActiveItems, p2.Character.name, 0)
	showTreasureCards(p2.PassiveItems, p2.Character.name, al)
	ans := readInput(0, al+len(p2.PassiveItems)-1)
	var f cardEffect = func(roll uint8) {
		idx, _ := p.getItemIndex(tCard.getId(), false)
		dCard := p.popActiveItem(idx)
		if ans < al {
			p2.addCardToBoard(dCard)
			p.addCardToBoard(p2.popActiveItem(uint8(ans)))
		} else {
			p2.addCardToBoard(dCard)
			p.addCardToBoard(p2.popPassiveItem(uint8(ans - al)))
		}
	}
	return f, false, nil
}

// Constant Passive Item
// +1 HP
func dinnerFunc(p *player, b *Board, c card, leavingField bool) {
	modifyHealth(p, 1, leavingField)
}

// Active Item
// This becomes a copy of any non-eternal passive *treasureCard in play
// till the end of the turn.
func diplopiaFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	passives := b.getAllPassiveItems(false)
	l := len(passives)
	var i uint8
	if l == 0 {
		return nil, false, errors.New("no passives to copy")
	} else if l > 1 {
		showTreasureCards(passives, "board", 0)
		i = uint8(readInput(0, l-1))
	}
	toCopy := passives[i]
	var f cardEffect = func(roll uint8) {
		j, err := p.getItemIndex(tCard.getId(), false)
		if err == nil {
			p.popPassiveItem(j)
			p.addCardToBoard(toCopy)
		}
		p.activeEffects[diplopia] = struct{}{}
	}
	return f, false, nil
}

// Paid Item
// Give one of your other Items to another Player:
// Gain 8 cents.
func donationMachineFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	items := p.getAllItems(false)
	tcId := tCard.getId()
	l := len(items)
	if l == 1 && items[0].getId() == tcId {
		return nil, false, errors.New("no other item to give away")
	}
	showItems(items, 0)
	fmt.Println("Choose an item to give away.")
	ans := uint8(readInput(0, l-1))
	item := items[ans]
	if item.getId() == tcId {
		return nil, false, errors.New("cannot give away donation machine value")
	}
	others := b.getOtherPlayers(p, false)
	var i uint8
	l = len(others)
	if l > 1 {
		showPlayers(others, 0)
		i = uint8(readInput(0, l-1))
	}
	others[i].stealItem(tcId, false, p)
	return func(roll uint8) { p.gainCents(8) }, false, nil
}

// Constant with conditions passive item
// All damage done to you is reduced to 1.
//
// This card's effect will resolve when the player who owns the card
// takes damage.
func dryBabyFunc(p *player) bool {
	var b bool
	if _, err := p.getItemIndex(dryBaby, true); err == nil {
		p.decreaseHP(1)
		b = true
	}
	return b
}

// Event based passive item
// If you have 0 cents at the end of your turn, gain 6 cents
func edensBlessingFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkEndOfTurn(p); err == nil && p.Pennies == 0 {
		err, f = nil, func(roll uint8) { p.gainCents(6) }
	}
	return f, false, err
}

// Constant with conditions passive item
// If you have 0 loot cards in your hand, you gain +1 Attack
// If you have 0 cents, you gain +1 to all attack rolls
//
// The first effect is resolved at the battle step.
// The second effect is resolved at the rolling dice step
func emptyVesselChecker(p *player, raiseAttack bool) bool {
	var b bool
	if _, err := p.getItemIndex(emptyVessel, true); err == nil {
		if raiseAttack && len(p.Hand) == 0 { // Raise Character's attack
			b = true
		} else if !raiseAttack && p.Pennies == 0 { // Add 1 to the dice roll
			b = true
		}
	}
	return b
}

// Event based passive item
// When anyone rolls a 5, gain 3 cents.
func eyeOfGreedFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(5); err == nil {
		f = func(roll uint8) { p.gainCents(5) }
	}
	return f, false, err
}

// Event based passive item
// Each time you take damage, loot 1.
func fannyPackFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if _, err := en.checkDamageToPlayer(p.Character.id); err == nil {
		f = func(roll uint8) { p.loot(b.loot) }
	}
	return f, false, err
}

// Event based passive item
// When anyone rolls a 2, you may steal an item from that player.
// If you do, give that player one of your items.
func fingerFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(2); err == nil {
		target := en.event.p
		items := target.getAllItems(false)
		showItems(items, 0)
		fmt.Println("1) Swap an item with the player who rolled the dice.\n2) Do nothing.")
		if readInput(1, 2) == 1 {
			var i uint8
			l := len(items)
			if l > 1 {
				fmt.Println("Choose which item to steal.")
				i = uint8(readInput(0, len(items)-1))
			}
			item := items[i]
			f = func(roll uint8) {
				id, isPassive := item.getId(), item.isPassive()
				if i, err := target.getItemIndex(id, isPassive); err == nil {
					p.addCardToBoard(target.popItemByIndex(i, isPassive))
					items := p.getAllItems(false)
					showItems(items, 0)
					fmt.Println("Which item to give up?")
					toGive := items[uint8(readInput(0, len(items)-1))]
					j, _ := p.getItemIndex(toGive.getId(), toGive.isPassive())
					target.addCardToBoard(p.popItemByIndex(j, toGive.isPassive()))

				}
			}
		}
	}
	return f, false, err
}

// Active Item
// 1) Put all monsters not being attacked on the bottom of the monster deck.
// 2) Put all shop items on the bottom of the treasure deck.
func flushFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	var f cardEffect = func(roll uint8) {
		for _, zone := range b.monster.zones {
			monster := zone.peek()
			if !monster.inBattle {
				mon := zone.pop()
				b.monster.placeInDeck(mon, false)
			}
		}
	}
	buyItemEvents := b.eventStack.getIntentionToPurchaseEvents()
	if len(buyItemEvents) == 0 {
		fmt.Println("1) Put all active monsters not being attacked at the bottom of the monster deck.\n" +
			"2) Put all shop items on the bottom of the Treasure Deck.")
		ans := readInput(1, 2)
		if ans == 2 {
			f = func(roll uint8) {
				for i, c := range b.treasure.zones {
					b.treasure.placeInDeck(c, false)
					b.treasure.zones[i] = b.treasure.draw()
				}
			}
		}
	}
	return f, false, nil
}

// Starting Item Blue Baby
// Active Item
// 1) Steal 1 cent from a player. 2) Look at the top card of any deck.
// 3) Discard a Loot Card, then draw a Loot Card.
// When you take damage, recharge this.
func foreverAloneFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	var f cardEffect
	fmt.Println("Choose one:\n" +
		"1) Steal 1 cent from a Player.\n" +
		"2) Look at the top value of any deck.\n" +
		"3) Discard a Loot Card, then draw a Loot Card.")
	ans := readInput(1, 3)
	switch ans {
	case 1:
		var i int
		players := b.getOtherPlayers(p, false)
		if len(players) > 1 {
			showPlayers(players, 0)
			i = readInput(0, len(players))
		}
		if players[i].Pennies == 0 {
			return nil, false, errors.New("target player is dirt poor")
		}
		f = func(roll uint8) {
			players[i].loseCents(1)
			p.gainCents(1)
		}
	case 2:
		f = func(roll uint8) {
			fmt.Println("1) Loot Deck. 2) Monster Deck. 3) Treasure Deck.")
			ans := readInput(1, 3)
			var c card
			switch ans {
			case 1:
				c = b.loot.draw()
			case 2:
				c = b.monster.draw()
			case 3:
				c = b.treasure.draw()
			}
			if c != nil {
				fmt.Println(c.showCard(0))
				b.placeInDeck(c, true)
			}
		}
	case 3:
		f = func(roll uint8) {
			showLootCards(p.Hand, p.Character.name, 0)
			fmt.Println("Discard one.")
			ans := readInput(0, len(p.Hand)-1)
			p.popHandCard(uint8(ans))
			p.loot(b.loot)
		}
	}
	return f, false, nil
}

// Active Item
// Destroy another Item in play, then roll:
// 1-5: Destroy this and loot 2.
// 6: Recharge this.
func glassCannonFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	c := tCard.(*treasureCard)
	return func(roll uint8) {
		items, owner := b.getAllItems(false, nil)
		showItems(items, 0)
		ans := uint8(readInput(0, len(items)-1))
		id, isPassive := items[ans].getId(), items[ans].isPassive()
		i, _ := owner[id].getItemIndex(id, isPassive)
		b.discard(owner[id].popItemByIndex(i, isPassive))
		idx, e := p.getItemIndex(c.id, false)
		if e == nil {
			if roll != 6 {
				b.discard(p.popActiveItem(idx))
				p.loot(b.loot)
				p.loot(b.loot)
			} else {
				c.recharge()
			}
		}
	}, true, nil
}

// Event based passive item
// At the end of your turn, you may discard any number of Loot Cards,
// then loot equal to the number of cards discarded this way.
func goatHeadFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkEndOfTurn(p); err == nil {
		f = func(roll uint8) {
			var numDiscarded uint8
			for len(p.Hand) > 0 {
				l := len(p.Hand)
				showLootCards(p.Hand, "self", 0)
				fmt.Println(fmt.Sprintf("%d) Stop discarding.", l))
				ans := uint8(readInput(0, l))
				if ans < uint8(l) {
					b.discard(p.popHandCard(ans))
					numDiscarded += 1
				} else {
					break
				}
			}
			var i uint8
			for i = 0; i < numDiscarded; i++ {
				p.loot(b.loot)
			}
		}
	}
	return f, false, err
}

// Active Item
// Change the result of a dice roll to a 1 or a 6
func godheadFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	rollEvents := b.eventStack.getDiceRollEvents()
	l := len(rollEvents)
	var ans uint8
	if l == 0 {
		return nil, false, errors.New("no dice roll events on stack")
	} else if l > 1 {
		showEvents(rollEvents)
		ans = uint8(readInput(0, l-1))
	}
	fmt.Println("1) Change the roll to 1.\n2) Change the roll to 6.")
	var n uint8 = 1
	if readInput(1, 2) == 2 {
		n = 6
	}
	var f cardEffect = func(roll uint8) {
		rollEvents[ans].event = event{p: rollEvents[ans].event.p, e: diceRollEvent{n: n}}
	}
	return f, false, nil
}

// Paid Item
// Pay 5 cents:
// Deal 1 damage to a monster or player.
func goldenRazorBladeFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	if p.Pennies < 5 {
		return nil, false, errors.New("not enough cents to pay cost")
	}
	p.loseCents(5)
	monsters, players := b.monster.getActiveMonsters(), b.getPlayers(true)
	l := len(monsters)
	showMonsterCards(monsters, 0)
	showPlayers(players, l)
	ans := readInput(0, l-1)
	var f cardEffect
	if ans < l {
		f = func(roll uint8) { b.damagePlayerToMonster(p, monsters[ans], 1, 0) }
	} else {
		f = func(roll uint8) { b.damagePlayerToPlayer(p, players[ans-l], 1) }
	}
	return f, false, nil
}

// Event based passive
// Each time you die, gain 8 cents.
func greedsGulletFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDeath(p); err == nil {
		f = func(roll uint8) { p.gainCents(8) }
	}
	return f, false, err
}

// Event based preventative passive
// Each time you die, roll:
// 1-3: Prevent deathPenalty. If it was your turn, end it.
// 4-6: You die :(
func guppysCollarFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDeath(p); err == nil {
		f = func(roll uint8) {
			if roll >= 1 || roll <= 3 {
				en.event.e = fizzledEvent{}
			}
		}
	}
	return f, true, err
}

// Active Item
// Steal a loot card from a player.
// That player decides which card is stolen.
func guppysHeadFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	others := b.getOtherPlayers(p, false)
	l := len(others)
	var ans uint8
	if l > 1 {
		showPlayers(others, 0)
		ans = uint8(readInput(0, l-1))
	}
	p2 := others[ans]
	var f cardEffect = func(roll uint8) {
		showLootCards(p2.Hand, p2.Character.name, 0)
		fmt.Println("Pick which value to give away.")
		ans := uint8(readInput(0, len(p2.Hand)))
		c := p2.popHandCard(ans)
		p.Hand = append(p.Hand, c)
	}
	return f, false, nil
}

// Active Item
// Take 1 damage. Prevent up to two damage to a player.
func guppysPawFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	return func(roll uint8) {
		if checkActiveEffects(p.activeEffects, guppysPaw, true) {
			dEvents := b.eventStack.getDamageOfCharacterEvents()
			l := len(dEvents)
			var ans uint8
			if l != 0 {
				if l > 1 {
					showEvents(dEvents)
					ans = uint8(readInput(0, l-1))
				}
				fmt.Println("Prevent how much damage?\n1) 1.\n2) 2.")
				n := uint8(readInput(1, 2))
				b.eventStack.preventDamage(n, dEvents[ans])
			}
		}
	}, false, nil
}

// Active Item
// Prevent 1 damage to you.
// If any damage was prevented, deal 1 damage to another player
func hostHatFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	dEvents := b.eventStack.getDamageOfCharacterEvents()
	if len(dEvents) == 0 {
		return nil, false, errors.New("no damage of character events")
	}
	valid := make([]*eventNode, 0, len(dEvents))
	for _, d := range dEvents {
		if d.event.p.Character.id == p.Character.id {
			valid = append(valid, d)
		}
	}
	l := len(valid)
	var ans uint8
	if l == 0 {
		return nil, false, errors.New("no damage events targeting self")
	} else if l > 1 {
		showEvents(valid)
		ans = uint8(readInput(0, l-1))
	}
	damageNode := valid[ans]
	var f cardEffect = func(roll uint8) {
		err := b.eventStack.preventDamage(1, damageNode)
		if err == nil {
			var ans uint8
			players := b.getOtherPlayers(p, true)
			l := len(players)
			if l > 0 {
				if l > 1 {
					showPlayers(players, 0)
					ans = uint8(readInput(0, l-1))
				}
				b.eventStack.push(event{p: players[ans], e: damageEvent{n: 1}})
			}
		}

	}
	return f, false, nil
}

// Starting Item Lilith
// Active Item
// Look at a player's hand. You may switch a card from your hand with one of theirs.
func incubusFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	fmt.Println("Choose One:\n" +
		"1) Look at a Player's Hand, you may switch a value from your hand with one of theirs.\n" +
		"2) Loot 1, then place a value from your hand on top of the loot deck.")
	ans := readInput(1, 2)
	var f cardEffect
	switch ans {
	case 1:
		players := b.getOtherPlayers(p, false)
		l := len(players)
		var playerIdx uint8
		if l > 1 {
			showPlayers(players, 0)
			playerIdx = uint8(readInput(0, l-1))
		}
		p2 := players[playerIdx]
		f = b.incubus(p, p2)
	case 2:
		f = p.incubus(b.loot)
	}
	return f, false, nil
}

// Hybrid passive item
// +1 Attack
// Each time you roll a 6 while attacking, deal 1 damage to all other players
func ipecacFuncConstant(p *player, b *Board, tCard card, isLeaving bool) {
	modifyAttack(p, 1, isLeaving)
}

func ipecacFuncEvent(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	_, isAttack := b.eventStack.peek().event.e.(declareAttackEvent)
	if err = en.checkDiceRoll(6); err == nil && isAttack {
		f = func(roll uint8) {
			for _, o := range b.getOtherPlayers(p, true) {
				b.damagePlayerToPlayer(p, o, 1)
			}
		}
	}
	return f, false, err
}

// Active Item
// Steal 3 cents from another player
func jawboneFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	others := b.getOtherPlayers(p, false)
	l := len(others)
	var ans uint8
	if l > 1 {
		showPlayers(others, 0)
		ans = uint8(readInput(0, l-1))
	}
	p2 := others[ans]
	if p2.Pennies == 0 {
		return nil, false, errors.New("no pennies to steal")
	}
	var f cardEffect = func(roll uint8) {
		var n int8 = 3
		if p2.Pennies < 3 {
			n = p2.Pennies
		}
		p2.loseCents(n)
		p.gainCents(n)
	}
	return f, false, nil
}

// Active Item
// Add up to two to any non-attack roll.
func luckyFootFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	rolls := b.eventStack.getNonAttackDiceRollEvents()
	l := len(rolls)
	if l == 0 {
		return nil, false, errors.New("no non-attack dice roll events")
	}
	var i uint8
	if l > 1 {
		showEvents(rolls)
		i = uint8(readInput(0, l-1))
	}
	fmt.Println("1) Add 1.\n2) Add 2.")
	n := int8(readInput(1, 2))
	return func(roll uint8) { b.eventStack.addToDiceRoll(n, rolls[i]) }, false, nil
}

// Constant passive item
// +1 to all your attack rolls.
func meatFunc(p *player, b *Board, tCard card, isLeaving bool) {
	if !isLeaving {
		p.Character.attackDiceModifier += 1
	} else {
		p.Character.attackDiceModifier -= 1
	}
}

// Active Item
// Subtract up to 2 from any dice roll.
func miniMushFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	rolls := b.eventStack.getDiceRollEvents()
	l := len(rolls)
	if l == 0 {
		return nil, false, errors.New("no dice roll events")
	}
	var ans uint8
	if l > 1 {
		showEvents(rolls)
		ans = uint8(readInput(0, l-1))
	}
	fmt.Println("1) Subtract 1.\n2) Subtract 2.")
	n := int8(readInput(1, 2)) * -1
	return func(roll uint8) { b.eventStack.addToDiceRoll(n, rolls[ans]) }, false, nil
}

// Active Item
// This becomes a copy of any non-eternal Item in play.
// This change is permanent.
func modelingClayFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	items, owners := b.getAllItems(false, nil)
	showItems(items, 0)
	fmt.Println("Which value to copy?")
	ans := uint8(readInput(0, len(items)-1))
	id, isPassive := items[ans].getId(), items[ans].isPassive()
	var owner *player = owners[id]
	return func(roll uint8) {
		clayIdx, err := p.getItemIndex(tCard.getId(), false)
		if err == nil {
			var i uint8
			i, err := owner.getItemIndex(id, isPassive)
			if err == nil {
				clayCard := p.popItemByIndex(clayIdx, false)
				if !isPassive {
					c := owner.ActiveItems[i]
					c.id = clayCard.getId()
					p.addCardToBoard(&c)
				} else {
					pc := owner.PassiveItems[i]
					switch pc.(type) {
					case *treasureCard:
						card := pc.(*treasureCard)
						newC := *card
						newC.id = clayCard.getId()
						p.addCardToBoard(newC)
					case lootCard:
						card := pc.(lootCard)
						card.id = clayCard.getId()
						p.addCardToBoard(card)
					default:
						panic("need treasure value pointer or loot value")
					}
				}
			}
		}
	}, false, nil
}

// Active Item
// Reduce the damage dealt to any player or monster to 1.
func momsBraFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	dNodes := b.eventStack.getDamageEventsGT1()
	l := len(dNodes)
	if l == 0 {
		return nil, false, errors.New("no damage events")
	}
	var ans uint8
	if l > 1 {
		showEvents(dNodes)
		ans = uint8(readInput(0, l-1))
	}
	d := dNodes[ans]
	return func(roll uint8) { d.event = event{p: d.event.p, e: damageEvent{n: 1}} }, false, nil
}

// Event based passive
// When anyone rolls a 4, you may lot 1 and then discard a card
func momsBoxFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(4); err == nil {
		fmt.Println("1) Loot 1 then discard 1.\n2) Do Nothing")
		if readInput(1, 2) == 1 {
			f = func(roll uint8) {
				p.loot(b.loot)
				showLootCards(p.Hand, "self", 0)
				fmt.Println("Discard which value?")
				b.discard(p.popHandCard(uint8(readInput(0, len(p.Hand)-1))))
			}
		}
	}
	return f, false, err
}

// Event based passive (mom's coin purse AND mom's purse)
// Loot +1 at the start of your turn.
func momsPursesFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkStartOfTurn(p); err == nil {
		f = func(roll uint8) { p.loot(b.loot) }
	}
	return f, false, err
}

// Event based passive
// When anyone rolls a 6, you may deal 1 damage to them
func momsRazorFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(6); err == nil {
		fmt.Println("1) Deal one damage\n2) Do Nothing")
		if readInput(1, 2) == 1 {
			f = func(roll uint8) { b.damagePlayerToPlayer(p, en.event.p, 1) }
		}
	}
	return f, false, err
}

// Active Item
// This enters play deactivated.
// Destroy this: steal a soul card from a player.
func momsShovelFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	idx, err := p.getItemIndex(tCard.getId(), false)
	if err != nil {
		return nil, false, err
	}
	b.discard(p.popActiveItem(idx))
	var f cardEffect = func(roll uint8) {
		others := b.getOtherPlayers(p, false)
		var i uint8
		l := len(others)
		if l > 1 {
			showPlayers(others, 0)
			i = uint8(readInput(0, l-1))
		}
		p2 := others[i]
		l = len(p2.Souls)
		if l > 0 {
			var j uint8
			if l > 1 {
				showSouls(p2.Souls, p2.Character.name, 0)
				j = uint8(readInput(0, l-1))
			}
			p.addSoulToBoard(p2.popSoul(j))
		}
	}
	return f, false, nil
}

// Active Item
// Force the active player to attack. You choose what they attack.
func monsterManualFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	ap := &b.players[b.api]
	if ap.inBattle {
		return nil, false, errors.New("active player already in battle")
	}
	monsters := b.monster.getActiveMonsters()
	showMonsterCards(monsters, 0)
	ans := readInput(0, len(monsters)-1)
	m := monsters[ans]
	return func(roll uint8) { m.inBattle, ap.inBattle = true, true }, false, nil
}

// Event based passive item
// At the start of your turn, choose a player at random.
// That player destroys an item they own of their choosing.
func monstrosToothFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkStartOfTurn(p); err == nil {
		randIdx := uint8(rand.Intn(len(b.players)))
		target := &b.players[randIdx]
		f = func(roll uint8) {
			items := target.getAllItems(false)
			l := len(items)
			var i uint8
			if l > 0 {
				if l > 1 {
					showItems(items, 0)
					i = uint8(readInput(0, l-1))
				}
			}
			item := items[i]
			i, _ = p.getItemIndex(item.getId(), item.isPassive())
			b.discard(p.popItemByIndex(i, item.isPassive()))
		}
	}
	return f, false, err
}

// Active Item
// Deal 1 damage to a monster.
func mrBoomFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	monsters := b.monster.getActiveMonsters()
	showMonsterCards(monsters, 0)
	ans := readInput(0, len(monsters)-1)
	m := monsters[ans]
	return func(roll uint8) { b.damagePlayerToMonster(p, m, 1, 0) }, false, nil
}

// Active Item
// Roll:
// 1 - 2: Loot 1. 3-4: Gain 4 cents. 5-6: Nothing
func mysterySackFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	return func(roll uint8) {
		if roll == 1 || roll == 2 {
			p.loot(b.loot)
		} else if roll == 3 || roll == 4 {
			p.gainCents(4)
		}
	}, true, nil
}

// Active Item
// Cancel the effect of any Active Item
func noFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	activeItemEvents := b.eventStack.getActivateItemEvents()
	l := len(activeItemEvents)
	if l == 0 {
		return nil, false, errors.New("no active items tapped")
	}
	var i uint8
	if l > 1 {
		showEvents(activeItemEvents)
		i = uint8(readInput(0, l-1))
	}
	node := activeItemEvents[i]
	return func(roll uint8) { _ = b.eventStack.fizzle(node) }, false, nil
}

// Active Item
// Destroy this. Then roll:
// 1: Gain 1 cent. 2: Gain 6 cents. 3: Kill a monster.
// 4: Loot 3. 5: Gain 9 cents. 6: This becomes a Soul. Gain it.
func pandorasBoxFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	idx, _ := p.getItemIndex(tCard.getId(), false)
	b.discard(p.popActiveItem(idx))
	var f cardEffect = func(roll uint8) {
		switch roll {
		case 1:
			p.gainCents(1)
		case 2:
			p.gainCents(6)
		case 3:
			monsters := b.monster.getActiveMonsters()
			showMonsterCards(monsters, 0)
			fmt.Println("Kill which monster?")
			ans := uint8(readInput(0, len(monsters)-1))
			b.killMonster(p, monsters[ans].id)
		case 4:
			for i := 0; i < 3; i++ {
				p.loot(b.loot)
			}
		case 5:
			p.gainCents(9)
		case 6:
			if card, err := b.treasure.discardPile.popById(tCard.getId()); err == nil {
				p.addSoulToBoard(card)
			}
		}
	}
	return f, true, nil
}

// Paid Item
// Pay 10 Cents:
// Steal an Item from any Player.
func payToPlayFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	if p.Pennies < 10 {
		return nil, false, errors.New("not enough pennies to pay cost")
	}
	p.loseCents(10)
	items, owners := b.getAllItems(false, p)
	showItems(items, 0)
	fmt.Println("Choose an item to steal.")
	ans := uint8(readInput(0, len(items)-1))
	id, isPassive := items[ans].getId(), items[ans].isPassive()
	owner := owners[id]
	return func(roll uint8) { p.stealItem(id, isPassive, owner) }, false, nil
}

// Active Item
// Copy the activated effect of any non-eternal item in play.
func placeboFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	active := make([]*treasureCard, 0)
	others := b.getOtherPlayers(p, false)
	for _, o := range others {
		active = append(active, o.getActiveItems(false)...)
	}
	active = append(active, p.getActiveItems(false)...)
	l := len(active)
	if l == 1 && active[0].id == placebo {
		return nil, false, errors.New("no new effects to copy")
	}
	showTreasureCards(active, "board", 0)
	ans := readInput(0, len(active)-1)
	var f cardEffect = func(roll uint8) {
		tempTc := treasureCard{baseCard: active[ans].baseCard, paid: active[ans].paid, active: true, f: active[ans].f}
		tempTc.id = placebo
		tempTc.activate(p, b)

	}
	return f, false, nil
}

// Constant passive item
// You may play an additional loot card on your turn.
// Gain +1 Attack for the first attack roll of your turn
func polydactylyFunc(p *player, b *Board, tCard card, isLeaving bool) {
	if !isLeaving {
		p.numLootPlayed += 1
		p.activeEffects[polydactyly] = struct{}{}
	} else {
		p.numLootPlayed -= 1
		delete(p.activeEffects, polydactyly)
	}
}

// Paid item
// Pay 3 cents:
// Roll:
// 1-2: Loot 1. 3-4: Gain 4 cents. 5-6: Nothing.
func portableSlotMachineFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	if p.Pennies < 3 {
		return nil, false, errors.New("not enough cents to pay cost")
	}
	return func(roll uint8) {
		if roll == 1 || roll == 2 {
			p.loot(b.loot)
		} else if roll == 3 || roll == 4 {
			p.gainCents(4)
		}
	}, true, nil
}

// Active Item
// Put the top cards of all decks into their discard piles.
func potatoPeelerFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	return func(roll uint8) { b.discard(b.loot.draw()); b.discard(b.monster.draw()); b.discard(b.treasure.draw()) }, false, nil
}

// Active Item
// Deal 1 damage to another player
func razorBladeFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	players := b.getOtherPlayers(p, true)
	l := len(players)
	var i uint8
	if l > 1 {
		showPlayers(players, 0)
		i = uint8(readInput(0, l-1))
	}
	target := players[i]
	var f cardEffect = func(roll uint8) {
		b.damagePlayerToPlayer(p, target, 1)
	}
	return f, false, nil
}

// Active Item
// Each player votes on an *treasureCard in play.
// Destroy the *treasureCard with the most votes.
// If there is a tie, cancel this effect.
func remoteDetonatorFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	return func(roll uint8) {
		itemVotes, cardTypes, owners := b.remoteDetonatorVoteHelper()
		type kv struct {
			cardId   uint16
			numVotes uint8
		}
		var sorted []kv
		for k, v := range itemVotes {
			sorted = append(sorted, kv{k, v})
		}
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].numVotes > sorted[j].numVotes
		})
		if sorted[0].numVotes > sorted[1].numVotes {
			id, isPassive := sorted[0].cardId, cardTypes[sorted[0].cardId]
			owner := owners[id]
			j, err := owner.getItemIndex(id, isPassive)
			if err == nil {
				owner.popItemByIndex(j, isPassive)
			}
		}
	}, false, nil
}

// Event based passive item
// At the start of your turn, you may discard any shop items and replace them
// with the top cards of the treasure deck
func restockFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkStartOfTurn(p); err == nil {
		f = func(roll uint8) {
			fmt.Println("1) Change the shop\n2) Don't")
			if readInput(1, 2) == 1 {
				numZones := len(b.treasure.zones)
				for len(b.treasure.zones) > 0 {
					l := len(b.treasure.zones)
					showTreasureCards(b.treasure.zones, "shop", 0)
					fmt.Println(fmt.Scanf("%d) Stop discarding."))
					i := readInput(0, l)
					if i < l {
						b.discard(b.treasure.zones[i])
						copy(b.treasure.zones[i:], b.treasure.zones[i+1:])
						b.treasure.zones[l-1] = treasureCard{}
						b.treasure.zones = b.treasure.zones[:l-1]
					} else {
						break
					}
				}
				for len(b.treasure.zones) < numZones {
					b.treasure.zones = append(b.treasure.zones, b.treasure.draw())
				}
			}
		}
	}
	return f, false, err
}

// Active Item
// Look at the top card of any deck.
// You may put that card on the bottom of that deck.
func sackHeadFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	fmt.Println("1) Loot Deck\n2) Monster Deck\n3) Treasure Deck")
	ans := readInput(1, 3)
	var f cardEffect = func(roll uint8) {
		var c card
		switch ans {
		case 1:
			c = b.loot.draw()
		case 2:
			c = b.monster.draw()
		case 3:
			c = b.treasure.draw()
		default:
			panic("input error")
		}
		c.showCard(0)
		fmt.Println("1) Place on Bottom. 2) Do nothing.")
		ans = readInput(1, 2)
		if ans == 1 {
			b.placeInDeck(c, false)
		} else {
			b.placeInDeck(c, true)
		}
	}
	return f, false, nil
}

// Active Item
// Gain 1 cents.
// When any player rolls a 1 you may recharge this.
func sackOfPenniesFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	return func(roll uint8) { p.gainCents(1) }, false, nil
}

// Event based passive item
// Each time you roll a 1, you may turn it into a 6
func sacredHeartFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(1); err == nil && en.event.p.Character.id == p.Character.id {
		fmt.Println("1) Change 1 to 6\n2) Do not.")
		if readInput(1, 2) == 1 {
			f = func(roll uint8) { en.event.e = diceRollEvent{n: 6} }
		}
	}
	return f, false, err
}

// Event based passive with conditions
// Each time another player dies, you choose what items they destroy
// Each time another player dies, you gain any loot or cents they lose.
//
// Helper to character deathPenalty event
// Returns true if shadow triggers. False if does not.
func shadowFunc(p *player, b *Board) bool {
	var activated bool
	if _, p2 := b.getItemIndex(shadow, true); p2 != nil && p2 != p {
		activated = true
		pItems := p.getAllItems(false)
		l := len(pItems)
		if l > 0 {
			fmt.Println("Choose which item ", p.Character.name, " destroys")
			i := uint8(readInput(0, l-1))
			card, _ := p.popItem(pItems[i])
			b.discard(card)
		}
		p.loseCents(1)
		p2.gainCents(1)
		showLootCards(p.Hand, p.Character.name, 0)
		fmt.Println("Choose which value to discard and add to your hand.")
		p2.Hand = append(p2.Hand, p.popHandCard(uint8(readInput(0, len(p.Hand)-1))))
	}
	return activated
}

// Event passive item
// Each time you activate an item, gain 1 cent
func shinyRockFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkActivateItemEvent(); err == nil && p.Character.id == en.event.p.Character.id {
		f = func(roll uint8) { p.gainCents(1) }
	}
	return f, false, err
}

// Starting Item (Cain)
// Active Item
// Look at the top 3 cards of a deck, put them back in any order.
func sleightOfHandFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	var f cardEffect = func(roll uint8) {
		cards := make([]card, 3)
		fmt.Println("Choose a deck:\n1) Loot Deck.\n2) Monster Deck.\n3) Treasure Deck.")
		deckType := readInput(1, 3)
		switch deckType {
		case 1:
			for i := 0; i < 3; i++ {
				cards[i] = b.loot.draw()
			}
			showLootCards(cards, "deck", 0)
		case 2:
			for i := 0; i < 3; i++ {
				cards[i] = b.monster.draw()
			}
			showMonsterCards(cards, 0)
		case 3:
			for i := 0; i < 3; i++ {
				cards[i] = b.treasure.draw()
			}
			showTreasureCards(cards, "deck", 0)
		}
		for len(cards) > 1 {
			fmt.Println("Pick a value to go back on the top of the deck.")
			switch deckType {
			case 1:
				showLootCards(cards, "deck", 0)
			case 2:
				showMonsterCards(cards, 0)
			case 3:
				showTreasureCards(cards, "deck", 0)
			}
			ans := readInput(0, len(cards))
			b.placeInDeck(cards[ans], true)
			cards = append(cards[:ans], cards[ans+1:]...)
		}
		b.placeInDeck(cards[0], true)
	}
	return f, false, nil
}

// Active Item
// Look at the top card of any deck.
// You may discard it or place it back on top.
func smartFlyFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	fmt.Println("1) Loot Deck\n2) Monster Deck\n3) Treasure Deck")
	ans := uint8(readInput(1, 3))
	return func(roll uint8) {
		var c card
		switch ans {
		case 1:
			c = b.loot.draw()
		case 2:
			c = b.monster.draw()
		case 3:
			c = b.treasure.draw()
		default:
			panic("input error")
		}
		c.showCard(0)
		fmt.Println("1) Discard it. 2) Place back on top.")
		ans = uint8(readInput(1, 2))
		if ans == 1 {
			b.discard(c)
		} else {
			b.placeInDeck(c, true)
		}
	}, false, nil
}

// Paid Item
// Discard a Loot Card:
// Gain 3 cents
func smelterFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	showLootCards(p.Hand, "self", 0)
	fmt.Println("Discard which value?")
	b.loot.discard(p.popHandCard(uint8(readInput(0, len(p.Hand)-1))))
	return func(roll uint8) { p.gainCents(3) }, false, nil
}

// Event based passive item
// When anyone rolls a 5, discard an active monster that isn't being attacked and replace
// it with the top card of the deck.
func spiderModFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(5); err == nil {
		monsters := make([]*monsterCard, 0, len(b.monster.zones))
		for _, z := range b.monster.zones {
			if m := z.peek(); err == nil && m.inBattle == false {
				monsters = append(monsters, m)
			}
		}
		l := len(monsters)
		var i uint8
		if l > 1 {
			showMonsterCards(monsters, 0)
			fmt.Println("Choose which value to get rid of and replace")
			i = uint8(readInput(0, l-1))
		}
		f = func(roll uint8) {
			card := b.monster.zones[i].pop()
			b.discard(card)
			b.monster.fillMonsterZone(&b.players[b.api], b, i)
		}
	}
	return f, false, err
}

// Active Item
// Add 1 to any dice roll.
func spoonBenderFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	diceRolls := b.eventStack.getDiceRollEvents()
	l := len(diceRolls)
	if l == 0 {
		return nil, false, errors.New("no dice roll events")
	}
	var ans uint8
	if l > 1 {
		showEvents(diceRolls)
		ans = uint8(readInput(0, l-1))
	}
	node := diceRolls[ans]
	return func(roll uint8) { b.eventStack.addToDiceRoll(1, node) }, false, nil
}

// Event based passive item
// If you have 8 or more loot cards in your hand at the end of your turn, loot 2
func starterDeckFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkEndOfTurn(p); err == nil && len(p.Hand) >= 8 {
		f = func(roll uint8) {
			for i := 0; i < 2; i++ {
				p.loot(b.loot)
			}
		}
	}
	return f, false, err
}

// Constant with conditions passive item
// You may buy shop items for 5 cents.
func steamySaleFunc(p *player) int8 {
	var cost int8 = 10
	if _, err := p.getItemIndex(steamySale, true); err == nil {
		cost = 5
	}
	return cost
}

// Event based passive
// Each time you die, loot 3.
func suicideKingFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDeath(p); err == nil {
		f = func(roll uint8) {
			for i := 0; i < 3; i++ {
				p.loot(b.loot)
			}
		}
	}
	return f, false, err
}

// Constant passive item
// +1 to all your attack rolls
func synthoilFunc(p *player, b *Board, tCard card, isLeaving bool) {
	if !isLeaving {
		p.Character.attackDiceModifier += 1
	} else {
		p.Character.attackDiceModifier -= 1
	}
}

// Event based passive item
// When anyone rolls a 4, that player must choose a loot card in their hand and give it to you
func tarotClothFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(4); err == nil && en.event.p != p {
		f = func(roll uint8) {
			target := en.event.p
			l := len(target.Hand)
			if l > 0 {
				showLootCards(target.Hand, target.Character.name, 0)
				fmt.Println("Choose which value to give to", p.Character.name)
				ans := uint8(readInput(0, l-1))
				p.Hand = append(p.Hand, target.popHandCard(ans))
			}
		}
	}
	return f, false, err
}

// Active / Paid Item hybrid
// Put a counter on this.
// Remove three counters from this: kill a Player or Monster.
func techXFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	c := tCard.(*treasureCard)
	var f cardEffect = func(roll uint8) {
		c.counters += 1
	}
	var usedPaidEff bool
	if c.counters >= 3 {
		fmt.Println("1) Add a counter.\n2) Kill a Player or Monster.")
		ans := uint8(readInput(1, 2))
		if ans == 2 {
			usedPaidEff = true
			monsters, players := b.monster.getActiveMonsters(), b.getPlayers(true)
			l := len(monsters)
			showMonsterCards(monsters, 0)
			showPlayers(players, l)
			i := uint8(readInput(0, l-1))
			f = func(roll uint8) {
				if i < uint8(l) {
					b.killMonster(p, monsters[i].id)
				} else {
					b.killPlayer(players[i-uint8(l)])
				}
			}
		}
	}
	return f, usedPaidEff, nil
}

// Active Item
// Recharge another Item
func theBatteryFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	items := p.getTappedActiveItems()
	showTreasureCards(items, p.Character.name, 0)
	ans := readInput(0, len(items)-1)
	var f cardEffect = func(roll uint8) { p.rechargeActiveItemById(items[ans].id) }
	return f, false, nil
}

// Event based passive item
// At the end of your turn, look at the top four cards of the Treasure deck.
// Yu may put them back in any order.
func theBlueMapFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkEndOfTurn(p); err == nil {
		err, f = nil, func(roll uint8) {
			cards := make([]treasureCard, 4)
			for i := 0; i < 4; i++ {
				cards[i] = b.treasure.draw()
			}
			fmt.Println("Choose order to go back from bottom to top")
			for len(cards) > 1 {
				showTreasureCards(cards, "deck", 0)
				ans := readInput(0, len(cards))
				b.treasure.placeInDeck(cards[ans], true)
				cards = append(cards[:ans], cards[ans+1:]...)
			}
		}
	}
	return f, false, err
}

func theBoneFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	tc := tCard.(*treasureCard)
	msgs := [3]string{
		"1) Remove 1 counter: Add +1 to a dice roll.",
		"2) Remove 2 counters: Deal 1 damage to a Monster or Player.",
		"3) Remove 3 counters: This loses all abilities and becomes a Soul.",
	}
	var f cardEffect = func(roll uint8) {
		tc.counters += 1
	}
	var usePaidEff bool
	if tc.counters == 0 {
		return f, false, nil
	}
	var i, n int8 = 0, tc.counters
	if n > 3 {
		n = 3
	}
	fmt.Println("0) Put a counter on this.")
	for i = 0; i < n; i++ {
		fmt.Println(msgs[i])
	}
	ans := uint8(readInput(0, int(n)))
	if ans > 0 {
		usePaidEff = true
	}
	var err error
	switch ans {
	case 0:
		return f, false, nil
	case 1:
		tc.loseCounters(1)
		f, err = b.eventStack.theBoneFirstPaidHelper(tc)
	case 2:
		tc.loseCounters(2)
		f = b.theBoneSecondPaidHelper(p, b.getPlayers(true), b.monster.getActiveMonsters())
	case 3:
		tc.loseCounters(3)
		f = func(roll uint8) {
			tc.f = func(p *player, b *Board, tCard card) (cardEffect, bool, error) {
				return func(roll uint8) {}, false, errors.New("bone lost all abilities")
			}
			i, _ := p.getItemIndex(theBone, false)
			p.addSoulToBoard(p.popActiveItem(i))
		}
	}
	return f, usePaidEff, err
}

// Constant passive item
// If this item is destroyed, it becomes a soul for the player who owned it.
func theChestFunc(p *player, b *Board, tCard card, isLeaving bool) {
	if isLeaving {
		p.addSoulToBoard(tCard)
	}
}

// Event based passive
// At the end of your turn, look at the top 4 cards of the Loot deck.
// You may put them back in any order.
func theCompassFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkEndOfTurn(p); err == nil {
		err, f = nil, func(roll uint8) {
			cards := make([]lootCard, 4)
			for i := 0; i < 4; i++ {
				cards[i] = b.loot.draw()
			}
			fmt.Println("Choose order to go back from bottom to top")
			for len(cards) > 1 {
				showTreasureCards(cards, "deck", 0)
				ans := readInput(0, len(cards))
				b.placeInDeck(cards[ans], true)
				cards = append(cards[:ans], cards[ans+1:]...)
			}
		}
	}
	return f, false, err
}

// Starting Item (Eve)
// Active Item
// Put the top card of any discard pile on top of its deck.
func theCurseFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	var f cardEffect = func(roll uint8) {
		var i = 1
		choices := make(map[int]card)
		lDiscard := len(b.loot.discardPile)
		if lDiscard > 0 {
			lc := b.loot.discardPile[lDiscard-1]
			lc.showCard(i)
			choices[i] = lc
			i += 1
		}
		mDiscard := len(b.monster.discardPile)
		if mDiscard > 0 {
			mc := b.monster.discardPile[mDiscard-1]
			mc.showCard(i)
			choices[i] = mc
			i += 1
		}
		tDiscard := len(b.treasure.discardPile)
		if tDiscard > 0 {
			tc := b.treasure.discardPile[tDiscard-1]
			tc.showCard(i)
			choices[i] = tc
			i += 1
		}
		fmt.Println("Which value?")
		ans := readInput(1, 3)
		b.placeInDeck(choices[ans], true)
	}
	return f, false, nil
}

// Active Item
// Destroy this.
// Choose a Player, that player destroys all Items they control,
// Then gains treasure equal the the number of items destroyed.
func theD4Func(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	idx, _ := p.getItemIndex(tCard.getId(), false)
	b.discard(p.popActiveItem(idx))
	players := b.getPlayers(false)
	ans := uint8(readInput(0, len(players)-1))
	target := players[ans]
	return func(roll uint8) {
		var numDiscarded uint8
		for len(target.ActiveItems) > 0 {
			c := target.popActiveItem(0)
			b.discard(c)
			numDiscarded += 1
		}
		for len(target.PassiveItems) > 0 {
			c := target.popPassiveItem(0)
			b.discard(c)
			numDiscarded += 1
		}
		var i uint8
		for i = 0; i < numDiscarded; i++ {
			c := b.treasure.draw()
			target.addCardToBoard(&c)
		}
	}, false, nil
}

// Starting Item (Isaac)
// Active Item
// Force a Player to reroll any dice roll.
func theD6Func(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	var ans int
	var node *eventNode
	nodes := b.eventStack.getDiceRollEvents()
	l := len(nodes)
	if l > 0 {
		if l > 1 {
			showEvents(nodes)
			ans = readInput(0, l-1)
		}
		node = nodes[ans]
	}
	if node == nil {
		return nil, false, errors.New("no dice events to change")
	}
	var f cardEffect = func(roll uint8) {
		node.event = event{p: node.event.p, e: diceRollEvent{n: uint8(rand.Intn(6) + 1)}}
	}
	return f, false, nil
}

// Event based passive item
// When anyone rolls a 3, you may put the top card of the monster deck into an active slot
// that isn't being attacked.
func theD10Func(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(3); err == nil {
		fmt.Println("1) Overlay a monster with the top value of the monster deck?\n2)Do nothing.")
		if readInput(1, 2) == 1 {
			showMonsterCards(b.monster.getActiveMonsters(), 0)
			fmt.Println("Which zone to place in?")
			i := uint8(readInput(0, len(b.monster.zones)-1))
			f = func(roll uint8) {
				monsters := b.monster.getActiveMonsters()
				if !monsters[i].inBattle {
					b.addMonsterToZone(i)
				}
			}
		}
	}
	return f, false, err
}

// Active Item
// Destroy any *treasureCard in play and replace it with the top card of the treasure deck.
func theD20Func(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	players := b.getPlayers(false)
	showPlayers(players, 0)
	ans := readInput(0, len(players)-1)
	player := players[ans]
	al := len(player.ActiveItems)
	showTreasureCards(player.ActiveItems, player.Character.name, 0)
	showTreasureCards(player.PassiveItems, player.Character.name, al)
	i := readInput(0, al+len(player.PassiveItems)-1)
	var id uint16
	var isPassive bool
	if i < al {
		id = player.ActiveItems[i].id
	} else {
		id, isPassive = player.PassiveItems[i-al].getId(), true
	}
	return func(roll uint8) {
		i, err := player.getItemIndex(id, isPassive)
		if err == nil {
			b.discard(player.popItemByIndex(i, isPassive))
			item := b.treasure.draw()
			player.addCardToBoard(&item)
		}
	}, false, nil
}

// Active Item
// roll:
// 1: Loot 1. 2: Loot 2. 3: Gain 3 Cents.
// 4: Gain 4 cents. 5: Gain 1 hp till the end of the turn. 6: Gain 1 ap till the end of the turn.
func theD100Func(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	return func(roll uint8) {
		switch roll {
		case 1:
			p.loot(b.loot)
		case 2:
			p.loot(b.loot)
			p.loot(b.loot)
		case 3:
			p.gainCents(3)
		case 4:
			p.gainCents(4)
		case 5:
			p.increaseHP(1)
		case 6:
			p.increaseAP(1)
		}
	}, true, nil
}

// Hybrid passive item
// This item starts with 9 Counters on it.
// Each time you take damage, remove that many counters from this, and prevent that damage.
// This counts as a Guppy Item.
func theDeadCatFuncConstant(p *player, b *Board, tCard card, isLeaving bool) {
	if !isLeaving {
		c := tCard.(*treasureCard)
		c.counters = 9
	}
}

func theDeadCatChecker(deadCat *treasureCard, es *eventStack, damageNode *eventNode) (cardEffect, error) {
	var f cardEffect
	var err = errors.New("dead cat could not activate")
	if d, ok := damageNode.event.e.(damageEvent); ok {
		if deadCat.id == theDeadCat && deadCat.counters > 0 {
			err, f = nil, func(roll uint8) {
				var x int8 = int8(d.n)
				if deadCat.counters < x {
					x = deadCat.counters
				}
				_ = es.preventDamage(uint8(x), damageNode)
				deadCat.loseCounters(x)
			}
		} else {
			panic("need a pointer to the dead cat card!")
		}
	}
	return f, err
}

// Hybrid passive item
// When you take damage for the first time each turn, you may recharge an item.
func theHabitFuncConstant(p *player, b *Board, tCard card, isLeaving bool) {
	if !isLeaving {
		p.activeEffects[theHabit] = struct{}{}
	} else {
		delete(p.activeEffects, theHabit)
	}
}

func theHabitFuncEvent(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if _, err = en.checkDamageToPlayer(p.Character.id); err == nil && checkActiveEffects(p.activeEffects, theHabit, true) {
		fmt.Println("1) Recharge an item\n2) Do nothing.")
		if readInput(1, 2) == 1 {
			a := p.getTappedActiveItems()
			l := len(a)
			var i uint8
			if l > 0 {
				if l > 1 {
					showTreasureCards(a, "self", 0)
					i = uint8(readInput(0, l-1))
				}
				f = func(roll uint8) { p.rechargeActiveItemById(a[i].id) }
			}
		}
	}
	return f, false, err
}

// Event based passive item
// At the end of your turn, look at the top four cards of the Loot deck.
// Yu may put them back in any order.
func theMapFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkEndOfTurn(p); err == nil {
		err, f = nil, func(roll uint8) {
			cards := make([]lootCard, 4)
			for i := 0; i < 4; i++ {
				cards[i] = b.loot.draw()
			}
			fmt.Println("Choose order to go back from bottom to top")
			for len(cards) > 1 {
				showLootCards(cards, "deck", 0)
				ans := readInput(0, len(cards))
				b.loot.placeInDeck(cards[ans], true)
				cards = append(cards[:ans], cards[ans+1:]...)
			}
		}
	}
	return f, false, err
}

// Event based passive item with conditions
// Each time a monster is killed gain 3 cents.
//
// Implemented in the kill monster helper
func theMidasTouchFunc(p *player, b *Board, tCard card, isLeaving bool) {
	if !isLeaving {
		b.monster.theMidasTouch[p] = struct{}{}
	} else {
		delete(b.monster.theMidasTouch, p)
	}
}

// Event based passive item
// If you have 0 loot cards in your hand at the end of your turn, loot 2.
func thePolaroidFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkEndOfTurn(p); err == nil && len(p.Hand) == 0 {
		f = func(roll uint8) {
			for i := 0; i < 2; i++ {
				p.loot(b.loot)
			}
		}
	}
	return f, false, err
}

// Paid / Passive Hybrid
// Whenever you take damage put a counter on this.
// Remove 1 counter: Prevent 1 damage done to you.
func thePoopFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	tc := tCard.(*treasureCard)
	if tc.counters == 0 {
		return nil, false, errors.New("no counters to remove")
	}
	tc.loseCounters(1)
	nodes := b.eventStack.getDamageOfCharacterEvents()
	valid := make([]*eventNode, 0, len(nodes))
	for _, n := range nodes {
		if n.event.p.Character.id == p.Character.id {
			valid = append(valid, n)
		}
	}
	l := len(valid)
	if l == 0 {
		return nil, false, errors.New("no damage done to player")
	}
	var i uint8
	if l > 1 {
		showEvents(valid)
		i = uint8(readInput(0, l-1))
	}
	return func(roll uint8) { _ = b.eventStack.preventDamage(1, valid[i]) }, false, nil
}

// Event based passive item
// When anyone rolls a 1, loot 1
func theRelicFunc(p *player, b *Board, tCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(1); en == nil {
		f = func(roll uint8) { p.loot(b.loot) }
	}
	return f, false, err
}

// Hybrid passive item
// You may look at the top card of the treasure deck at any time during your turn
// You may purchase an additional item
func theresOptionsFunc(p *player, b *Board, tCard card, isLeaving bool) {
	if !isLeaving {
		p.numPurchases += 1
		p.baseNumPurchases += 1
	} else {
		p.numPurchases -= 1
		p.baseNumPurchases -= 1
	}
}

// Active Item
// Put any discarded monster card back on top of the monster deck.
func theShovelFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	return func(roll uint8) {
		if l := len(b.monster.discardPile); l > 0 {
			showDeck(b.monster.discardPile, true)
			fmt.Println("Put which card on top of the monster deck?")
			ans := l - readInput(0, l-1) - 1
			b.monster.placeInDeck(b.monster.popCardFromDiscardPile(uint8(ans)), true)
		}
	}, false, nil
}

// Constant passive item
// Other players can't play loot cards or activate items on your turn.
//
// Will just return if the item is presnt or not
func trinityShieldFunc(p *player) bool {
	if _, err := p.getItemIndex(trinityShield, true); err == nil {
		return true
	}
	return false
}

// Active Item
// Double the number of loot cards a player would draw, till the end of the turn.
func twoOfClubsFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	players := b.getPlayers(false)
	showPlayers(players, 0)
	ans := readInput(0, len(players)-1)
	player := players[ans]
	return func(roll uint8) { player.activeEffects[twoOfClubs] = struct{}{} }, false, nil
}

// Starting Item (Maggy)
// Active Item
// Prevent 1 damage dealt to any Player or Monster
func yumHeartFunc(p *player, b *Board, tCard card) (cardEffect, bool, error) {
	var ans uint8
	damage := b.eventStack.getDamageEvents()
	max := len(damage)
	if max == 0 {
		return nil, false, errors.New("no damage to prevent")
	} else if max > 1 {
		showEvents(damage)
		ans = uint8(readInput(0, len(damage)-1))
	}
	var f cardEffect = func(roll uint8) {
		_ = b.eventStack.preventDamage(1, damage[ans])
	}
	return f, false, nil
}
