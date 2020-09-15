/*
This file will host all the activated and resolved effects of every loot value
in the game. These cards are all activated from the player's hand.

There are three loot value types:
Basic: Effects that are used to accumulate resources, such as pennies, or aid in combat, such as stat buffs
Tarot: More powerful versions of the basic loot cards
Trinkets: They have no activation effect from the hand, they are added to the player's board as a passive item

Broadly speaking there are two types of passive items:
1) Event-based: Require some external event to occur prior to triggering its effect (dice roll, damage, start / end of turn)
	- These effects will trigger upon some successfully resolved event
	- Ex: board's roll variable being set, damage / deathPenalty events prior to inflicting damage / deathPenalty, start / end of turn
2) Constant: Provides a constant game changing effect until the value leaves play

Only event-based passive effects can be pushed to the stack and are not directly activated by
the player.

*/

package four_souls

import (
	"errors"
	"fmt"
	"math/rand"
)

// Basic loot
// Gain 1 cent
// Blank Card will double the number of cents gained
func aPennyFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var n int8 = 1
		if blankCard {
			n *= 2
		}
		p.gainCents(n)
	}
	return f, false, nil
}

// Basic loot
// Gain 2 cent
// Blank Card will double the number of cents gained
func twoCentsFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var n int8 = 2
		if blankCard {
			n *= 2
		}
		p.gainCents(n)
	}
	return f, false, nil
}

// Basic loot
// Gain 3 cent
// Blank Card will double the number of cents gained
func threeCentsFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var n int8 = 3
		if blankCard {
			n *= 2
		}
		p.gainCents(n)
	}
	return f, false, nil
}

// Basic loot
// Gain 4 cent
// Blank Card will double the number of cents gained
func fourCentsFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var n int8 = 4
		if blankCard {
			n *= 2
		}
		p.gainCents(n)
	}
	return f, false, nil
}

// Basic loot
// Gain 5 cent
// Blank Card will double the number of cents gained
func aNickelFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var n int8 = 5
		if blankCard {
			n *= 2
		}
		p.gainCents(n)
	}
	return f, false, nil
}

// Basic loot
// Gain 10 cent
// Blank Card will double the number of cents gained
func aDimeFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var n int8 = 10
		if blankCard {
			n *= 2
		}
		p.gainCents(n)
	}
	return f, false, nil
}

// Basic loot
// Loot 3.
// Blank Card will double the loot gained
func aSackFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		n := 3
		if blankCard {
			n *= 2
		}
		for i := 0; i < n; i++ {
			p.loot(b.loot)
		}
	}
	return f, false, nil
}

// Basic loot
// Roll:
// 1: Everyone gains 1 cent.
// 2: Everyone gains loots 2.
// 3: Everyone takes 3 damage.
// 4: Everyone gains 4 cents.
// 5: Everyone loots 5.
// 6: Everyone gains 6 cents.
// Blank card will double gains / damages
func blankRuneFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		for i := 0; i < len(b.players); i++ {
			p := b.players[i]
			switch roll {
			case 1:
				var n int8 = 1
				if blankCard {
					n *= 2
				}
				p.gainCents(n)
			case 2:
				n := 2
				if blankCard {
					n *= 2
				}
				for i := 0; i < n; i++ {
					p.loot(b.loot)
				}
			case 3:
				var n uint8 = 3
				if blankCard {
					n *= 2
				}
				p.decreaseHP(n)
			case 4:
				var n int8 = 4
				if blankCard {
					n *= 2
				}
				p.gainCents(n)
			case 5:
				n := 5
				if blankCard {
					n *= 2
				}
				for i := 0; i < n; i++ {
					p.loot(b.loot)
				}
			case 6:
				var n int8 = 6
				if blankCard {
					n *= 2
				}
				p.gainCents(n)
			}
		}
	}
	return f, true, nil
}

// Basic loot
// Deal 1 damage to a Monster or Player.
// The blank card will double the damage dealt.
func bombFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	mCards := b.monster.getActiveMonsters()
	a := len(mCards)
	players := b.getPlayers(true)
	showMonsterCards(mCards, 0)
	showPlayers(players, a)
	ans := readInput(0, a+len(players)-1)
	var f lootCardEffect
	if ans < a {
		f = b.monster.bombHelper(p, b, mCards[ans], 1)
	} else {
		f = players[ans-a].bombHelper(p, b, 1)
	}
	return f, false, nil
}

