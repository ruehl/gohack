package term

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"gohack/internal/entity"
	"gohack/internal/game"
	"gohack/internal/geom"
)

func TestArrowsAndWASD(t *testing.T) {
	var in Input
	for _, c := range []struct {
		ev   *tcell.EventKey
		want geom.Dir
	}{
		{tcell.NewEventKey(tcell.KeyUp, 0, 0), geom.N},
		{tcell.NewEventKey(tcell.KeyDown, 0, 0), geom.S},
		{tcell.NewEventKey(tcell.KeyLeft, 0, 0), geom.W},
		{tcell.NewEventKey(tcell.KeyRight, 0, 0), geom.E},
		{tcell.NewEventKey(tcell.KeyRune, 'w', 0), geom.N},
		{tcell.NewEventKey(tcell.KeyRune, 'D', 0), geom.E},
	} {
		cmd, quit := in.Handle(c.ev, &game.View{})
		if m, ok := cmd.(game.MoveCmd); !ok || m.Dir != c.want || quit {
			t.Errorf("%v: got %v quit=%v, want move %v", c.ev.Name(), cmd, quit, c.want)
		}
	}
}

func TestAttackWithArrow(t *testing.T) {
	var in Input
	v := &game.View{}
	in.Handle(tcell.NewEventKey(tcell.KeyRune, 'f', 0), v)
	if !in.Attacking() {
		t.Fatal("F should enter attack mode")
	}
	cmd, _ := in.Handle(tcell.NewEventKey(tcell.KeyLeft, 0, 0), v)
	if a, ok := cmd.(game.AttackCmd); !ok || a.Dir != geom.W || in.Attacking() {
		t.Fatalf("got %v attacking=%v", cmd, in.Attacking())
	}
	in.Handle(tcell.NewEventKey(tcell.KeyRune, 'f', 0), v)
	if cmd, _ := in.Handle(tcell.NewEventKey(tcell.KeyRune, 'x', 0), v); cmd != nil || in.Attacking() {
		t.Fatal("other key must cancel attack mode")
	}
}

// gearView is a player wearing a helmet, with a spare helmet and a ring.
func gearView() *game.View {
	return &game.View{Gear: game.Gear{
		Slots: []game.SlotView{
			{Slot: entity.SlotHead, Item: "leather cap"},
			{Slot: entity.SlotNeck},
			{Slot: entity.SlotRing1},
		},
		Items: []game.ItemView{
			{Name: "leather cap", Worn: entity.SlotHead, Fits: []entity.Slot{entity.SlotHead}},
			{Name: "copper ring", Fits: []entity.Slot{entity.SlotRing1, entity.SlotRing2}},
			{Name: "iron helm", Fits: []entity.Slot{entity.SlotHead}},
		},
	}}
}

func press(in *Input, v *game.View, ev *tcell.EventKey) game.Command {
	cmd, _ := in.Handle(ev, v)
	return cmd
}

func TestGearMenuEquipsAndUnequips(t *testing.T) {
	var in Input
	v := gearView()
	key := func(k tcell.Key) *tcell.EventKey { return tcell.NewEventKey(k, 0, 0) }
	press(&in, v, tcell.NewEventKey(tcell.KeyRune, 'i', 0))
	if !in.GearOpen() {
		t.Fatal("I should open the gear screen")
	}
	// Head slot, choose the helm: the list is "(nothing)" then "iron helm"
	// (the worn cap is not offered again).
	press(&in, v, key(tcell.KeyEnter))
	press(&in, v, key(tcell.KeyDown))
	cmd := press(&in, v, key(tcell.KeyEnter))
	if c, ok := cmd.(game.EquipCmd); !ok || c.Item != 2 || c.Slot != entity.SlotHead {
		t.Fatalf("got %#v, want EquipCmd of the helm into Head", cmd)
	}
	if !in.GearOpen() || in.gear.picking || in.gear.cursor != 0 {
		t.Fatalf("the screen stays open on the slot list: %+v", in.gear)
	}
	// "(nothing)" takes the cap off.
	press(&in, v, key(tcell.KeyEnter))
	if c, ok := press(&in, v, key(tcell.KeyEnter)).(game.UnequipCmd); !ok || c.Slot != entity.SlotHead {
		t.Fatalf("got %#v, want UnequipCmd for Head", c)
	}
}

func TestGearMenuNavigation(t *testing.T) {
	var in Input
	v := gearView()
	key := func(k tcell.Key) *tcell.EventKey { return tcell.NewEventKey(k, 0, 0) }
	press(&in, v, tcell.NewEventKey(tcell.KeyRune, 'i', 0))
	for range 5 { // the cursor stops at the last slot
		press(&in, v, key(tcell.KeyDown))
	}
	if in.gear.cursor != 2 {
		t.Fatalf("cursor = %d, want 2", in.gear.cursor)
	}
	press(&in, v, key(tcell.KeyEnter)) // Ring 1: copper ring only
	if cmd := press(&in, v, key(tcell.KeyEnter)); cmd == nil {
		t.Fatal("the ring should be offered")
	}
	press(&in, v, key(tcell.KeyUp))    // the cursor is back on Ring 1; move to Neck
	press(&in, v, key(tcell.KeyEnter)) // Neck: nothing fits
	if cmd := press(&in, v, key(tcell.KeyEnter)); cmd != nil {
		t.Fatalf("an empty list must not produce %#v", cmd)
	}
	press(&in, v, key(tcell.KeyEscape)) // back to the slot list, cursor on Neck
	if !in.GearOpen() || in.gear.picking || in.gear.cursor != 1 {
		t.Fatalf("Esc should step back: %+v", in.gear)
	}
	if _, quit := in.Handle(key(tcell.KeyEscape), v); quit || in.GearOpen() {
		t.Fatal("Esc on the slot list closes the screen without quitting")
	}
	press(&in, v, tcell.NewEventKey(tcell.KeyRune, 'i', 0))
	press(&in, v, key(tcell.KeyEnter))
	press(&in, v, tcell.NewEventKey(tcell.KeyRune, 'i', 0))
	if in.GearOpen() {
		t.Fatal("I closes the screen from the item list too")
	}
}
