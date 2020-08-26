package four_souls

import (
	"errors"
	"fmt"
	"math/rand"
)

func giveCurseHelper(ap *player, b *Board, mCard card) (cardEffect, bool, error) {
	others := b.getOtherPlayers(ap, false)
	showPlayers(others, 0)
	l, i := len(others), uint8(0)
	if l > 1 {
		i = uint8(readInput(0, l-1))
	}
	var f cardEffect = func(roll uint8) {
		fmt.Println("Giving curse to ", others[i].Character.name)
		others[i].addCardToBoard(mCard.(monsterCard))
	}
	return f, false, nil
}

func rewardCentsHelper(b *Board, n int8) (cardEffect, bool) {
	ap := b.getActivePlayer()
	rollRequired := n == 0
	var f cardEffect = func(roll uint8) {
		if rollRequired {
			n = int8(roll)
		}
		ap.gainCents(n)
	}
	return f, rollRequired
}

func rewardLootHelper(b *Board, n uint8) (cardEffect, bool) {
	ap := b.getActivePlayer()
	rollRequired := n == 0
	var f cardEffect = func(roll uint8) {
		var i uint8
		if rollRequired {
			n = roll
		}
		for i = 0; i < n; i++ {
			ap.loot(b.loot)
		}
	}
	return f, rollRequired
}

func rewardTreasureHelper(b *Board, n uint8) (cardEffect, bool) {
	ap := b.getActivePlayer()
	var f cardEffect = func(roll uint8) {
		var i uint8
		for i = 0; i < n; i++ {
			ap.addCardToBoard(b.treasure.draw())
		}
	}
	return f, false
}

func rewardLootAndCents(b *Board, loot, cents int8) (cardEffect, bool) {
	ap := b.getActivePlayer()
	var f cardEffect = func(roll uint8) {
		ap.gainCents(cents)
		var i int8
		for i = 0; i < loot; i++ {
			ap.loot(b.loot)
		}
	}
	return f, false
}

func rewardMegaBoss(b *Board) (cardEffect, bool) {
	ap := b.getActivePlayer()
	var f cardEffect = func(roll uint8) {
		ap.addCardToBoard(b.treasure.draw())
		ap.gainCents(6)
	}
	return f, false
}

// Basic enemy
// When this dies, you may attack the monster deck an additional time
func bigSpiderDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	ap := b.getActivePlayer()
	return func(roll uint8) { ap.numAttacks += 1 }, false, nil
}

// Loot 1
func bigSpiderReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 1)
}

// Basic enemy
// When this dies, it deals 1 damage to the player that killed it
func blackBonyDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	return func(roll uint8) { b.damagePlayerToPlayer(p, p, 1) }, false, nil
}

// Roll: Loot X
func blackBonyReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 0)
}

// When this dies, it deals 1 damage to all players.
func boomFlyDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	return func(roll uint8) {
		for _, player := range b.getPlayers(true) {
			b.damagePlayerToPlayer(p, player, 1)
		}
	}, false, nil
}

func boomFlyReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 4)
}

func clottyReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 4)
}

func codWormReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 3)
}

func conjoinedFattyReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 2)
}

// Basic enemy
// When this dies, force a player to discard 2 loot cards.
func dankGlobinDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect
	var err error = errors.New("no player with hand size greater than 2")
	players := b.getPlayers(false)
	valid := make([]*player, 0, len(players))
	for _, player := range players {
		if len(player.Hand) >= 2 {
			valid = append(valid, player)
		}
	}
	l := len(valid)
	if l > 0 {
		err = nil
		var i uint8
		if l > 1 {
			showPlayers(players, 0)
			fmt.Println("Choose a player to discard cards")
			i = uint8(readInput(0, l-1))
		}
		target := players[i]
		f = func(roll uint8) {
			if len(target.Hand) >= 2 {
				fmt.Println("Discard 2 cards")
				for i := 0; i < 2; i++ {
					showLootCards(target.Hand, target.Character.name, 0)
					b.discard(target.popHandCard(uint8(readInput(0, len(target.Hand)-1))))
				}
			}
		}
	}
	return f, false, err
}

func dankGlobinReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 2)
}

// Basic Enemy
// When this kis killed on a roll of 6, double its rewards.
func dingaReward(b *Board) (cardEffect, bool) {
	return func(roll uint8) {
		var double = roll == 6
		if double {
			roll *= 2
		}
		b.players[b.api].gainCents(int8(roll))
	}, true
}

func dipReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 1)
}

// Basic Enemy
// Event based passive
// Any damage done to this is also done to the player to your left (the next player)
func dopleEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var damage damageEvent
	var err error
	if damage, err = en.checkDamageToSpecificMonster(mCard.getId()); err == nil {
		f = func(roll uint8) { b.damagePlayerToPlayer(p, b.getNextPlayer(p), damage.n) }
	}
	return f, false, err
}

func dopleReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 7)
}

// Basic Enemy
// Event based passive
// Any damage done to this is also done to the player to your left (the next player)
func evilTwinEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	return dopleEvent(p, b, mCard, en)
}

func evilTwinReward(b *Board) (cardEffect, bool) {
	return rewardTreasureHelper(b, 1)
}

func fatBatReward(b *Board) (cardEffect, bool) {
	return rewardTreasureHelper(b, 1)
}

func fattyReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 1)
}

func flyReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 1)
}

