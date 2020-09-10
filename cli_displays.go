package four_souls

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func showCharacterCards(players []*player, offset int) {
	var s = "Player Characters\n"
	s += characterCard{}.header()
	for i, p := range players {
		s += p.showCard(i + offset)
	}
	writeToStdout(s)
}

func showDeck(cards deck, reverse bool) {
	var s = "Some collection of cards\n"
	if card, err := cards.peek(); err == nil {
		s += card.header()
		if !reverse {
			for i, c := range cards {
				s += c.showCard(i)
			}
		} else {
			l := len(cards)
			for i := l - 1; i >= 0; i-- {
				s += cards[i].showCard(l - i - 1)
			}
		}
	}
}

func showEvents(events []*eventNode) {
	var s = "Events (in resolve order).\n"
	s += headerEventStack()
	for i, e := range events {
		s += e.showEvent(i)
	}
	writeToStdout(s)
}

func showItems(cards []itemCard, offset int) {
	var s = fmt.Sprintf("Items\n%s", treasureCard{}.header())
	for i, c := range cards {
		s += c.showCard(i + offset)
	}
	writeToStdout(s)
}

func showLootCards(lc interface{}, owner string, offset int) {
	var s = fmt.Sprintf("Loot Cards owned by %s\n", owner)
	s += lootCard{}.header()
	switch lc.(type) {
	case []*lootCard:
		cards := lc.([]*lootCard)
		for i, l := range cards {
			s += l.showCard(i + offset)
		}
	case []lootCard:
		cards := lc.([]lootCard)
		for i, l := range cards {
			s += l.showCard(i + offset)
		}
	default:
		panic("not a loot value slice")
	}
	writeToStdout(s)
}

func showMonsterCards(monsters interface{}, offset int) {
	var s = "Monsters, Curses, or Bonuses\n"
	s += monsterCard{}.header()
	switch monsters.(type) {
	case []*monsterCard:
		cards := monsters.([]*monsterCard)
		for i, m := range cards {
			s += m.showCard(i + offset)
		}
	case []monsterCard:
		cards := monsters.([]monsterCard)
		for i, m := range cards {
			s += m.showCard(i + offset)
		}
	default:
		panic("not a monster value.")
	}

	writeToStdout(s)
}

func showPlayers(players interface{}, offset int) {
	var s = fmt.Sprintf("Players\n%s", player{}.header())
	switch players.(type) {
	case []player:
		ps := players.([]player)
		for i, p := range ps {
			s += fmt.Sprintf("\t%d\t%s\t%d\t%v\t%v\t%v\t\n",
				i+offset, p.Character.name, p.Pennies, p.Souls, p.ActiveItems, p.PassiveItems)
		}
	case []*player:
		ps := players.([]*player)
		for i, p := range ps {
			s += fmt.Sprintf("\t%d\t%s\t%d\t%v\t%v\t%v\t\n",
				i+offset, p.Character.name, p.Pennies, p.Souls, p.ActiveItems, p.PassiveItems)
		}
	default:
		panic("not a players type")
	}
	writeToStdout(s)
}

func showSouls(souls []card, owner string, offset int) {
	var s = fmt.Sprintf("Souls for %s\n\tIndex\tName\n", owner)
	for i := range souls {
		s += fmt.Sprintf("\t%d\t%s", i+offset, souls[i].getName())
	}
	writeToStdout(s)
}

func showSoulsByPlayer(souls []card, playerMap map[uint16]*player, offset int) {
	var s string
	for i := range souls {
		s += fmt.Sprintf("%s owned by %s\n", souls[i].getName(), playerMap[souls[i].getId()].Character.name)
	}
	writeToStdout(s)
}

func showTreasureCards(items interface{}, owner string, offset int) {
	var s = fmt.Sprintf("active Items for %s\n", owner)
	s += treasureCard{}.header()
	switch items.(type) {
	case []*treasureCard:
		cards := items.([]*treasureCard)
		for i, t := range cards {
			s += t.showCard(i + offset)
		}
	case []treasureCard:
		cards := items.([]treasureCard)
		for i, t := range cards {
			s += t.showCard(i + offset)
		}
	default:
		panic("not an itemCard value collection.")
	}
	writeToStdout(s)
}

func (cc characterCard) header() string {
	return "\tIndex\tName\thp\tap\tSoul Hearts\ttapped\n"
}

func headerEventStack() string {
	return "\tIndex\tName\tEvent Type\tTargeting\n"
}