// Basic loot
// Cancel the effect of any Active Item or Loot Card being played.
// The Blank Card active itemCard has no doubling effect on Butter Bean
// Check for the valid events on the event stack.
// If no valid effect is found, return an error.
// Else fizzle the event and any events that might have spawned from it.
func butterBeanFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	aIEvents := b.eventStack.getActivateItemEvents()
	lCEvents := b.eventStack.getLootCardEvents()
	events := mergeEventSlices(aIEvents, lCEvents)
	showEvents(events)
	ans := readInput(0, len(events)-1)
	node := events[ans]
	n := events[ans].event.e
	var err error
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		nextNodeEvent := node.next.event.e
		if _, ok := nextNodeEvent.(damageEvent); ok { // golden razor blade, bombs, troll bombs, etc
			_ = b.eventStack.fizzle(node.next)
		} else if _, ok := nextNodeEvent.(diceRollEvent); ok { // pills, high priestess, the d6
			_ = b.eventStack.fizzle(node.next)
		} else if _, ok := nextNodeEvent.(deathOfCharacterEvent); ok { // deathPenalty tarot value
			_ = b.eventStack.fizzle(node.next)
		}
		_ = b.eventStack.fizzle(node)
	}
	_, ok1 := n.(activateEvent)
	_, ok2 := n.(lootCardEvent)
	if !ok1 && !ok2 {
		err = errors.New("butter bean has no applicable target")
	}
	return f, false, err
}

// Basic loot
// Choose one:
// Destroy a curse.
// Prevent 1 damage to any player.
// Will return an activation error if there are
// no curses or damage events
// Blank card will double the amount of damage prevented, but does
// nothing for destroying curses.
func dagazFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var e error
	var f lootCardEffect
	damageEvents := b.eventStack.getDamageOfCharacterEvents()
	lde, lpc := len(damageEvents), len(p.Curses)
	if lde == 0 && lpc == 0 {
		e = errors.New("no requirements for dagaz met")
		return f, false, e
	} else if lpc > 0 && lde == 0 { // destroy curse only option
		f = p.dagazCurseHelper(b.monster, lpc)
	} else if lpc == 0 && lde > 0 {
		f = b.eventStack.preventDamageWithLootHelper(damageEvents, 1)
	} else {
		fmt.Println("Choose which effect to activate:\n" +
			"1) Destroy a curse\n" +
			"2) Prevent 1 Damage to a player.")
		ans := readInput(1, 2)
		if ans == 1 {
			f = p.dagazCurseHelper(b.monster, lpc)
		} else {
			f = b.eventStack.preventDamageWithLootHelper(damageEvents, 1)
		}
	}
	return f, false, e
}

// Tarot Card
// Kill a player
// The blank card has no effect
func deathTarotCardFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect
	players := b.getPlayers(true)
	l := len(players)
	var i uint8
	if l == 0 {
		return f, false, errors.New("no live players for deathPenalty tarot")
	} else if l > 1 {
		showPlayers(players, 0)
		i = uint8(readInput(0, l-1))
	}
	target := players[i]
	f = func(roll uint8, blankCard bool) { b.killPlayer(target) }
	return f, false, nil
}

// Basic Loot
// Reroll any dice roll.
// Will return an error if there are no dice roll events
// on the stack.
// The Blank Card has no effect.
func diceShardFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect
	var e error
	var i int
	nodes := b.eventStack.getDiceRollEvents()
	l := len(nodes)
	if l == 0 {
		e = errors.New("no dice roll events on the stack")
		return f, false, e
	} else if l > 1 {
		showEvents(nodes)
		i = readInput(0, l-1)
	}
	f = func(roll uint8, blankCard bool) {
		diceRollNode := nodes[i]
		if _, ok := diceRollNode.event.e.(diceRollEvent); ok { // double confirm
			diceRollNode.event = event{p: diceRollNode.event.p, e: diceRollEvent{n: uint8(rand.Intn(6) + 1)}}
		}
	}
	return f, false, e
}

// Basic loot
// Discard all active monsters not being attacked and replace them
// with cards from the top of the monster deck.
// The blank card has no effect
func ehwazFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	m := b.monster
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var i uint8
		for i = range m.zones {
			mCard := m.zones[i].peek()
			if !mCard.inBattle {
				card := m.zones[i].pop()
				m.discard(&card)
			}
		}
	}
	return f, false, nil
}