// Basic Enemy
// When this dies, force a player to lose 7 cents.
func greedlingDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect
	var err error = errors.New("no player with at least 7 cents")
	players := b.getPlayers(false)
	valid := make([]*player, 0, len(players))
	for _, player := range players {
		if player.Pennies >= 7 {
			valid = append(valid, player)
		}
	}
	l := len(valid)
	if l > 0 {
		err = nil
		var i uint8
		if l > 1 {
			showPlayers(players, 0)
			fmt.Println("Choose a player to lose cents")
			i = uint8(readInput(0, l-1))
		}
		target := players[i]
		f = func(roll uint8) { target.loseCents(7) }
	}
	return f, false, err
}

func greedlingReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 7)
}

// Basic enemy
// When this dies, add an additional item from the top of the Treasure deck to the Shop.
func hangerDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	return func(roll uint8) { b.treasure.zones = append(b.treasure.zones, b.treasure.draw()) }, false, nil
}

func hangerReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 7)
}

// Basic enemy
// Prevent any damage done to this on a roll of 6
func hopperEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(6); err == nil {
		nextNode := b.eventStack.peek()
		if _, ok := nextNode.event.e.(damageEvent); ok {
			f = func(roll uint8) { nextNode.event.e = fizzledEvent{} }
		} else {
			err = errors.New("not a damage event")
		}
	}
	return f, false, err
}

func hopperReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 3)
}

// Basic enemy
// Constant passive with conditions
// This deals 1 additional damage whenever the player attacking it rolls a 2.
//
// Handled in the battle step
func horfChecker(mId uint16, roll uint8) bool {
	return mId == horf && roll == 2
}

func horfReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 3)
}

// Basic enemy
// When this deals damage to a player, they also lose 2 cents.
func keeperHeadEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if _, err = en.checkDamageToPlayerFromMonster(p.Character.id, mCard.getId()); err == nil {
		f = func(roll uint8) { p.loseCents(2) }
	}
	return f, false, err
}

func keeperReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 0)
}

// Basic enemy
// Constant passive with condition
// This deals double damage on a roll of 1
func leaperChecker(mId uint16, roll uint8) bool {
	return mId == leaper && roll == 1
}

func leaperReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 5)
}

func leechReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 1)
}

// Basic enemy
// When this dies, you may steal an item from a player
func momsDeadHandDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	items, playerMap := b.getAllItems(false, p)
	l := len(items)
	if l == 0 {
		err = errors.New("no items to steal")
	} else {
		showItems(items, 0)
		fmt.Println(fmt.Sprintf("%d) Don't steal", l))
		ans := readInput(0, len(items))
		if ans == l {
			err = errors.New("decided not to steal")
		} else {
			item, target := items[ans], playerMap[items[ans].getId()]
			f = func(roll uint8) {
				if i, err := target.getItemIndex(item.getId(), item.isPassive()); err == nil {
					p.addCardToBoard(target.popItemByIndex(i, item.isPassive()))
				}
			}
		}
	}
	return f, false, err
}

func momsDeadHandReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 4)
}

// Basic enemy
// When this dies, you may look at a player's hand.
func momsEyeDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	fmt.Println("1) Look at player's hand\n2) Do nothing")
	ans := readInput(1, 2)
	if ans == 2 {
		err = errors.New("decided to not look")
	} else {
		others := b.getOtherPlayers(p, false)
		var i uint8
		if len(others) > 1 {
			showPlayers(others, 0)
			i = uint8(readInput(0, len(others)-1))
		}
		f = func(roll uint8) { showLootCards(others[i].Hand, others[i].Character.name, 0) }
	}
	return f, false, err
}

func momsEyeReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 1)
}

// Basic enemy
// When the attacking player rolls a 6, cancel combat and end the turn
func momsHandEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = p.checkAttackingPlayer(); err == nil {
		if err = en.checkDiceRoll(6); err == nil {
			f = func(roll uint8) { b.forceEndOfTurn() }
		}
	}

	return f, false, err
}

func momsHandReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 4)
}

// Basic enemy
// When this dies, deal 3 damage to any player.
func mulliboomDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	players := b.getPlayers(true)
	showPlayers(players, 0)
	fmt.Println("Choose who receives 3 damage")
	ans := readInput(0, len(players)-1)
	var f cardEffect = func(roll uint8) { b.damagePlayerToPlayer(p, players[ans], 3) }
	return f, false, nil
}

func mulliboomReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 6)
}

// Basic enemy
// When this dies, expand the number of active monsters by 1
func mulliganDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	return func(roll uint8) { b.monster.zones = append(b.monster.zones, activeSlot{}) }, false, nil
}

func mulliganReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 3)
}

func paleFattyReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 6)
}

func pooterReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 1)
}

// basic enemy
// When this dies, you must attack the monster deck an additional time
func portalDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect
	var err = errors.New("not the active player")
	if p.Character.id == b.players[b.api].Character.id {
		err, f = nil, func(roll uint8) {
			p.forceAttack = true
			p.numAttacks += 1
			p.activeEffects[portal] = struct{}{}
		}
	}
	return f, false, err
}

func portalReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 3)
}

// Basic Enemy
// When this dies, recharge all of your Active Items
func psyHorfDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	items := p.getActiveItems(true)
	return func(roll uint8) {
		for _, c := range items {
			c.recharge()
		}
	}, false, nil
}

func psyHorfReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 1)
}

