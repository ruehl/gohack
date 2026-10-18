package entity

import (
	"slices"
	"testing"
)

// bare returns an entity with an empty inventory and no skills.
func bare() *Entity {
	return &Entity{Skills: map[Skill]bool{}, Equipment: map[Slot]*Item{}}
}

func carry(e *Entity, d *ItemDef) *Item {
	it := &Item{Def: d}
	e.Inventory.Add(it)
	return it
}

func TestFits(t *testing.T) {
	e := bare()
	dagger, sword := &Item{Def: Dagger}, &Item{Def: Greatsword}
	if e.Fits(dagger, SlotOffHand) || !e.Fits(dagger, SlotMainHand) {
		t.Fatal("a one-handed weapon goes in the off hand only with dual-wielding")
	}
	e.Skills[DualWielding] = true
	if !e.Fits(dagger, SlotOffHand) || e.Fits(sword, SlotOffHand) || !e.Fits(sword, SlotMainHand) {
		t.Fatal("dual-wielding never allows a two-handed weapon in the off hand")
	}
	if !e.Fits(&Item{Def: WoodenShield}, SlotOffHand) || e.Fits(&Item{Def: WoodenShield}, SlotMainHand) {
		t.Fatal("a shield goes in the off hand only")
	}
	ring := &Item{Def: CopperRing}
	if !e.Fits(ring, SlotRing1) || !e.Fits(ring, SlotRing2) || e.Fits(ring, SlotNeck) {
		t.Fatal("rings go in either ring slot")
	}
	if e.Fits(&Item{Def: LeatherCap}, SlotFeet) || !e.Fits(&Item{Def: LeatherBoots}, SlotFeet) || e.Fits(ring, SlotNone) {
		t.Fatal("armor goes in its own slot only")
	}
}

func TestEquipAtTwoHandedClearsHands(t *testing.T) {
	e := bare()
	sword, shield, great := carry(e, LongSword), carry(e, WoodenShield), carry(e, Greatsword)
	e.EquipAt(sword, SlotMainHand)
	e.EquipAt(shield, SlotOffHand)
	removed, err := e.EquipAt(great, SlotMainHand)
	if err != nil || !slices.Equal(removed, []*Item{sword, shield}) {
		t.Fatalf("removed=%v err=%v", removed, err)
	}
	if e.Equipment[SlotMainHand] != great || e.Equipment[SlotOffHand] != nil {
		t.Fatalf("equipment = %v", e.Equipment)
	}
	if len(e.Inventory.Items) != 3 {
		t.Fatal("removed items must stay in the inventory")
	}
}

func TestEquipAtShieldClearsTwoHanded(t *testing.T) {
	e := bare()
	great, shield := carry(e, Greatsword), carry(e, WoodenShield)
	e.EquipAt(great, SlotMainHand)
	removed, err := e.EquipAt(shield, SlotOffHand)
	if err != nil || !slices.Equal(removed, []*Item{great}) || e.Equipment[SlotMainHand] != nil {
		t.Fatalf("removed=%v err=%v equipment=%v", removed, err, e.Equipment)
	}
}

func TestEquipAtKeepsOneHandedWeaponWithShield(t *testing.T) {
	e := bare()
	sword, shield := carry(e, LongSword), carry(e, WoodenShield)
	e.EquipAt(sword, SlotMainHand)
	if removed, err := e.EquipAt(shield, SlotOffHand); err != nil || len(removed) != 0 {
		t.Fatalf("removed=%v err=%v", removed, err)
	}
}

func TestEquipAtMovesRing(t *testing.T) {
	e := bare()
	a, b := carry(e, CopperRing), carry(e, CopperRing)
	e.EquipAt(a, SlotRing1)
	e.EquipAt(b, SlotRing2)
	removed, err := e.EquipAt(a, SlotRing2)
	if err != nil || !slices.Equal(removed, []*Item{b}) {
		t.Fatalf("removed=%v err=%v", removed, err)
	}
	if e.Equipment[SlotRing1] != nil || e.Equipment[SlotRing2] != a {
		t.Fatalf("a ring that moves must leave its old slot empty: %v", e.Equipment)
	}
}

