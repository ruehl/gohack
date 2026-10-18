package game

import (
	"slices"
	"strings"
	"testing"

	"gohack/internal/entity"
	"gohack/internal/geom"
)

// itemIndex returns the inventory index of the first item with the given name.
func itemIndex(t *testing.T, e *entity.Entity, name string) int {
	t.Helper()
	for i, it := range e.Inventory.Items {
		if it.Def.Name == name {
			return i
		}
	}
	t.Fatalf("no %q in inventory", name)
	return -1
}

func kinds(evs []Event) []EventKind {
	var ks []EventKind
	for _, ev := range evs {
		ks = append(ks, ev.Kind)
	}
	return ks
}

func TestDualWielderAttacksTwice(t *testing.T) {
	// Long sword: roll 20, 1d8 rolls IntN 1 -> 2, +2 STR = 4. Dagger: roll 20, 1d4 rolls IntN 0 -> 1, +2 = 3.
	g, p := newGame(t, &seq{19, 1, 19, 0}, map[string]string{"a": duel})
	gob := g.EntityAt("a", geom.Point{X: 2, Y: 1})
	p.Skills[entity.DualWielding] = true
	p.Unequip(entity.SlotOffHand) // the shield
	if _, err := p.EquipAt(p.Inventory.Items[itemIndex(t, p, "dagger")], entity.SlotOffHand); err != nil {
		t.Fatal(err)
	}
	evs := g.Submit(p.ID, AttackCmd{geom.E})
	var mine []Event
	for _, ev := range evs {
		if ev.Kind == EvAttack && ev.Actor == p.ID {
			mine = append(mine, ev)
		}
	}
	if len(mine) != 2 || mine[0].Damage != 4 || mine[1].Damage != 3 || gob.HP != 18-7 {
		t.Fatalf("attacks=%+v goblin hp=%d, want two hits for 4 and 3", mine, gob.HP)
	}
	if g.Turn != 1 {
		t.Fatalf("both attacks are one turn, turn=%d", g.Turn)
	}
}

func TestDualWielderStopsWhenTargetDies(t *testing.T) {
	g, p := newGame(t, &seq{19, 1, 19, 0}, map[string]string{"a": duel})
	gob := g.EntityAt("a", geom.Point{X: 2, Y: 1})
	gob.HP = 1
	p.Skills[entity.DualWielding] = true
	p.Unequip(entity.SlotOffHand)
	p.EquipAt(p.Inventory.Items[itemIndex(t, p, "dagger")], entity.SlotOffHand)
	evs := g.Submit(p.ID, AttackCmd{geom.E})
	if want := []EventKind{EvAttack, EvDied}; !slices.Equal(kinds(evs), want) {
		t.Fatalf("events = %v, want %v", kinds(evs), want)
	}
}

func TestUnarmedAttacksOnce(t *testing.T) {
	g, p := newGame(t, &seq{19, 1}, map[string]string{"a": duel})
	p.Unequip(entity.SlotMainHand)
	evs := g.Submit(p.ID, AttackCmd{geom.E})
	if len(evs) < 2 || evs[0].Kind != EvAttack || evs[0].Damage != 2+2 {
		t.Fatalf("expected a bare-hands hit for 4 (1d2=2, +2 STR), got %+v", evs)
	}
	if evs[1].Actor == p.ID {
		t.Fatal("the second event should be the goblin's, not another knight attack")
	}
}

func TestEquipGreatswordCostsTurn(t *testing.T) {
	g, p := newGame(t, &seq{}, map[string]string{"a": duel}) // the goblin misses on roll 1
	ac := p.ArmorClass()
	evs := g.Submit(p.ID, EquipCmd{Item: itemIndex(t, p, "greatsword"), Slot: entity.SlotMainHand})
	want := []EventKind{EvUnequipped, EvUnequipped, EvEquipped, EvAttack}
	if !slices.Equal(kinds(evs), want) {
		t.Fatalf("events = %v, want %v", kinds(evs), want)
	}
	if evs[0].Text != "long sword" || evs[1].Text != "wooden shield" || evs[2].Text != "greatsword" {
		t.Fatalf("texts = %q %q %q", evs[0].Text, evs[1].Text, evs[2].Text)
	}
	if g.Turn != 1 || p.ArmorClass() != ac-1 {
		t.Fatalf("turn=%d ac=%d (was %d)", g.Turn, p.ArmorClass(), ac)
	}
	log := strings.Join(g.BuildView(p.ID, 9, 5).Messages, "|")
	if !strings.Contains(log, "You take off the wooden shield.") || !strings.Contains(log, "You equip the greatsword.") {
		t.Fatalf("messages = %q", log)
	}
}