// Basic loot
// Deal 3 damage to a Monster or Player.
// The blank card will double the damage dealt.
func goldBombFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	mCards := b.monster.getActiveMonsters()
	a := len(mCards)
	players := b.getPlayers(true)
	showMonsterCards(mCards, 0)
	showPlayers(players, a)
	ans := readInput(0, a+len(players)-1)
	var f lootCardEffect
	if ans < a {
		f = b.monster.bombHelper(p, b, mCards[ans], 3)
	} else {
		f = players[ans-a].bombHelper(p, b, 3)
	}
	return f, false, nil
}

// Basic Loot
// If a Player would die, prevent that deathPenalty and end that player's turn.
// The blank card has no effect
func holyCardFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	deathEvents := b.eventStack.getDeathOfCharacterEvents()
	l := len(deathEvents)
	if l == 0 {
		return nil, false, errors.New("no deaths to prevent")
	}
	var i uint8
	if l > 1 {
		showEvents(deathEvents)
		i = uint8(readInput(0, l-1))
	}
	node := deathEvents[i]
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		if err := b.eventStack.fizzle(node); err == nil {
			target := node.event.p
			if target == &b.players[b.api] {
				b.forceEndOfTurn()
			}
		}
	}
	return f, false, nil
}

// Tarot Card
// Choose the player with the most souls or tied for the most souls.
// That player discards a Soul card they control.
// The blank card doubles the amount of souls lost.
func judgementFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	soulsMap, mostSouls := b.judgementHelper()
	if mostSouls == 0 {
		return nil, false, errors.New("no one has a soul value")
	}
	players := soulsMap[mostSouls]
	l := len(players)
	var i uint8
	if l > 1 {
		showPlayers(players, 0)
		i = uint8(readInput(0, l-1))
	}
	target := players[i]
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var i, n uint8 = 0, 1
		if blankCard {
			n *= 2
		}
		for i = 0; i < n; i++ {
			fmt.Println("Discard a soul value.")
			showSouls(target.Souls, target.Character.name, 0)
			ans := uint8(readInput(0, len(target.Souls)-1))
			card := p.popSoul(ans)
			b.discard(card)
		}
	}
	return f, false, nil
}

// Tarot Card
// Choose a Player: Gain Loot and Cents up to the amount that Player has.
// The blank card has no effect.
func justiceFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	players := b.getOtherPlayers(p, false)
	l := len(players)
	var i uint8
	if l > 1 {
		showPlayers(players, 0)
		i = uint8(readInput(0, l-1))
	}
	target := players[i]
	loot := b.loot
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		if p.Pennies < target.Pennies {
			p.gainCents(target.Pennies - p.Pennies)
		}
		diff := int8(len(target.Hand) - len(p.Hand))
		for i := diff; i > 0; i-- {
			p.loot(loot)
		}
	}
	return f, false, nil
}

// Basic loot
// Recharge an itemCard
// The blank card has no effect.
func lilBatteryFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		cards := p.getTappedActiveItems()
		l := len(cards)
		if l > 0 {
			showTreasureCards(cards, p.Character.name, 0)
			ans := readInput(0, l-1)
			cards[ans].recharge()
		}
	}
	return f, false, nil
}

// Basic loot
// Gain this soul
// The blank card has no effect
func lostSoulFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		p.addSoulToBoard(lootCard{baseCard: baseCard{name: "Lost Soul", effect: "Gain this Soul.", id: lostSoul}})
	}
	return f, false, nil
}

// Basic loot
// Choose a player, recharge all of their items.
// The blank card has no effect.
func megaBatteryFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		for _, card := range p.ActiveItems {
			card.recharge()
		}
	}
	return f, false, nil
}

// Basic loot
// Roll:
// 1-2: Draw 2 loot. 3-4: Draw 4 loot. 5-6: Discard 1 loot.
func pillsBlueFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var i, n uint8 = 0, 1
		if roll == 1 || roll == 2 {
			if blankCard {
				n *= 2
			}
			for i = 0; i < n; i++ {
				p.loot(b.loot)
			}
		} else if roll == 3 || roll == 4 {
			n = 3
			if blankCard {
				n *= 2
			}
			for i = 0; i < n; i++ {
				p.loot(b.loot)
			}
		} else if roll == 5 || roll == 6 {
			if blankCard {
				n *= 2
			}
			for i = 0; i < n; i++ {
				if len(p.Hand) > 0 {
					showLootCards(p.Hand, p.Character.name, 0)
					fmt.Println("Discard a value.")
					ans := uint8(readInput(0, len(p.Hand)))
					b.loot.discard(p.popHandCard(ans))
				}
			}
		}
	}
	return f, true, nil
}

