package four_souls

import (
	"errors"
	"fmt"
)

// Get all of the events where a character card is activated.
func (es eventStack) getActivateCharacterEvents() []*eventNode {
	var nodes = make([]*eventNode, 0, es.size)
	if es.head != nil {
		curr := es.head.top
		for curr != nil {
			if _, ok := curr.event.e.(activateEvent).c.(*characterCard); ok {
				nodes = append(nodes, curr)
			}
			curr = curr.next
		}
	}
	return nodes
}

// Get all active items that a player controls.
// getEternal bool: If true, add eternal items to the list. Else, do not.
// Note active items are cards where active == true.
// Paid items will be excluded from collection.
func (p player) getActiveItems(getEternal bool) []*treasureCard {
	var cards = make([]*treasureCard, 0, len(p.ActiveItems))
	for i := range p.ActiveItems {
		c := &p.ActiveItems[i]
		if getEternal || (!getEternal && !c.eternal) {
			cards = append(cards, c)
		}
	}
	return cards
}

// Get all activated active itemCard events in the event stack.
func (es eventStack) getActivateItemEvents() []*eventNode {
	var nodes = make([]*eventNode, 0, es.size)
	if es.head != nil {
		curr := es.head.top
		for curr != nil {
			if _, ok := curr.event.e.(activateEvent).c.(*treasureCard); ok {
				nodes = append(nodes, curr)
			}
			curr = curr.next
		}
	}
	return nodes
}

// Get the index of a monster whose id matches the id param.
// Required for cards like Bomb! that require the monster to
// be on the field for it to activate.
// id uint16: the monster card's id
func (m mArea) getActiveMonster(id uint16) (uint8, *monsterCard) {
	var i uint8
	var c *monsterCard
	for i = range m.zones {
		c = m.zones[i].peek()
		if c.id == id {
			break
		}
	}
	return i, c
}

// Get a pointer to all active monsters
func (m mArea) getActiveMonsters() []*monsterCard {
	monsters := make([]*monsterCard, len(m.zones))
	for i := range m.zones {
		monsters[i] = m.zones[i].peek()
	}
	return monsters
}

func (m mArea) getActiveMonsterInBattle() *monsterCard {
	var monster *monsterCard
	for i := range m.zones {
		c := m.zones[i].peek()
		if c.inBattle {
			monster = c
			break
		}
	}
	return monster
}

// Return the active player
func (b Board) getActivePlayer() *player {
	return &b.players[b.api]
}

// Get all items that are in play by every player.
// getEternal bool: If true, get eternal items as well. Else, do not.
// excludePlayer *player: exclude getting items from this player (always the player who invoked this function).
// Returns: a slice of all the cards and a map where key = index in slice; value = player owner
func (b *Board) getAllItems(getEternal bool, excludePlayer *player) ([]itemCard, map[uint16]*player) {
	items := make([]itemCard, 0)
	itemOwners := make(map[uint16]*player)
	for _, p := range b.getPlayers(false) {
		if p != excludePlayer {
			playerItems := p.getAllItems(getEternal)
			for _, item := range playerItems {
				items = append(items, item)
				itemOwners[item.getId()] = p
			}
		}
	}
	return items, itemOwners
}

// Get all items in play by the player
// getEternal bool: If true, get eternal items as well. Else, do not.
func (p player) getAllItems(getEternal bool) []itemCard {
	l := len(p.ActiveItems)
	var items = make([]itemCard, l+len(p.PassiveItems))
	for i, c := range p.ActiveItems {
		if (!c.eternal && !getEternal) || getEternal {
			items[i] = &c
		}
	}
	for i, c := range p.PassiveItems {
		if (!c.isEternal() && !getEternal) || getEternal {
			items[i+l] = c
		}
	}
	return items
}

func (b Board) getAllPassiveItems(getEternal bool) []*passiveItem {
	items := make([]*passiveItem, 0)
	for _, p := range b.getPlayers(false) {
		items = append(items, p.getPassiveItems(getEternal)...)
	}
	return items
}

func (es eventStack) getAttackDiceRollEvents() []*eventNode {
	var nodes = make([]*eventNode, 0, es.size)
	if es.head != nil {
		curr := es.head.top
		for curr != nil {
			if _, ok := curr.event.e.(diceRollEvent); ok {
				if _, ok := curr.next.event.e.(declareAttackEvent); ok {
					nodes = append(nodes, curr)
				}
			}
			curr = curr.next
		}
	}
	return nodes
}

