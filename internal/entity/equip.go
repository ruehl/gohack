package entity

import (
	"errors"
	"slices"
)

var (
	ErrNotCarried  = errors.New("item is not carried")
	ErrDoesNotFit  = errors.New("item does not fit in that slot")
	ErrAlreadyWorn = errors.New("item is already worn in that slot")
)

// Fits reports whether the item may be worn in slot s by this entity.
func (e *Entity) Fits(it *Item, s Slot) bool {
	d := it.Def
	switch s {
	case SlotNone:
		return false
	case SlotRing1, SlotRing2:
		return d.Kind == KindRing
	case SlotMainHand:
		return d.Kind == KindWeapon
	case SlotOffHand:
		return d.Kind == KindShield ||
			(d.Kind == KindWeapon && !d.TwoHanded && e.Has(DualWielding))
	}
	return d.Slot == s
}

// SlotOf returns the slot the item is worn in, or SlotNone.
func (e *Entity) SlotOf(it *Item) Slot {
	for _, s := range Slots {
		if e.Equipment[s] == it {
			return s
		}
	}
	return SlotNone
}

// EquipAt wears a carried item in slot s. An item already worn elsewhere
// moves. Items in the way (the slot's occupant, and the off hand or a
// two-handed weapon that cannot be combined) are taken off and returned; they
// stay in the inventory. On error nothing changes.
func (e *Entity) EquipAt(it *Item, s Slot) (removed []*Item, err error) {
	switch {
	case !slices.Contains(e.Inventory.Items, it):
		return nil, ErrNotCarried
	case !e.Fits(it, s):
		return nil, ErrDoesNotFit
	case e.Equipment[s] == it:
		return nil, ErrAlreadyWorn
	}
	if from := e.SlotOf(it); from != SlotNone {
		delete(e.Equipment, from)
	}
	takeOff := func(slot Slot) {
		if old := e.Equipment[slot]; old != nil {
			delete(e.Equipment, slot)
			removed = append(removed, old)
		}
	}
	takeOff(s)
	switch {
	case s == SlotMainHand && it.Def.TwoHanded:
		takeOff(SlotOffHand)
	case s == SlotOffHand:
		if main := e.Equipment[SlotMainHand]; main != nil && main.Def.TwoHanded {
			takeOff(SlotMainHand)
		}
	}
	e.Equipment[s] = it
	return removed, nil
}

// Unequip takes off whatever is in slot s and returns it (nil if the slot was
// empty). The item stays in the inventory.
func (e *Entity) Unequip(s Slot) *Item {
	it := e.Equipment[s]
	delete(e.Equipment, s)
	return it
}

// Equip adds the item to the inventory and wears it in the slot that suits it
// best, returning anything that had to come off. Items without a slot are
// only carried.
func (e *Entity) Equip(it *Item) []*Item {
	if !slices.Contains(e.Inventory.Items, it) {
		e.Inventory.Add(it)
	}
	s := e.slotFor(it)
	if s == SlotNone {
		return nil
	}
	removed, _ := e.EquipAt(it, s)
	return removed
}

// slotFor picks where Equip puts an item.
func (e *Entity) slotFor(it *Item) Slot {
	d := it.Def
	switch {
	case d.Kind == KindRing:
		switch {
		case e.Equipment[SlotRing1] == nil:
			return SlotRing1
		case e.Equipment[SlotRing2] == nil:
			return SlotRing2
		}
		return SlotRing1
	case d.Kind == KindWeapon && !d.TwoHanded:
		main := e.Equipment[SlotMainHand]
		if main != nil && e.Has(DualWielding) && !main.Def.TwoHanded && e.Equipment[SlotOffHand] == nil {
			return SlotOffHand
		}
		return SlotMainHand
	}
	return d.Slot
}

// Weapons returns the weapons held, main hand first. A shield is not a weapon.
func (e *Entity) Weapons() []*Item {
	var ws []*Item
	for _, s := range []Slot{SlotMainHand, SlotOffHand} {
		if it := e.Equipment[s]; it != nil && it.Def.Kind == KindWeapon {
			ws = append(ws, it)
		}
	}
	return ws
}