// Basic enemy
// Whenever this deals damage, it also deals damage to the player to your right (previous)
func rageCreepEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	var damage damageEvent
	if damage, err = en.checkDamageToPlayerFromMonster(p.Character.id, mCard.getId()); err == nil {
		f = func(roll uint8) {
			pp := b.getPreviousPlayer(p.Character.id)
			b.damagePlayerToPlayer(p, pp, damage.n)
		}
	}
	return f, false, err
}

func rageCreepReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 2)
}

func redHostReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 5)
}

// Basic enemy
// When the attacking player rolls a 3, they must steal a loot card at random from another player
func ringOfFliesEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = p.checkAttackingPlayer(); err == nil {
		if err = en.checkDiceRoll(3); err == nil {
			others := b.getOtherPlayers(p, false)
			l := len(others)
			var i uint8
			if l > 1 {
				showPlayers(others, 0)
				i = uint8(readInput(0, l-1))
			}
			target := others[i]
			f = func(roll uint8) { p.Hand = append(p.Hand, target.popHandCard(uint8(rand.Intn(l)))) }
		}
	}
	return f, false, err
}

func ringOfFliesReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 3)
}

func spiderReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 1)
}

func squirtReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 1)
}

// Special enemy
// All monsters gain +1 Dice Roll while this is active.
// This can't be attacked.
// When another active Monster dies, this dies.
func stoneyEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	var intention intentionToAttackEvent
	if intention, err = en.checkIntentionToAttack(); err == nil {
		f = func(roll uint8) { intention.m.roll += 1 }
	}
	return f, false, err
}

func stoneyReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 1)
}

// Basic enemy
// Each time the attacking player rolls a 5, they take 1 damage
func swarmOfFliesEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = p.checkAttackingPlayer(); err == nil {
		if err = en.checkDiceRoll(5); err == nil {
			f = func(roll uint8) { b.damagePlayerToPlayer(p, p, 1) }
		}
	}
	return f, false, err
}

func swarmOfFliesReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 5)
}

func triteReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 2)
}

// Basic Enemy
// When this dies, you may force a player to discard a Soul card.
func wizoobDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	fmt.Println("1) Force a player to discard a soul\n2) Do nothing.")
	if readInput(1, 2) == 1 {
		souls, playerMap := b.getSouls()
		if len(playerMap) == 0 {
			err = errors.New("no souls to collect")
		}
		showSoulsByPlayer(souls, playerMap, 0)
		i := readInput(0, len(souls)-1)
		f = func(roll uint8) {
			target := playerMap[souls[i].getId()]
			if j, err := target.getSoulIndex(souls[i].getId()); err == nil {
				p.addSoulToBoard(target.popSoul(j))
			}
		}
	} else {
		err = errors.New("player declined")
	}
	return f, false, err
}

func wizoobReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 3)
}

// Basic enemy
// When any player rolls a 5, they discard a loot card
func cursedFattyEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(5); err == nil {
		f = func(roll uint8) {
			showLootCards(p.Hand, "self", 0)
			fmt.Println("Discard one")
			b.loot.discard(p.popHandCard(uint8(readInput(0, len(p.Hand)-1))))
		}
	}
	return f, false, nil
}

func cursedFattyReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 2)
}

// Basic enemy
// When any player rolls a 4, all active monsters gain +1 attack till the end of the turn
func cursedGaperEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(4); err == nil {
		f = func(roll uint8) {
			for _, m := range b.monster.getActiveMonsters() {
				m.increaseAP(1)
			}
		}
	}
	return f, false, err
}

func cursedGaperReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 3)
}

// Basic enemy
// When any player rolls a 2 that player takes 2 damage
func cursedHorfEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(2); err == nil {
		f = func(roll uint8) { b.damagePlayerToPlayer(p, p, 2) }
	}
	return f, false, nil
}

func cursedHorfReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 3)
}

// Basic Enemy
// When any player rolls a 1 they lose 2 cents
func cursedKeeperHeadEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(1); err == nil {
		f = func(roll uint8) { p.loseCents(2) }
	}
	return f, false, err
}

func cursedKeeperHeadReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 0)
}

// Basic enemy
// When any player rolls a 6 end that player's turn
func cursedMomsHandEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(6); err == nil && b.players[b.api].Character.id == p.Character.id {
		f = func(roll uint8) { b.forceEndOfTurn() }
	}
	return f, false, err
}

func cursedMomsHandReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 4)
}

// Basic enemy
// When any player activates an item, they take 1 damage
func cursedPsyHorfEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkActivateItemEvent(); err == nil && !p.isDead() {
		f = func(roll uint8) { b.damagePlayerToPlayer(p, p, 1) }
	}
	return f, false, err
}

func cursedPsyHorfReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 2)
}

// Basic Enemy
// When any player rolls a 6 they heal 1 HP
func holyDingaEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(6); err == nil {
		f = func(roll uint8) { p.heal(1) }
	}
	return f, false, err
}

func holyDingaReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 0)
}

// Basic enemy
// When any player rolls a 1 that player gains 1 cent
func holyDipEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(1); err == nil {
		f = func(roll uint8) { p.gainCents(1) }
	}
	return f, false, err
}

func holyDipReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 1)
}

// Basic enemy
// When any player rolls a 4 they gain 2 cents.
func holyKeeperHeadEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(4); err == nil {
		f = func(roll uint8) { p.gainCents(2) }
	}
	return f, false, err
}

func holyKeeperHeadReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 0)
}