// Basic loot
// Roll:
// 1-2: +1 AP till the end of the turn.
// 3-4: +1 HP till the end of the turn.
// 5-6: Take 1 damage.
func pillsRedFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var n uint8 = 1
		if blankCard {
			n = 2
		}
		if roll == 1 || roll == 2 {
			p.increaseAP(n)
		} else if roll == 3 || roll == 4 {
			p.increaseHP(n)
		} else if roll == 5 || roll == 6 {
			b.damagePlayerToPlayer(p, p, 1)
		}
	}
	return f, true, nil
}

// Basic loot
// Roll:
// 1-2: Gain 4 cents. 3-4: Gain 7 cents. 5-6: Lose 8 cents.
func pillsYellowFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var n int8 = 4 // default to resulitng roll of
		var f func(n int8)
		if roll == 1 || roll == 2 {
			f = p.gainCents
		} else if roll == 3 || roll == 4 {
			n, f = 7, p.gainCents
		} else if roll == 5 || roll == 6 {
			f = p.loseCents
		} else {
			panic("impossible dice roll result.")
		}
		f(n)
	}
	return f, true, nil
}

// Basic Loot
// Prevent 1 damage to any Player.
func soulHeartFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	damageEvents := b.eventStack.getDamageOfCharacterEvents()
	l := len(damageEvents)
	var i uint8
	if l == 0 {
		return nil, false, errors.New("no damage events on the stack")
	} else if l > 1 {
		showEvents(damageEvents)
		i = uint8(readInput(0, len(damageEvents)-1))
	}
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var n uint8 = 1
		if blankCard {
			n = 2
		}
		_ = b.eventStack.preventDamage(n, damageEvents[i])
	}
	return f, false, nil
}

// Tarot Card
// A Player gains +1 Attack till the end of the turn and may attack an additional time.
// The blank card doubles the attack and number of times to attack.
func strengthFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var n uint8 = 1
		if blankCard {
			n = 2
		}
		p.increaseAP(n)
		p.numAttacks += int8(n)
	}
	return f, false, nil
}

// Tarot Card
// Choose one:
// Take 1 damage, gain 4 cents.
// Take 2 damage, gain 8 cents.
// The blank card should double the amount of damage and the reward.
func temperanceFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var take2damage bool
	fmt.Println("Choose One:\n1) Take 1 Damage: Gain 4 Cents.\n2) Take 2 Damage: Gain 8 Cents.")
	ans := uint8(readInput(1, 2))
	if ans == 2 {
		take2damage = true
	}
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		if checkActiveEffects(p.activeEffects, temperance, true) {
			var n int8 = 4 // assume one damage unless shown otherwise
			var centsGain = p.gainCents
			if p.Character.hp != 0 {
				if take2damage {
					n = 8
				}
				centsGain(n)
			}
		}

	}
	return f, take2damage, nil
}

// Tarot Card
// A Player gains +1 Attack and +1 Health till the end of the turn.
// The blank card doubles the buffs
func theChariotFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var n uint8 = 1
		if blankCard {
			n = 2
		}
		p.increaseAP(n)
		p.increaseHP(n)
	}
	return f, false, nil
}