func TestEquipAtErrorsChangeNothing(t *testing.T) {
	e := bare()
	sword, ring := carry(e, LongSword), carry(e, CopperRing)
	e.EquipAt(sword, SlotMainHand)
	for _, c := range []struct {
		it   *Item
		slot Slot
		want error
	}{
		{&Item{Def: Dagger}, SlotOffHand, ErrNotCarried},
		{ring, SlotHead, ErrDoesNotFit},
		{sword, SlotOffHand, ErrDoesNotFit},
		{sword, SlotMainHand, ErrAlreadyWorn},
	} {
		if removed, err := e.EquipAt(c.it, c.slot); err != c.want || removed != nil {
			t.Errorf("EquipAt(%s, %d) = %v, %v; want %v", c.it.Def.ID, c.slot, removed, err, c.want)
		}
	}
	if len(e.Equipment) != 1 || e.Equipment[SlotMainHand] != sword {
		t.Fatalf("equipment = %v", e.Equipment)
	}
}

func TestUnequip(t *testing.T) {
	e := bare()
	helmet := carry(e, LeatherCap)
	e.EquipAt(helmet, SlotHead)
	if got := e.Unequip(SlotHead); got != helmet || e.Equipment[SlotHead] != nil || len(e.Inventory.Items) != 1 {
		t.Fatalf("Unequip = %v, equipment = %v", got, e.Equipment)
	}
	if e.Unequip(SlotHead) != nil {
		t.Fatal("an empty slot yields nil")
	}
}

func TestEquipPicksSlots(t *testing.T) {
	e := bare()
	r1, r2, r3 := &Item{Def: CopperRing}, &Item{Def: CopperRing}, &Item{Def: CopperRing}
	e.Equip(r1)
	e.Equip(r2)
	if e.Equipment[SlotRing1] != r1 || e.Equipment[SlotRing2] != r2 {
		t.Fatal("rings fill Ring1 then Ring2")
	}
	if removed := e.Equip(r3); !slices.Equal(removed, []*Item{r1}) || e.Equipment[SlotRing1] != r3 {
		t.Fatalf("a third ring replaces Ring1, removed=%v", removed)
	}

	d1, d2 := &Item{Def: Dagger}, &Item{Def: Dagger}
	e.Equip(d1)
	if removed := e.Equip(d2); !slices.Equal(removed, []*Item{d1}) || e.Equipment[SlotMainHand] != d2 || e.Equipment[SlotOffHand] != nil {
		t.Fatalf("without dual-wielding a second weapon replaces the first, removed=%v", removed)
	}

	e = bare()
	e.Skills[DualWielding] = true
	e.Equip(d1)
	if removed := e.Equip(d2); len(removed) != 0 || e.Equipment[SlotMainHand] != d1 || e.Equipment[SlotOffHand] != d2 {
		t.Fatalf("a dual-wielder wields both, removed=%v", removed)
	}
	if ws := e.Weapons(); !slices.Equal(ws, []*Item{d1, d2}) {
		t.Fatalf("Weapons = %v", ws)
	}
}

func TestWeaponsIgnoresShield(t *testing.T) {
	e := bare()
	sword, shield := &Item{Def: LongSword}, &Item{Def: WoodenShield}
	e.Equip(sword)
	e.Equip(shield)
	if ws := e.Weapons(); !slices.Equal(ws, []*Item{sword}) {
		t.Fatalf("Weapons = %v", ws)
	}
}

func TestStartingGearFitsWithoutSwaps(t *testing.T) {
	e := bare()
	e.Skills = skillSet(Knight.Skills)
	for _, g := range Knight.Gear {
		if removed := e.Equip(&Item{Def: g}); len(removed) != 0 {
			t.Errorf("knight gear %s displaced %v", g.ID, removed)
		}
	}
	for id, tpl := range NPCTemplates {
		e := bare()
		e.Skills = skillSet(tpl.Skills)
		for _, g := range tpl.Gear {
			if removed := e.Equip(&Item{Def: g}); len(removed) != 0 {
				t.Errorf("%s gear %s displaced %v", id, g.ID, removed)
			}
		}
	}
}