// Basic enemy
// When any player rolls a 2 that player may recharge an item
func holyMomsEyeEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(2); err == nil {
		items := p.getTriggeredActiveItems()
		l := len(items)
		if l > 0 {
			showTreasureCards(items, "self", 0)
			fmt.Println(fmt.Sprintf("%d) Do not recharge", l))
			if i := readInput(0, l); i != l {
				f = func(roll uint8) { items[i].recharge() }
			}
		}
	}
	return f, false, err
}

func holyMomsEyeReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 1)
}

// Basic enemy
// When any player rolls a 5 that player loots 1.
func holySquirtEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkDiceRoll(5); err == nil {
		f = func(roll uint8) { p.loot(b.loot) }
	}
	return f, false, err
}

func holySquirtReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 2)
}

func carrionQueenReward(b *Board) (cardEffect, bool) {
	return rewardTreasureHelper(b, 1)
}

// Boss
// Whenever the attacking player rolls a 1, this heals 2 HP
func chubEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = p.checkAttackingPlayer(); err == nil {
		if err = en.checkDiceRoll(1); err == nil {
			m := mCard.(*monsterCard)
			f = func(roll uint8) { m.heal(2) }
		}
	}

	return f, false, err
}

func chubReward(b *Board) (cardEffect, bool) {
	return rewardLootAndCents(b, 2, 3)
}

// Boss
// When this dies, the active Player must make an additional attack this turn
func conquestDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	return func(roll uint8) {
		ap := &b.players[b.api]
		ap.forceAnAttack(true)
	}, false, nil
}

func conquestReward(b *Board) (cardEffect, bool) {
	return rewardLootAndCents(b, 2, 3)
}

// Boss
// Each time the attacking player rolls a 1, all monsters gain +1 Dice Roll till the end of the turn.
func daddyLongLegsEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = p.checkAttackingPlayer(); err == nil {
		if err = en.checkDiceRoll(1); err == nil {
			f = func(roll uint8) {
				for _, m := range b.monster.getActiveMonsters() {
					m.roll += 1
				}
			}
		}
	}

	return f, false, err
}

func daddyLongLegsReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 7)
}

// Boss
// Whenever this takes damage, it gets +1 Attack till the end of the turn
func darkOneEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	m := mCard.(*monsterCard)
	_, err := en.checkDamageToSpecificMonster(mCard.getId())
	if err == nil {
		f = func(roll uint8) { m.increaseAP(1) }
	}
	return f, false, err
}

func darkOneReward(b *Board) (cardEffect, bool) {
	return rewardTreasureHelper(b, 1)
}

// Boss
// When this dies, the Active Player must kill a player.
func deathMonsterDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	players := b.getPlayers(true)
	showPlayers(players, 0)
	fmt.Println("Who dies?")
	ans := readInput(0, len(players)-1)
	return func(roll uint8) { b.killPlayer(players[ans]) }, false, nil
}

func deathMonsterReward(b *Board) (cardEffect, bool) {
	return rewardTreasureHelper(b, 1)
}

// Mini Boss?
// When this dies, place it back inside the Monster deck 6 cards from the top.
// Other Monsters gain +1 Dice Roll while this is active
//
// First effect is handled elsewhere
func deliriumDeathHandler(deliriumCard monsterCard, mA *mArea) {
	if deliriumCard.id == delirium {
		cards := make([]monsterCard, 6, 6)
		for i := 0; i < 6; i++ {
			cards[i] = mA.draw()
		}
		mA.placeInDeck(deliriumCard, true)
		for i := 5; i >= 0; i-- {
			mA.placeInDeck(cards[i], true)
		}
	}
}

func deliriumEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	intention, err := en.checkIntentionToAttack()
	if err == nil {
		f = func(roll uint8) { intention.m.roll += 1 }
	}
	return f, false, err
}

func deliriumReward(b *Board) (cardEffect, bool) {
	return rewardTreasureHelper(b, 2)
}

// Boss
// When this dies, you must make an additional attack
func envyDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect
	err := errors.New("not the active player")
	if p.Character.id == b.players[b.api].Character.id {
		err, f = nil, func(roll uint8) { p.forceAnAttack(true) }
	}
	return f, false, err
}

func envyReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 1)
}

// Boss
// When this dies, the Active Player skips their next turn.
func famineDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	ap := &b.players[b.api]
	return func(roll uint8) { ap.activeEffects[famine] = struct{}{} }, false, nil
}

func famineReward(b *Board) (cardEffect, bool) {
	return rewardLootAndCents(b, 2, 3)
}

// Boss
// When this is at 1 HP, it gains +1 Attack till the end of turn.
func geminiEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if _, err = en.checkDamageToSpecificMonster(gemini); err == nil {
		m := mCard.(*monsterCard)
		if m.hp == 1 {
			f = func(roll uint8) { m.increaseAP(1) }
		}
	}
	return f, false, err
}

func geminiReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 5)
}

// When this takes damage on a roll of 6, deal 1 damage to the player to your left
func gluttonyEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if _, err = en.checkDamageToSpecificMonster(gluttony); err == nil {
		if en.event.roll == 6 {
			pp := b.getPreviousPlayer(p.Character.id)
			m := mCard.(*monsterCard)
			f = func(roll uint8) { b.damageMonsterToPlayer(m, pp, 1, 0) }
		}
	}
	return f, false, err
}

func gluttonyReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 2)
}