// Tarot Card
// Destroy an Item you control:
// Steal any Item another Player controls, or any Item in the Shop.
// The Blank card has no effect.
func theDevilFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	items := p.getAllItems(false)
	l := len(items)
	if l == 0 {
		return nil, false, errors.New("no items to pay the cost")
	}
	showItems(items, 0)
	fmt.Println("Destroy which value?")
	ans := uint8(readInput(0, l-1))
	card := items[ans]
	i, _ := p.getItemIndex(card.getId(), card.isPassive())
	b.discard(p.popItemByIndex(i, card.isPassive()))
	var owners map[uint16]*player
	items, owners = b.getAllItems(false, p)
	l = len(items)
	showItems(items, 0)
	showTreasureCards(b.treasure.zones, "shop", l)
	fmt.Println("Which card to steal?")
	ans = uint8(readInput(0, l+len(b.treasure.zones)-1))
	card = items[ans]
	id, isPassive, f := card.getId(), card.isPassive(), func(roll uint8, blankCard bool) {}
	if owner, ok := owners[id]; ok { // The selected value is NOT in the shop.
		f = func(roll uint8, blankCard bool) {
			i, err := owner.getItemIndex(id, isPassive)
			if err == nil {
				p.addCardToBoard(owner.popItemByIndex(i, isPassive))
			}
		}
	} else {
		f = func(roll uint8, blankCard bool) {
			for i, c := range b.treasure.zones {
				if c.id == id {
					p.addCardToBoard(b.treasure.zones[i])
					b.treasure.zones[i] = treasureCard{}
					break
				}
			}
		}
	}
	return f, false, nil
}

// Tarot Card
// Look at the top 5 cards of the Monster deck.
// Put 4 on the bottom of the deck and one back on top.
// The blank card will double the number of cards to look at.
func theEmperorFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	m := b.monster
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var i, n uint8 = 0, 5
		if blankCard {
			n = 10
		}
		mCards := make([]monsterCard, n)
		for i = 0; i < n; i++ {
			mCards[i] = m.draw()
		}
		showMonsterCards(mCards, 0)
		fmt.Println("Which value should go on top?")
		ans := uint8(readInput(0, int(n-1)))
		for i = 0; i < n; i++ {
			if i == ans {
				m.placeInDeck(mCards[i], true)
			} else {
				m.placeInDeck(mCards[i], false)
			}
		}
	}
	return f, false, nil
}

// Tarot Card
// A player gains +1 attack and +1 to all dice rolls till the end of the turn.
// The blank card will double the stat boosting effects.
func theEmpressFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var n uint8 = 1
		if blankCard {
			n = 2
		}
		p.increaseAP(n)
		p.activeEffects[theEmpress] = struct{}{}
	}
	return f, false, nil
}

// Tarot Card
// End a Player's turn.
// Cancel any effects or Loot Cards that haven't resolved.
// The blank card has no effect
func theFoolFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		b.forceEndOfTurn()
	}
	return f, false, nil
}

// Tarot Card
// Look at the top card of all decks.
// You may put those cards on the bottom of their decks.
// Then Loot 2.
// The blank card doubles the number of loot cards drawn
func theHangedManFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		msg := "Choose what to do with each value.\n1) Place back on top of the deck.\n2) Place on the bottom of the deck."
		lc := b.loot.draw()
		fmt.Println(msg)
		ans := readInput(1, 2)
		switch ans {
		case 1:
			b.loot.placeInDeck(lc, true)
		case 2:
			b.loot.placeInDeck(lc, false)
		}
		mc := b.monster.draw()
		ans = readInput(1, 2)
		switch ans {
		case 1:
			b.monster.placeInDeck(mc, true)
		case 2:
			b.monster.placeInDeck(mc, false)
		}
		tc := b.treasure.draw()
		ans = readInput(1, 2)
		switch ans {
		case 1:
			b.treasure.placeInDeck(tc, true)
		case 2:
			b.treasure.placeInDeck(tc, false)
		}
		var i, n uint8 = 0, 2
		if blankCard {
			n = 4
		}
		for i = 0; i < n; i++ {
			p.loot(b.loot)
		}
	}
	return f, false, nil
}

// Tarot Card
// Look at the top 5 cards of the Treasure deck.
// Put 4 on the bottom of the deck and one back on top.
func theHermitFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	t := b.treasure
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var i, n uint8 = 0, 5
		if blankCard {
			n = 10
		}
		tCards := make([]treasureCard, n)
		for i = 0; i < n; i++ {
			tCards[i] = t.draw()
		}
		showMonsterCards(tCards, 0)
		fmt.Println("Which value should go on top?")
		ans := uint8(readInput(0, int(n-1)))
		for i = 0; i < n; i++ {
			if i == ans {
				t.placeInDeck(tCards[i], true)
			} else {
				t.placeInDeck(tCards[i], false)
			}
		}
	}
	return f, false, nil
}