// Get all player characters
// filterDead bool: If true, only get characters that are not dead.
func (b *Board) getCharacters(filterDead bool) []*characterCard {
	var characters []*characterCard
	for i := range b.players {
		if filterDead {
			if b.players[i].Character.hp > 0 {
				characters = append(characters, &b.players[i].Character)
			}
		} else {
			characters = append(characters, &b.players[i].Character)
		}
	}
	return characters
}

// Get the index of a curse the player has.
// id uint16: The curse id of the curse card.
func (p player) getCurseIndex(id uint16) (uint8, error) {
	var j uint8
	for j = 0; j < uint8(len(p.Curses)); j++ {
		if p.Curses[j].id == id {
			return j, nil
		}
	}
	return j, fmt.Errorf("cannot find curse with id %d", id)
}

// Get all damage events on the event stack.
func (es eventStack) getDamageEvents() []*eventNode {
	var nodes = make([]*eventNode, 0, es.size)
	if es.head != nil {
		curr := es.head.top
		for curr != nil {
			if _, ok := curr.event.e.(damageEvent); ok {
				nodes = append(nodes, curr)
			}
			curr = curr.next
		}
	}
	return nodes
}

func (es eventStack) getDamageEventsGT1() []*eventNode {
	var nodes = make([]*eventNode, 0, es.size)
	if es.head != nil {
		curr := es.head.top
		for curr != nil {
			if e, ok := curr.event.e.(damageEvent); ok {
				if e.n > 1 {
					nodes = append(nodes, curr)
				}
			}
			curr = curr.next
		}
	}
	return nodes
}

// Get only the damage events that only involve players / characters.
func (es eventStack) getDamageOfCharacterEvents() []*eventNode {
	nodes := es.getDamageEvents()
	valid := make([]*eventNode, 0, len(nodes))
	for i := range nodes {
		e := nodes[i].event.e
		if _, ok := e.(damageEvent).target.(*player); ok {
			valid = append(valid, nodes[i])
		}
	}
	return valid
}

// Get only the damage events that only involves monsters.
func (es eventStack) getDamageOfMonsterEvents() []*eventNode {
	nodes := es.getDamageEvents()
	valid := make([]*eventNode, 0, len(nodes))
	for i := range nodes {
		if _, ok := nodes[i].event.e.(damageEvent).target.(*monsterCard); ok {
			valid = append(valid, nodes[i])
		}
	}
	return valid
}

func (es eventStack) getDeathOfCharacterEvents() []*eventNode {
	var nodes = make([]*eventNode, 0, es.size)
	if es.head != nil {
		curr := es.head.top
		for curr != nil {
			if _, ok := curr.event.e.(deathOfCharacterEvent); ok {
				nodes = append(nodes, curr)
			}
			curr = curr.next
		}
	}
	return nodes
}

func (es eventStack) getDeclareAttackEvents() []*eventNode {
	var nodes = make([]*eventNode, 0, es.size)
	if es.head != nil {
		curr := es.head.top
		for curr != nil {
			if _, ok := curr.event.e.(declareAttackEvent); ok {
				nodes = append(nodes, curr)
			}
			curr = curr.next
		}
	}
	return nodes
}

func (es eventStack) getDeclarePurchaseEvents() []*eventNode {
	var nodes = make([]*eventNode, 0, es.size)
	if es.head != nil {
		curr := es.head.top
		for curr != nil {
			if _, ok := curr.event.e.(declarePurchaseEvent); ok {
				nodes = append(nodes, curr)
			}
			curr = curr.next
		}
	}
	return nodes
}

func (es eventStack) getDiceRollEvents() []*eventNode {
	var nodes = make([]*eventNode, 0, es.size)
	if es.head != nil {
		curr := es.head.top
		for curr != nil {
			if _, ok := curr.event.e.(diceRollEvent); ok {
				nodes = append(nodes, curr)
			}
			curr = curr.next
		}
	}
	return nodes
}

func (p player) getHandCardIndexById(lId uint16) (uint8, error) {
	err := errors.New("card not found")
	var i uint8
	for i = 0; i < uint8(len(p.Hand)); i++ {
		if p.Hand[i].id == lId {
			err = nil
			break
		}
	}
	return i, err
}

func (es eventStack) getIntentionToAttackEvents() []*eventNode {
	var nodes = make([]*eventNode, 0, es.size)
	if es.head != nil {
		curr := es.head.top
		for curr != nil {
			if _, ok := curr.event.e.(intentionToAttackEvent); ok {
				nodes = append(nodes, curr)
			}
			curr = curr.next
		}
	}
	return nodes
}

func (es eventStack) getIntentionToPurchaseEvents() []*eventNode {
	var nodes = make([]*eventNode, 0, es.size)
	if es.head != nil {
		curr := es.head.top
		for curr != nil {
			if _, ok := curr.event.e.(intentionToPurchaseEvent); ok {
				nodes = append(nodes, curr)
			}
			curr = curr.next
		}
	}
	return nodes
}