// Boss
// When this deals damage all players lose 4 cents
func greedMonsterEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	_, err := en.checkDamageFromMonster(greedMonster)
	if err == nil {
		f = func(roll uint8) {
			for _, p := range b.getPlayers(false) {
				p.loseCents(4)
			}
		}

	}
	return f, false, err
}

func greedMonsterReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 9)
}

// Boss
// Each time the Attacking Player activates an Item, they take 1 damage.
func gurdyJrEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = p.checkAttackingPlayer(); err == nil {
		if err = en.checkActivateItemEvent(); err == nil {
			m := mCard.(*monsterCard)
			f = func(roll uint8) { b.damageMonsterToPlayer(m, p, 1, 0) }
		}
	}
	return f, false, err
}

func gurdyJrReward(b *Board) (cardEffect, bool) {
	return rewardTreasureHelper(b, 1)
}

func gurdyReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 7)
}

// Boss
// When this is at 2 HP or lower, all Attack Rolls are -1 till the end of turn.
func larryJrEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error = errors.New("larry jr already activated")
	ap := &b.players[b.api]
	if !checkActiveEffects(p.activeEffects, larryJr, false) {
		m := mCard.(*monsterCard)
		if _, err = en.checkDamageToSpecificMonster(larryJr); err == nil && m.hp > 0 && m.hp <= 2 {
			f = func(roll uint8) { ap.activeEffects[larryJr] = struct{}{} }
		}
	}
	return f, false, err
}

func larryJrReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 6)
}

func littleHornReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 2)
}

// Boss
// When this takes damage from an attack, deal 1 damage to the attacking Player
func lustEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = p.checkAttackingPlayer(); err == nil {
		if _, err = en.checkDamageToSpecificMonster(lust); err == nil && en.event.roll > 0 {
			m := mCard.(*monsterCard)
			f = func(roll uint8) { b.damageMonsterToPlayer(m, p, 1, 0) }
		}
	}
	return f, false, err
}

func lustReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 2)
}

// Boss
// When this is at 1 HP, it gains +2 Dice Roll till the end of turn.
func maskOfInfamyEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if _, err = en.checkDamageToSpecificMonster(maskOfInfamy); err == nil {
		if m := mCard.(*monsterCard); m.hp == 1 {
			f = func(roll uint8) { m.roll += 2 }
		}
	}
	return f, false, err
}

func maskOfInfamyReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 2)
}

// Boss
// When this deals damage, it heals 1 HP
func megaFattyEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if _, err = en.checkDamageFromMonster(megaFatty); err == nil {
		m := mCard.(*monsterCard)
		f = func(roll uint8) { m.heal(1) }
	}
	return f, false, err
}

func megaFattyReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 2)
}

func monstroReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 6)
}

// Boss
// When this dies, search the Monster deck for The Bloat.
// Put it into an Active Slot and shuffle the deck.
func thePeepDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if _, _, err = b.monster.deck.search(theBloat); err == nil {
		showMonsterCards(b.monster.getActiveMonsters(), 0)
		fmt.Println("Overlay which zone with The Bloat?")
		i := uint8(readInput(0, len(b.monster.zones)-1))
		f = func(roll uint8) {
			if c, err := b.monster.deck.popById(theBloat); err == nil {
				b.monster.zones[i].push(c.(monsterCard))
				b.monster.deck.shuffle()
			}
		}
	}
	return f, false, err
}

func thePeepReward(b *Board) (cardEffect, bool) {
	return rewardTreasureHelper(b, 1)
}

// Boss
// When this dies, deal 2 damage divided as you choose to any number of players or monsters
func pestilenceDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect
	targets := make([]combatTarget, 0, 2)
	monsters, players := b.monster.getActiveMonsters(), b.getPlayers(true)
	l1, l2 := len(players), len(monsters)
	for len(targets) < 2 {
		max := l1 + l2 - 1
		showPlayers(players, 0)
		showMonsterCards(monsters, l1)
		if len(targets) == 1 {
			max += 1
			fmt.Println(fmt.Sprintf("%d) No additional targets", max))
		}
		ans := readInput(0, max)
		if ans >= 0 && ans < l1 {
			targets = append(targets, players[ans])
		} else {
			targets = append(targets, monsters[ans-l1])
		}
	}
	f = func(roll uint8) {
		var n = uint8(2 / len(targets))
		for _, ct := range targets {
			if m, ok := ct.(*monsterCard); ok {
				b.damagePlayerToMonster(p, m, n, 0)
			} else if p2, ok := ct.(*player); ok {
				b.damagePlayerToPlayer(p, p2, n)
			} else {
				panic("not an anticipated type switch for pestilence")
			}
		}
	}
	return f, false, nil
}

func pestilenceReward(b *Board) (cardEffect, bool) {
	return rewardLootAndCents(b, 2, 3)
}

// Boss
// This takes no damage on a roll of 6.
func pinReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 5)
}

// Boss
// When this is attacked, you must force a player to discard 2 loot cards
func prideEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	var da declareAttackEvent
	if da, err = en.checkDeclareAttack(p); err == nil {
		if da.m.id == pride {
			players := b.getPlayers(false)
			valid := make([]*player, 0, len(players))
			for _, p2 := range players {
				if len(p2.Hand) >= 2 {
					valid = append(valid, p2)
				}
			}
			l := len(valid)
			if l > 0 {
				showPlayers(valid, 0)
				fmt.Println("Choose who should discard 2 loot cards")
				i := uint8(readInput(0, l-1))
				f = func(roll uint8) { valid[i].discardHandChoiceHelper(b.loot, 2) }
			}
		}
	}
	return f, false, nil
}

func prideReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 5)
}

// Boss
// When this dies, roll:
// On a 1 or 6, put this card back on top of the monster deck
func ragmanDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect = func(roll uint8) {
		if roll == 1 || roll == 6 {
			ap := &b.players[b.api]
			for i := range ap.Souls {
				if ap.Souls[i].getId() == ragman {
					b.monster.placeInDeck(ap.popSoul(uint8(i)).(monsterCard), true)
				}
			}
		}
	}
	return f, true, nil
}

func ragmanReward(b *Board) (cardEffect, bool) {
	return rewardLootHelper(b, 3)
}

// Boss
// Each time this deals damage to a player, they also discard a loot card.
func scolexEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if _, err = en.checkDamageFromMonster(scolex); err == nil {
		f = func(roll uint8) { b.players[b.api].discardHandChoiceHelper(b.loot, 1) }
	}
	return f, false, err
}

func scolexReward(b *Board) (cardEffect, bool) {
	return rewardTreasureHelper(b, 1)
}

// Boss
// When this dies, the player that killed it discards all loot cards in their hand
func slothDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	return func(roll uint8) {
		for len(p.Hand) != 0 {
			b.discard(p.popHandCard(0))
		}
	}, false, nil
}

func slothReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 1)
}

// Each time this deals damage, it also deals 1 damage to all other Players.
func theBloatEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if _, err = en.checkDamageFromMonster(theBloat); err == nil && en.event.roll > 0 {
		f = func(roll uint8) {
			m := mCard.(*monsterCard)
			for _, p2 := range b.getOtherPlayers(p, true) {
				b.damageMonsterToPlayer(m, p2, 1, 0)
			}
		}
	}
	return f, false, err
}

func theBloatReward(b *Board) (cardEffect, bool) {
	return rewardTreasureHelper(b, 1)
}

// Boss
// When this takes damage roll:
// 1: Prevent that damage
// 2-6: Nothing.
func theDukeOfFliesEvent(en *eventNode) cardEffect {
	return func(roll uint8) {
		if roll == 1 {
			en.event.e = fizzledEvent{}
		}
	}
}

func theDukeOfFliesReward(b *Board) (cardEffect, bool) {
	return rewardLootAndCents(b, 2, 3)
}

// Boss
// When this takes 2 damage, the attacking player's dice rolls are -1 till the end of the attack
func theHauntEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	var d damageEvent
	if d, err = en.checkDamageToSpecificMonster(theHaunt); err == nil && d.n == 2 {
		ap := &b.players[b.api]
		f = func(roll uint8) { ap.activeEffects[theHaunt] = struct{}{} }
	}
	return f, false, nil
}

func theHauntReward(b *Board) (cardEffect, bool) {
	return rewardTreasureHelper(b, 1)
}

// bBoss
// Whenever this takes damage it gains +1 AP
func warEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if _, err = en.checkDamageToSpecificMonster(war); err == nil {
		m := mCard.(*monsterCard)
		f = func(roll uint8) { m.increaseAP(1) }
	}
	return f, false, err
}

func warReward(b *Board) (cardEffect, bool) {
	return rewardLootAndCents(b, 2, 3)
}

// Boss
// When this dies, roll:
// 1-3: All players take 1 damage
// 4-6: All players take 2 damage
func wrathDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	return func(roll uint8) {
		var n uint8 = 1
		if roll >= 4 {
			n = 2
		}
		for _, p2 := range b.getPlayers(true) {
			b.damageMonsterToPlayer(mCard.(*monsterCard), p2, n, 0)
		}
	}, true, nil
}

func wrathReward(b *Board) (cardEffect, bool) {
	return rewardLootAndCents(b, 2, 3)
}

// Mega Boss
// This deals double damage on a roll of 1
// When this dies, expand the number of active monsters by 1
//
// First effect handled at the damage calculation step
func momChecker(mId uint16, roll uint8) bool {
	return mId == mom && roll == 1
}

func momDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	return func(roll uint8) { b.monster.zones = append(b.monster.zones, activeSlot{}) }, false, nil
}

func momReward(b *Board) (cardEffect, bool) {
	return rewardMegaBoss(b)
}

// Mega Boss
// When the attacking player rolls a 6, they must kill a player of their choosing
func satanEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = p.checkAttackingPlayer(); err == nil {
		if err = en.checkDiceRoll(6); err == nil {
			players := b.getPlayers(true)
			showPlayers(players, 0)
			fmt.Println("Who to kill?")
			i := readInput(0, len(players)-1)
			f = func(roll uint8) { b.killPlayer(players[i]) }
		}
	}
	return f, false, nil
}

func satanReward(b *Board) (cardEffect, bool) {
	return rewardMegaBoss(b)
}

// Mega boss
// When this dies, you may force a player to give you a soul
func theLambDeath(p *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	valid := make([]*player, 0, len(b.players)-1)
	for _, player := range b.getOtherPlayers(p, false) {
		if len(player.Souls) > 0 {
			valid = append(valid, player)
		}
	}
	l := len(valid)
	if l > 0 {
		showPlayers(valid, 0)
		fmt.Println("Steal a soul from whom?")
		target := valid[readInput(0, l-1)]
		showSouls(target.Souls, target.Character.name, 0)
		fmt.Println("Which soul to steal?")
		soulId := target.Souls[uint8(readInput(0, len(target.Souls)-1))].getId()
		f = func(roll uint8) {
			if i, err := target.getSoulIndex(soulId); err == nil {
				p.addSoulToBoard(target.popSoul(i))
			}
		}
	} else {
		err = errors.New("no candidate with souls available")
	}
	return f, false, err
}