// Tarot Card
// Prevent up to 2 damage done to a Player or Monster.
// The blank card doubles the damage prevented.
func theHierophantFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	damage := b.eventStack.getDamageEvents()
	l := len(damage)
	if l == 0 {
		return nil, false, errors.New("no damage events")
	}
	var i uint8
	if l > 1 {
		showEvents(damage)
		i = uint8(readInput(0, l-1))
	}
	es := &b.eventStack
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var n uint8 = 2
		if blankCard {
			n = 4
		}
		_ = es.preventDamage(n, damage[i])
	}
	return f, false, nil
}

// Tarot Card
// Choose a Player or Monster. Then roll:
// Deal damage to the target equal to the number rolled.
func theHighPriestessFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	monsters := b.monster.getActiveMonsters()
	l := len(monsters)
	players := b.getPlayers(true)
	showMonsterCards(monsters, 0)
	showPlayers(players, l)
	ans := readInput(0, l+len(players)-1)
	var c combatTarget
	if ans < l {
	}
	if ans < l { // Damage a monster
		c = monsters[ans]
	} else {
		c = players[ans-l]
	}
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		if blankCard {
			roll *= 2
		}
		if monster, ok := c.(*monsterCard); ok {
			b.damagePlayerToMonster(p, monster, roll, 0)
		} else if target, ok := c.(*player); ok {
			b.damagePlayerToPlayer(p, target, roll)
		}
	}
	return f, true, nil
}

// Tarot Card
// A Player gains +2 Health till the end of the turn.
// The blank card doubles the amount of HP gained
func theLoversFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var n uint8 = 2
		if blankCard {
			n = 4
		}
		p.increaseHP(n)
	}
	return f, false, nil
}

// Tarot Card
// Change the result of a dice roll to the number of your choosing
// The blank card has no effect
func theMagicianFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	diceRolls := b.eventStack.getDiceRollEvents()
	max := len(diceRolls)
	if max == 0 {
		return nil, false, errors.New("no dice roll events on stack")
	}
	var i uint8
	if max > 1 {
		showEvents(diceRolls)
		i = uint8(readInput(0, max-1))
	}
	fmt.Println("Enter new roll.")
	ans := uint8(readInput(1, 6))
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		diceRolls[i].event.e = diceRollEvent{n: ans}
	}
	return f, false, nil
}

// Tarot Card
// Look at the top 5 cards of the Loot deck.
// Put 4 on the bottom of the deck and one back on top.
// The blank Card doubles the number of peeked cards
func theMoonFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	l := b.loot
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var i, n uint8 = 0, 5
		if blankCard {
			n = 10
		}
		lCards := make([]lootCard, n)
		for i = 0; i < n; i++ {
			lCards[i] = l.draw()
		}
		showMonsterCards(lCards, 0)
		fmt.Println("Which value should go on top?")
		ans := uint8(readInput(0, int(n-1)))
		for i = 0; i < n; i++ {
			if i == ans {
				l.placeInDeck(lCards[i], true)
			} else {
				l.placeInDeck(lCards[i], false)
			}
		}
	}
	return f, false, nil
}

// Tarot card
// Gain 1 treasure.
// The blank card doubles the number of treasure gained.
func theStarsFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var t = b.treasure
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var i, n uint8 = 0, 1
		for i = 0; i < n; i++ {
			card := t.draw()
			p.addCardToBoard(card)
		}
	}
	return f, false, nil
}

// Tarot Card
// If it is your turn, gain an additional turn after this one.
// TODO: Make the blank card work here.
func theSunFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		p.activeEffects[theSun] = struct{}{}
	}
	return f, false, nil
}

// Tarot card
// roll:
// 1-2: All players take 1 damage.
// 3-4: All Monsters take 1 damage.
// 5-6: All Players take 2 damage.
func theTowerFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		var n uint8 = 1
		if blankCard {
			n = 2
		}
		if roll == 1 || roll == 2 {
			for i := range b.players {
				b.damagePlayerToPlayer(p, &b.players[i], n)
			}
		} else if roll == 3 || roll == 4 {
			for i := range b.monster.zones {
				m := b.monster.zones[i].peek()
				d := damageEvent{target: m, n: n}
				b.eventStack.push(event{p: p, e: d})
			}
		} else if roll == 5 || roll == 6 {
			if blankCard {
				n = 4
			}
			for i := range b.players {
				b.damagePlayerToPlayer(p, &b.players[i], n)
			}
		}
		roll = 0
	}
	return f, true, nil
}