func (lc lootCard) header() string {
	return fmt.Sprintf("\tIndex\tName\tTrinket\n")
}

func (mc monsterCard) header() string {
	var s = fmt.Sprintf("\tIndex\tName")
	if mc.baseHealth > 0 {
		s = fmt.Sprintf("\t%s\t%s\t%s\t%s\n", "hp", "ap", "roll", "Reward")
	} else {
		s += "\n"
	}
	return s
}

func (tc treasureCard) header() string {
	return "\tIndex\tName\tEternal\tActive\tPaid\tPassive\tTriggered\tCounters\n"
}

func (p player) header() string {
	return "Players\n\tIndex\tCharacter\tHP\tAP\tPennies\tSouls\tActive Items\tPassive Items\n"

}

func (p player) showCard(idx int) string {
	return fmt.Sprintf("\t%d\t%s\t%d\t%d\t%d\t%v\t%v\t%v",
		idx, p.Character.name, p.Character.hp, p.Character.ap, p.Pennies, p.Souls, p.ActiveItems, p.PassiveItems)
}

func (cc characterCard) showCard(idx int) string {
	return fmt.Sprintf("\t%d\t%s\t%d\t%d\t%t", idx, cc.name, cc.hp, cc.ap, cc.tapped)
}

func (lc lootCard) showCard(idx int) string {
	var s = fmt.Sprintf("\t%d\t%s", idx, lc.name)
	if lc.trinket {
		s += fmt.Sprintf("\t%t\t%t\t%t\t%t", lc.eternal, false, false, true)
	}
	s += "\n"
	return fmt.Sprintf("\t%d\t%s\t%t\n", idx, lc.name, lc.trinket)
}

func (mc monsterCard) showCard(idx int) string {
	var s = fmt.Sprintf("\t%d\t%s", idx, mc.name)
	if mc.baseHealth > 0 {
		s += fmt.Sprintf("\t%d\t%d\t%d\n", mc.hp, mc.ap, mc.roll)
	} else {
		s += "\n"
	}
	return s
}

func (tc treasureCard) showCard(idx int) string {
	return fmt.Sprintf("\tc%d\tc%s\tc%tc\tc%tc\tc%tc\tc%tc\tc%tc\tc%d\n", idx, tc.name, tc.eternal, tc.active, tc.paid,
		tc.passive, tc.tapped, tc.counters)
}

func (en eventNode) showEvent(idx int) string {
	var name, eventType string
	value := en.event.e
	switch value.(type) {
	case activateEvent:
		name, eventType = value.(activateEvent).c.getName(), "Activated effect"
	case damageEvent:
		name, eventType = value.(damageEvent).target.getName(), "Damage"
	case deathOfCharacterEvent:
		name, eventType = en.event.p.Character.name, "Dead Character"
	case declareAttackEvent:
		name, eventType = value.(declareAttackEvent).m.name, "Attack Monster"
	case declarePurchaseEvent:
		eventType = "Buy Item"
	case intentionToAttackEvent:
		eventType = "Starting Attack Phase"
	case intentionToPurchaseEvent:
		eventType = "Starting Purchase"
	case lootCardEvent:
		name, eventType = value.(lootCardEvent).l.name, "Hand Activation"
	case paidItemEvent:
		name, eventType = value.(paidItemEvent).t.name, "paid Item tapped"
	case diceRollEvent:
		eventType = fmt.Sprintf("Rolled %d", value.(diceRollEvent).n)
	case triggeredEffectEvent:
		name, eventType = value.(triggeredEffectEvent).c.getName(), "tapped effect"
	}
	return fmt.Sprintf("\t%d\t%s\t%s\n", idx, name, eventType)
}

// Reads the input from the command line and returns the player's choice.
// decStatements: a message where each sentence is prepende with some auto-incremented positive id, starting from 1.
// max: the user's input must be a number that falls in the range of [0, max)
func readInput(min int, max int) int {
	var choice int = -1
	for choice < 0 {
		var val int
		fmt.Print("Enter id number -> ")
		_, err := fmt.Scanf("%d", &val)
		if err != nil {
			panic(err)
		}
		if val < min || val > max {
			fmt.Println("Not a valid target.")
		} else {
			choice = val
		}
	}
	return choice
}

func writeToStdout(s string) {
	w := new(tabwriter.Writer)
	defer w.Flush()
	w.Init(os.Stdout, 8, 8, 1, '\t', 0)
	_, _ = fmt.Fprintf(w, s)
}