func theLambReward(b *Board) (cardEffect, bool) {
	return rewardCentsHelper(b, 3)
}

// Bonus card
// You must attack the monster deck 2 times this turn
func ambushFunc(ap *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect = func(roll uint8) {
		if ap.numAttacks == 1 && !ap.inBattle {
			ap.numAttacks = 2
		} else {
			ap.numAttacks += 2
		}
		ap.forceAttackTarget[forceAttackDeck] = 2
	}
	return f, false, nil
}

// Bonus card
// Roll:
// 1-2: Gain 1 cent
// 3-4: Gain 3 cents
// 5-6: Gain 6 cents
func chestFunc(ap *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect = func(roll uint8) {
		if roll == 1 || roll == 2 {
			ap.gainCents(1)
		} else if roll == 3 || roll == 4 {
			ap.gainCents(3)
		} else {
			ap.gainCents(6)
		}
	}
	return f, true, nil
}

// Bonus card
// Roll:
// 1-3: Take 1 damage
// 4-5: Take 2 damage
// 6: Reveal cards from the top of the Treasure deck until you reveal a Guppy item.
// Gain it and shuffle all revealed cards into the deck.
func cursedChestFunc(ap *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect = func(roll uint8) {
		if roll >= 1 && roll <= 3 {
			b.damagePlayerToPlayer(ap, ap, 1)
		} else if roll == 4 || roll == 5 {
			b.damagePlayerToPlayer(ap, ap, 1)
		} else {
			validGuppyItems := getValidGuppyItems()
			revealedCards := make(deck, 0, b.treasure.deck.len())
			tc := treasureCard{}
			for b.treasure.deck.len() > 0 {
				tc = b.treasure.draw()
				if _, ok := validGuppyItems[tc.id]; !ok {
					revealedCards = append(revealedCards, tc)
				} else {
					break
				}
			}
			if _, ok := validGuppyItems[tc.id]; ok {
				fmt.Println("Found Guppy Item")
				tc.showCard(0)
				ap.addCardToBoard(tc)
			}
			showDeck(revealedCards, false)
			b.treasure.deck.merge(revealedCards, true, true)
		}
	}
	return f, true, nil
}

// Bonus card
// Roll: 1-2 Loot 1. 3-4: Gain 3 cents. 5-6: take 2 damage
func darkChestFunc(ap *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect = func(roll uint8) {
		if roll == 1 || roll == 2 {
			ap.loot(b.loot)
		} else if roll == 3 || roll == 4 {
			ap.gainCents(3)
		} else {
			b.damagePlayerToPlayer(ap, ap, 2)
		}
	}
	return f, true, nil
}

// Devil Deal
// Choose one: 1: Discard this. 2: Draw 2, take 1 damage. 3: Search the treasure deck for a guppy
// item. Gain it and take 2 damage. Shuffle the deck
func devilDealFunc(ap *player, b *Board, mCard card) (cardEffect, bool, error) {
	fmt.Println("Choose 1:\n1) Discard this.\n2) Draw 2, take 1 damage\n3) Search the Treasure deck for a Guppy" +
		"item, gain it and take 2 damage. Shuffle the deck.")
	ans := readInput(1, 3)
	var f cardEffect = func(roll uint8) {
		if ans == 2 {
			ap.loot(b.loot)
			ap.loot(b.loot)
			b.damagePlayerToPlayer(ap, ap, 1)
		} else if ans == 3 {
			guppyCards, idIndexMap := b.treasure.deck.scan([]uint16{
				guppysCollar, guppysEye, guppysHead, guppysTail, guppysPaw, theDeadCat,
			}...)
			l := len(guppyCards)
			if l > 0 {
				showTreasureCards(guppyCards, "deck", 0)
				fmt.Println("Which Guppy item to gain?")
				card := guppyCards[readInput(0, l-1)]
				if c, err := b.treasure.deck.popByIndex(idIndexMap[card.getId()]); err == nil {
					ap.addCardToBoard(c)
				}
			}
		}
	}
	return f, false, nil
}

// Bonus card
// Roll: 1-2: +1 Treasure. 3-4: Gain 5 cents. 5-6: Gain 7 cents.
func goldChestFunc(ap *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect = func(roll uint8) {
		if roll == 1 || roll == 2 {
			ap.addCardToBoard(b.treasure.draw())
		} else if roll == 3 || roll == 4 {
			ap.gainCents(5)
		} else {
			ap.gainCents(7)
		}
	}
	return f, true, nil
}

// Bonus card
// Choose the Player with the most cents or that is tied for the most,
// that player loses all their cents
func greedBonusFunc(ap *player, b *Board, mCard card) (cardEffect, bool, error) {
	players := b.getOtherPlayers(ap, false)
	conflict := []*player{players[0]}
	for i := 1; i < len(players); i++ {
		max := conflict[0]
		if players[i].Pennies > max.Pennies {
			conflict = append(conflict[:0], players[i])
		} else if players[i].Pennies == max.Pennies {
			conflict = append(conflict, players[i])
		}
	}
	var i uint8 = 0
	l := len(conflict)
	if l > 1 {
		showPlayers(conflict, 0)
		fmt.Println("Which players should lose all cents?")
		i = uint8(readInput(0, l-1))
	}
	var f cardEffect = func(roll uint8) { conflict[i].loseCents(100) }
	return f, false, nil
}