// Tarot Card
// Look at all Players' hands, then loot 2.
// The blank card doubles the amount looted
func theWorldFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		for _, player := range b.getOtherPlayers(p, false) {
			showLootCards(player.Hand, player.Character.name, 0)
		}
		var i, n uint8 = 0, 2
		if blankCard {
			n = 4
		}
		for i = 0; i < n; i++ {
			p.loot(b.loot)
		}
	}
	return f, false, nil
}

// Tarot Card
// roll:
// 1: Gain 1 Cent.
// 2: Take 2 damage.
// 3: Loot 3.
// 4: Lose 4 Cents.
// 5: Gain 5 Cents.
// 6: Gain 1 Treasure.
// The blank card doubles the effects of each roll.
func wheelOfFortuneFunc(p *player, b *Board) (lootCardEffect, bool, error) {
	var f lootCardEffect = func(roll uint8, blankCard bool) {
		switch roll {
		case 1:
			if blankCard {
				roll *= 2
			}
			p.gainCents(int8(roll))
		case 2:
			if blankCard {
				roll *= 2
			}
			b.damagePlayerToPlayer(p, p, roll)
		case 3:
			if blankCard {
				roll *= 2
			}
			var i uint8
			for i = 0; i < roll; i++ {
				p.loot(b.loot)
			}
		case 4:
			if blankCard {
				roll *= 2
			}
			p.loseCents(int8(roll))
		case 5:
			if blankCard {
				roll *= 2
			}
			p.gainCents(int8(roll))
		case 6:
			var i uint8
			if blankCard {
				roll *= 2
			}
			for i = 0; i < roll; i++ {
				tc := b.treasure.draw()
				p.addCardToBoard(tc)
			}
		}
	}
	return f, true, nil
}

// Event based passive
// Each time a player dies, loot 1
func bloodyPennyEvent(p *player, b *Board, lCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDeath(nil); err == nil {
		f = func(roll uint8) { p.loot(b.loot) }
	}
	return f, false, err
}

// Preventative event based passive
// Each time you die, roll: 1-5: You Die. 6: Prevent deathPenalty. If it was your turn, end it.
func brokenAnkhChecker(p *player) {}

// Event based passive
// At the start of your turn, look at the top card of the loot deck.
// You may put it on the bottom
func cainsEyeEvent(p *player, b *Board, lCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkStartOfTurn(p); err == nil {
		f = b.peekTrinketHelper(cainsEye)
	}
	return f, false, err
}

// Conditional constant item
// Each time you gain cents, gain an additional 1 cent.
func counterfeitPennyChecker(p *player) {
	if _, err := p.getItemIndex(counterfeitPenny, true); err == nil {
		p.Pennies += 1
	}
}

// Constant Passive
// Gain +1 Attack for the first attack roll of your turn.
func curvedHornConstant(p *player, b *Board, lCard card, isLeaving bool) {
	if !isLeaving {
		p.activeEffects[curvedHorn] = struct{}{}
	} else {
		delete(p.activeEffects, curvedHorn)
	}
}

// Event based passive
// At the start of your turn, look at the top card of the Treasure deck.
// You may put it on the bottom.
func goldenHorseShoeEvent(p *player, b *Board, lCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkStartOfTurn(p); err == nil {
		f = b.peekTrinketHelper(goldenHorseShoe)
	}
	return f, false, err
}

// Preventative event based passive
// Each time you take damage, Roll:
// 1-5: Nothing. 6: Prevent 1 damage.
func guppysHairballChecker(es *eventStack, damageNode *eventNode) cardEffect {
	return func(roll uint8) {
		if roll == 6 {
			_ = es.preventDamage(1, damageNode)
		}
	}
}

// Event based passive
// At the start of your turn, look at the top card of the Monster deck.
// You may put it on the bottom.
func purpleHeartEvent(p *player, b *Board, lCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkStartOfTurn(p); err == nil {
		f = b.peekTrinketHelper(purpleHeart)
	}
	return f, false, err
}

// Event based passive
// Each time you take damage, gain 1 cent.
func swallowedPennyEvent(p *player, b *Board, lCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if _, err = en.checkDamageToPlayer(p.Character.id); err == nil {
		f = func(roll uint8) { p.gainCents(1) }
	}
	return f, false, err
}