func (b Board) getItemIndex(itemId uint16, isPassive bool) (uint8, *player) {
	var j uint8
	var p *player
	var err = errors.New("item not found")
	players := b.getPlayers(false)
	for _, player := range players {
		j, err = player.getItemIndex(itemId, isPassive)
		if err == nil {
			p = player
			break
		}
	}
	return j, p
}

func (p player) getItemIndex(itemId uint16, isPassive bool) (uint8, error) {
	var e = errors.New("item not found")
	var median, low uint8
	if !isPassive {
		high := uint8(len(p.ActiveItems))
		for low <= high {
			median = (low + high) / 2
			id := p.ActiveItems[median].id
			if id < itemId {
				low = median + 1
			} else if id > itemId {
				high = median - 1
			} else {
				e = nil
				break
			}
		}
	} else {
		high := uint8(len(p.PassiveItems))
		for low <= high {
			median = (low + high) / 2
			id := p.PassiveItems[median].getId()
			if id < itemId {
				low = median + 1
			} else if id > itemId {
				high = median - 1
			} else {
				e = nil
				break
			}
		}
	}
	return median, e
	//var i uint8
	//if !isPassive {
	//	var target treasureCard
	//	for i, target = range p.ActiveItems {
	//		if target.id == itemId {
	//			e = nil
	//			break
	//		}
	//	}
	//} else {
	//	var target passiveItem
	//	for i, target = range p.PassiveItems {
	//		if target.getId() == itemId {
	//			e = nil
	//			break
	//		}
	//	}
	//}
	//return i, e
}

func (es eventStack) getLootCardEvents() []*eventNode {
	var nodes = make([]*eventNode, 0, es.size)
	if es.head != nil {
		curr := es.head.top
		for curr != nil {
			if _, ok := curr.event.e.(lootCardEvent); ok {
				nodes = append(nodes, curr)
			}
			curr = curr.next
		}
	}
	return nodes
}

// Get the next player in relation to the player that called this method
func (b Board) getNextPlayer(p *player) *player {
	var i int
	for i = range b.players {
		if b.players[i].Character.id == p.Character.id {
			break
		}
	}
	next := (i + 1) % len(b.players)
	return &b.players[next]
}

func (es eventStack) getNonAttackDiceRollEvents() []*eventNode {
	var nodes = make([]*eventNode, 0, es.size)
	if es.head != nil {
		curr := es.head.top
		for curr != nil {
			if _, ok := curr.event.e.(diceRollEvent); ok {
				if _, ok := curr.next.event.e.(declareAttackEvent); !ok {
					nodes = append(nodes, curr)
				}
			}
			curr = curr.next
		}
	}
	return nodes
}

func (b *Board) getOtherPlayers(excludePlayer *player, filterDead bool) []*player {
	var l = uint8(len(b.players))
	var players = make([]*player, l-1)
	var i = b.api
	var j uint8
	for j = 0; j < l-1; j++ {
		if b.players[i].Character.id != excludePlayer.Character.id {
			if (filterDead && b.players[i].Character.hp > 0) || !filterDead {
				players[i] = &b.players[i]
			}
		}
		i = (i + 1) % l
	}
	return players
}

func (p player) getPaidItems() []*treasureCard {
	cards := make([]*treasureCard, 0, len(p.ActiveItems))
	for i := range p.ActiveItems {
		if p.ActiveItems[i].paid {
			cards = append(cards, &p.ActiveItems[i])
		}
	}
	return cards
}

func (es eventStack) getPaidItemEvents() []*eventNode {
	var nodes = make([]*eventNode, 0, es.size)
	if es.head != nil {
		curr := es.head.top
		for curr != nil {
			if _, ok := curr.event.e.(paidItemEvent); ok {
				nodes = append(nodes, curr)
			}
			curr = curr.next
		}
	}
	return nodes
}

func (p player) getPassiveItems(getEternal bool) []*passiveItem {
	var cards = make([]*passiveItem, 0, len(p.PassiveItems))
	for i := range p.PassiveItems {
		if getEternal || (!getEternal && !p.PassiveItems[i].isEternal()) {
			c := &p.PassiveItems[i]
			cards = append(cards, c)
		}
	}
	return cards
}

