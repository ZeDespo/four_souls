package four_souls

import (
	"errors"
)

func mergeEventSlices(a []*eventNode, b []*eventNode) []*eventNode {
	var i, j, k int
	var l1, l2 = len(a), len(b)
	c := make([]*eventNode, l1+l2, l1+l2)
	for i < l1 && j < l2 {
		if a[i].id < b[j].id {
			c[k] = a[i]
			i += 1
		} else {
			c[k] = b[j]
			j += 1
		}
		k += 1
	}
	for i < l1 {
		c[k] = a[i]
		k += 1
		i += 1
	}
	for j < l2 {
		c[k] = b[j]
		k += 1
		j += 1
	}
	return c
}

// if tc.id == theBone or techX and func is not add counters, do not tap the card.
func (b *Board) activateActiveItem(p *player, tc *treasureCard) error {
	var err = errors.New("not a value that can be activated")
	e := activateEvent{c: tc}
	if tc.active {
		var f cardEffect
		var specialCondition bool
		f, specialCondition, err = tc.f(p, b, tc)
		if err == nil {
			tc.triggered = true
			if specialCondition && tc.id == guppysPaw {
				defer b.eventStack.push(event{p: p, e: damageEvent{target: p, n: 1}})
			} else if specialCondition && (tc.id == theBone || tc.id == techX) {
				tc.triggered = false
			} else if specialCondition {
				defer b.rollDiceAndPush()
			}
			e.f = f
			b.eventStack.push(event{p: p, e: e})
		}
	}
	return err
}

func (b *Board) eventDependentPassiveActivation(p *player, e *eventNode) []event {
	triggeredEvents := make([]event, 0, len(p.PassiveItems)*2)
	skip := map[uint16]struct{}{brokenAnkh: {}, guppysHairball: {}, theDeadCat: {},
		guppysCollar: {}, oneUp: {}} // These have preventative effects, not reactive. Skip
	for i := range p.PassiveItems {
		var ic itemCard = p.PassiveItems[i]
		if ef := ic.getEventPassive(); ef != nil {
			if f, specialCond, err := ef(p, b, ic, e); err == nil {
				if _, ok := skip[ic.getId()]; !ok {
					triggeredEvents = append(triggeredEvents, event{p: p, e: triggeredEffectEvent{c: ic, f: f}})
					if specialCond {
						e, _ := b.rollDice()
						triggeredEvents = append(triggeredEvents, event{p: p, e: e})
					}
				}
			}
		}
	}
	return triggeredEvents
}

// Activate a loot card from the hand
// If the loot card is a trinket, it will be added to the board as a passive item.
// p *player is the player who played the loot card
// i uint8 is the index of the card in the hand.
func (b *Board) playLootCardFromHand(p *player, i uint8) error {
	var lootCard = p.Hand[i]
	var err error
	e := lootCardEvent{l: lootCard}
	if lootCard.trinket {
		e.f = func(roll uint8, blankCard bool) {}
		p.addCardToBoard(lootCard)
	} else {
		var f lootCardEffect
		var specialCondition bool
		f, specialCondition, err = lootCard.f(p, b)
		if err == nil {
			if specialCondition && lootCard.id != temperance {
				defer b.rollDiceAndPush()
			} else if !specialCondition && lootCard.id == temperance {
				defer b.eventStack.push(event{p: p, e: damageEvent{target: p, n: 1}})
			} else if specialCondition && lootCard.id == temperance {
				defer b.eventStack.push(event{p: p, e: damageEvent{target: p, n: 2}})
			}
			e.f = f
			b.discard(p.popHandCard(i))
			b.eventStack.push(event{p: p, e: e})
		}
	}
	return err
}