// Bonus Card
// Look at the top 6 cards of the loot deck. You may put them back in any order, then loot 1
func iCanSeeForeverFunc(ap *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect = func(roll uint8) {
		cards := make([]lootCard, 6)
		for i := 0; i < 6; i++ {
			cards[i] = b.loot.draw()
		}
		for len(cards) > 0 {
			showLootCards(cards, "peek", 0)
			fmt.Println("Place which card on top of the deck?")
			i := uint8(readInput(0, len(cards)-1))
			b.loot.placeInDeck(cards[i], true)
			cards = append(cards[:i], cards[i+1:]...)
		}
		ap.loot(b.loot)
	}
	return f, false, nil
}

// Bonus Card
// All players take 2 damage!
func megaTrollBombFunc(ap *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect = func(roll uint8) {
		for _, p := range b.getPlayers(true) {
			b.damagePlayerToPlayer(ap, p, 1)
		}
	}
	return f, false, nil
}

// bonus Card
// You take 2 damage
func trollBombsFunc(ap *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect = func(roll uint8) {
		b.damagePlayerToPlayer(ap, ap, 2)
	}
	return f, false, nil
}

// Bonus Card
// Roll: 1: Take 3 damage. 2-3: Discard 2 loot. 4-5: Gain 7 cents. 6: Gain +1 Treasure
func secretRoomFunc(ap *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect = func(roll uint8) {
		if roll == 1 {
			b.damagePlayerToPlayer(ap, ap, 3)
		} else if roll == 2 || roll == 3 {
			ap.discardHandChoiceHelper(b.loot, 2)
		} else if roll == 4 || roll == 5 {
			ap.gainCents(7)
		} else {
			ap.addCardToBoard(b.treasure.draw())
		}
	}
	return f, true, nil
}

// Bonus Card
// Expand the number of items in the shop by 2.
// You may attack an additional time this turn.
func shopUpgradeFunc(ap *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect = func(roll uint8) {
		placeHolder := treasureCard{}
		b.treasure.zones = append(b.treasure.zones, placeHolder, placeHolder)
		ap.numAttacks += 1
	}
	return f, false, nil
}

// Bonus card
// Put any number of discarded monsters back on top of the Monster deck
// You may attack an additional time this turn.
func weNeedToGoDeeperFunc(ap *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect = func(roll uint8) {
		for len(b.monster.discardPile) > 0 {
			l := b.monster.discardPile.len()
			showDeck(b.monster.discardPile, false)
			fmt.Println(fmt.Sprintf("%d) Stop putting discarded monsters on top of the deck", l))
			ans := uint8(readInput(0, int(l)))
			if ans < l {
				c, _ := b.monster.discardPile.popByIndex(ans)
				b.monster.placeInDeck(c.(monsterCard), true)
			} else {
				break
			}
		}
		ap.numAttacks += 1
	}
	return f, false, nil
}

// Bonus card
// Expand the number of active monsters by 1.
// You may attack an additional time this turn.
func xlFloorFunc(ap *player, b *Board, mCard card) (cardEffect, bool, error) {
	var f cardEffect = func(roll uint8) {
		b.monster.zones = append(b.monster.zones, activeSlot{})
		ap.numAttacks += 1
	}
	return f, false, nil
}

// Curse
// When revealed, give this curse to any player (handled elsewhere)
// At the end of your turn, discard 2 loot.
// When you die, discard this. (handled elsewhere)
func curseOfAmnesiaEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkEndOfTurn(p); err == nil {
		f = func(roll uint8) { p.discardHandChoiceHelper(b.loot, 2) }
	}
	return f, false, err
}

// Curse
// When revealed, give this curse to any player (handled elsewhere)
// At the end of your turn, lose 4 cents
// When you die, discard this. (handled elsewhere)
func curseOfGreedEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkEndOfTurn(p); err == nil {
		f = func(roll uint8) { p.loseCents(4) }
	}
	return f, false, err
}

// Curse
// When revealed, give this curse to any player (handled elsewhere)
// You now need 5 souls to win.
// When you die, discard this. (handled elsewhere)
func curseOfLossChecker(p player) bool {
	var found bool
	for _, c := range p.Curses {
		if c.id == curseOfLoss {
			found = true
			break
		}
	}
	return found
}

// Curse
// When revealed, give this curse to any player (handled elsewhere)
// At the start of your turn, take 1 damage.
// When you die, discard this. (handled elsewhere)
func curseOfPainEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	if err = en.checkStartOfTurn(p); err == nil {
		f = func(roll uint8) { b.damagePlayerToPlayer(p, p, 1) }
	}
	return f, false, err
}

// Curse
// When revealed, give this curse to any player (handled elsewhere)
// All monsters you attack gain +1 Dice Roll
// When you die, discard this. (handled elsewhere)
func curseOfTheBlindEvent(p *player, b *Board, mCard card, en *eventNode) (cardEffect, bool, error) {
	var f cardEffect
	var err error
	var intention intentionToAttackEvent
	if intention, err = en.checkIntentionToAttack(); err == nil && p.getId() == en.event.p.getId() {
		f = func(roll uint8) { intention.m.modifyDiceRoll(1) }
	}
	return f, false, err
}