func (p player) getPlayerActions(isActivePlayer bool, emptyEs bool) []actionReaction {
	actions := make([]actionReaction, 0, 6)
	if isActivePlayer {
		if p.numLootPlayed > 0 {
			actions = append(actions, actionReaction{msg: "Play a Loot Card from your hand", value: playLootCard})
		}
		var shopCost int8 = 10
		if _, err := p.getItemIndex(steamySale, true); err == nil {
			shopCost = 5
		}
		if _, ok := p.activeEffects[creditCard]; ok {
			shopCost = 0
		}
		if p.Pennies > shopCost && emptyEs {
			actions = append(actions, actionReaction{msg: "Buy an Item from the Shop", value: buyItem})
		}
		if p.numAttacks > 0 && emptyEs {
			actions = append(actions, actionReaction{msg: "Attack!", value: attackMonster})
		}
	}
	if !p.Character.tapped {
		m := fmt.Sprintf("Activate Character Card (%s)", p.Character.name)
		actions = append(actions, actionReaction{msg: m, value: activateCharacter})
	}
	if len(p.getUsableActiveItems()) > 0 { // TODO check conditions for active items to activate
		actions = append(actions, actionReaction{msg: "Activate an Item", value: activateItem})
	}
	if _, err := p.getItemIndex(theresOptions, true); err == nil {
		actions = append(actions, actionReaction{msg: "Peek at the Treasure deck", value: peekTheresOptions})
	}
	if isActivePlayer && !p.inBattle {
		actions = append(actions, actionReaction{msg: "End your turn", value: endActivePlayerTurn})
	} else if !isActivePlayer {
		actions = append(actions, actionReaction{msg: "Do nothing", value: doNothing})
	}
	return actions
}

// Get the player with the matching character.
func (b *Board) getPlayerFromCharacterId(id uint16) (*player, error) {
	var player *player
	var err = errors.New("no character of this id")
	for _, p := range b.players {
		if p.Character.id == id {
			player, err = &p, nil
			break
		}
	}
	return player, err
}

// Get the players, in turn order. With the 0th
// element in the list being the current active player.
func (b *Board) getPlayers(filterDead bool) []*player {
	var l = uint8(len(b.players))
	var players = make([]*player, l)
	var i = b.api
	var j uint8
	for j = 0; j < l; j++ {
		if (filterDead && b.players[i].Character.hp > 0) || !filterDead {
			players[i] = &b.players[i]
		}
		i = (i + 1) % l
	}
	return players
}

func (b Board) getPreviousPlayer(cId uint16) *player {
	var i int
	for i = range b.players {
		if b.players[i].Character.id == cId {
			break
		}
	}
	next := (i - 1) % len(b.players)
	return &b.players[next]
}

func (b *Board) getSouls() ([]card, map[uint16]*player) {
	var souls []card
	playerMap := make(map[uint16]*player)
	for _, p := range b.players {
		for _, s := range p.Souls {
			souls = append(souls, s)
			playerMap[s.getId()] = &p
		}
	}
	return souls, playerMap
}

func (p player) getSouls() []card {
	return p.Souls
}

func (p player) getSoulIndex(id uint16) (uint8, error) {
	var i uint8
	var err = errors.New("soul not found")
	for i = range p.Souls {
		if p.Souls[i].getId() == id {
			err = nil
			break
		}
	}
	return i, err
}

func (p player) getTappedActiveItems() []*treasureCard {
	var cards = make([]*treasureCard, 0, len(p.ActiveItems))
	for i := range p.ActiveItems {
		c := &p.ActiveItems[i]
		if c.active && c.tapped {
			cards = append(cards, c)
		}
	}
	return cards
}

func (es eventStack) getTriggeredEffectEvents() []*eventNode {
	var nodes = make([]*eventNode, 0, es.size)
	if es.head != nil {
		curr := es.head.top
		for curr != nil {
			if _, ok := curr.event.e.(triggeredEffectEvent); ok {
				nodes = append(nodes, curr)
			}
			curr = curr.next
		}
	}
	return nodes
}

// TODO: Add a requirement function to each card to see if it can be activated
func (p player) getUsableActiveItems() []*treasureCard {
	var cards = make([]*treasureCard, 0, len(p.ActiveItems))
	for i := range p.ActiveItems {
		c := &p.ActiveItems[i]
		if c.active && c.paid {
			// TODO check if can be activated
			cards = append(cards, c)
		} else if c.active && !c.tapped {
			// TODO check if can be activated
			cards = append(cards, c)
		} else { // if c.paid
			// TODO check if can be activated
			cards = append(cards, c)
		}
	}
	return cards
}

func getValidGuppyItems() map[uint16]struct{} {
	return map[uint16]struct{}{
		guppysCollar: {},
		guppysEye:    {},
		guppysHead:   {},
		guppysTail:   {},
		guppysPaw:    {},
		theDeadCat:   {},
	}
}