func TestRefusedGearChangesAreFree(t *testing.T) {
	g, p := newGame(t, &seq{}, map[string]string{"a": duel})
	for name, cmd := range map[string]Command{
		"ring in head slot":   EquipCmd{Item: itemIndex(t, p, "copper ring"), Slot: entity.SlotHead},
		"bad index":           EquipCmd{Item: 99, Slot: entity.SlotHead},
		"already worn":        EquipCmd{Item: itemIndex(t, p, "long sword"), Slot: entity.SlotMainHand},
		"dagger off hand":     EquipCmd{Item: itemIndex(t, p, "dagger"), Slot: entity.SlotOffHand},
		"unequip empty slot":  UnequipCmd{Slot: entity.SlotHead},
		"unequip unknown":     UnequipCmd{Slot: entity.SlotNone},
		"greatsword off hand": EquipCmd{Item: itemIndex(t, p, "greatsword"), Slot: entity.SlotOffHand},
	} {
		evs := g.Submit(p.ID, cmd)
		if g.Turn != 0 || len(evs) != 1 || evs[0].Kind != EvMessage {
			t.Errorf("%s: turn=%d events=%+v, want one message and no turn", name, g.Turn, evs)
		}
	}
	if p.Equipment[entity.SlotMainHand].Def != entity.LongSword || p.Equipment[entity.SlotOffHand].Def != entity.WoodenShield {
		t.Fatal("refused changes must not alter the equipment")
	}
	msgs := g.BuildView(p.ID, 9, 5).Messages
	if !strings.Contains(strings.Join(msgs, "|"), "The copper ring doesn't fit there.") {
		t.Fatalf("messages = %v", msgs)
	}
}

func TestUnequipCostsTurn(t *testing.T) {
	g, p := newGame(t, &seq{}, map[string]string{"a": duel})
	evs := g.Submit(p.ID, UnequipCmd{Slot: entity.SlotOffHand})
	if evs[0].Kind != EvUnequipped || g.Turn != 1 || p.Equipment[entity.SlotOffHand] != nil {
		t.Fatalf("events=%+v turn=%d", evs, g.Turn)
	}
	if len(p.Inventory.Items) != 12 {
		t.Fatalf("the shield must stay in the inventory, have %d items", len(p.Inventory.Items))
	}
}

func TestViewGear(t *testing.T) {
	g, p := newGame(t, &seq{}, map[string]string{"a": duel})
	gear := g.BuildView(p.ID, 9, 5).Gear
	if len(gear.Slots) != len(entity.Slots) || len(gear.Items) != len(p.Inventory.Items) {
		t.Fatalf("slots=%d items=%d", len(gear.Slots), len(gear.Items))
	}
	for _, s := range gear.Slots {
		want := ""
		switch s.Slot {
		case entity.SlotBody:
			want = "chain mail"
		case entity.SlotMainHand:
			want = "long sword"
		case entity.SlotOffHand:
			want = "wooden shield"
		}
		if s.Item != want {
			t.Errorf("slot %d holds %q, want %q", s.Slot, s.Item, want)
		}
	}
	byName := map[string]ItemView{}
	for _, it := range gear.Items {
		byName[it.Name] = it
	}
	if it := byName["long sword"]; it.Worn != entity.SlotMainHand || !slices.Equal(it.Fits, []entity.Slot{entity.SlotMainHand}) {
		t.Errorf("long sword = %+v", it)
	}
	if it := byName["copper ring"]; it.Worn != entity.SlotNone || !slices.Equal(it.Fits, []entity.Slot{entity.SlotRing1, entity.SlotRing2}) {
		t.Errorf("copper ring = %+v", it)
	}
	if it := byName["wooden shield"]; !slices.Equal(it.Fits, []entity.Slot{entity.SlotOffHand}) {
		t.Errorf("wooden shield = %+v", it)
	}
}
