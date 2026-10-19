package term

import (
	"slices"
	"unicode"

	"github.com/gdamore/tcell/v2"

	"gohack/internal/entity"
	"gohack/internal/game"
	"gohack/internal/geom"
)

// prompt is a command waiting for a direction key.
type prompt uint8

const (
	promptNone   prompt = iota
	promptAttack        // "F" pressed
	promptLight         // "L" pressed
)

// Input translates key presses into game commands. It holds the UI state: the
// two-key commands (attack or toggle a light: a letter, then a direction) and
// the gear screen.
type Input struct {
	prompt prompt
	gear   gearMenu
}

// gearMenu is the state of the gear screen: a list of slots and, once a slot
// is chosen, a list of the items that can go there.
type gearMenu struct {
	open     bool
	picking  bool        // choosing an item for slot
	slot     entity.Slot // the slot being filled, while picking
	cursor   int
	slotCurs int // cursor on the slot list, restored when leaving the item list
}

// choice is one row of the item list. Item is the inventory index, or -1 for
// "(nothing)", which takes the slot's item off.
type choice struct {
	Item int
	Name string
}

// Attacking reports whether the next direction key will attack.
func (in *Input) Attacking() bool { return in.prompt == promptAttack }

// Toggling reports whether the next direction key will toggle a light.
func (in *Input) Toggling() bool { return in.prompt == promptLight }

// GearOpen reports whether the gear screen is showing.
func (in *Input) GearOpen() bool { return in.gear.open }

// choices lists what can be put in slot: "(nothing)" first if the slot is
// occupied, then every carried item that fits and is not already there.
func choices(g *game.Gear, slot entity.Slot) []choice {
	var cs []choice
	for _, s := range g.Slots {
		if s.Slot == slot && s.Item != "" {
			cs = append(cs, choice{-1, "(nothing)"})
		}
	}
	for i, it := range g.Items {
		if it.Worn != slot && slices.Contains(it.Fits, slot) {
			cs = append(cs, choice{i, it.Name})
		}
	}
	return cs
}

// rowCount is the number of rows in the list currently shown.
func (m *gearMenu) rowCount(g *game.Gear) int {
	if m.picking {
		return len(choices(g, m.slot))
	}
	return len(g.Slots)
}

// handleGear drives the gear screen. Esc steps back a level; I closes it.
func (in *Input) handleGear(ev *tcell.EventKey, g *game.Gear) game.Command {
	m := &in.gear
	switch {
	case ev.Key() == tcell.KeyEscape:
		if m.picking {
			m.picking, m.cursor = false, m.slotCurs
		} else {
			m.open = false
		}
		return nil
	case ev.Key() == tcell.KeyRune && unicode.ToLower(ev.Rune()) == 'i':
		*m = gearMenu{}
		return nil
	case ev.Key() == tcell.KeyEnter:
		return in.chooseGear(g)
	}
	if d, ok := dirOfKey(ev); ok {
		switch d {
		case geom.N:
			m.cursor = max(0, m.cursor-1)
		case geom.S:
			m.cursor = max(0, min(m.cursor+1, m.rowCount(g)-1))
		}
	}
	return nil
}

// chooseGear acts on Enter: pick the slot under the cursor, or apply the item under it.
func (in *Input) chooseGear(g *game.Gear) game.Command {
	m := &in.gear
	if !m.picking {
		if m.cursor >= len(g.Slots) {
			return nil
		}
		m.slot, m.slotCurs, m.picking, m.cursor = g.Slots[m.cursor].Slot, m.cursor, true, 0
		return nil
	}
	cs := choices(g, m.slot)
	if m.cursor >= len(cs) {
		return nil
	}
	c, slot := cs[m.cursor], m.slot
	m.picking, m.cursor = false, m.slotCurs
	if c.Item < 0 {
		return game.UnequipCmd{Slot: slot}
	}
	return game.EquipCmd{Item: c.Item, Slot: slot}
}

// dirOfKey maps WASD and the arrow keys to a direction.
func dirOfKey(ev *tcell.EventKey) (geom.Dir, bool) {
	switch ev.Key() {
	case tcell.KeyUp:
		return geom.N, true
	case tcell.KeyLeft:
		return geom.W, true
	case tcell.KeyDown:
		return geom.S, true
	case tcell.KeyRight:
		return geom.E, true
	case tcell.KeyRune:
		return dirOf(ev.Rune())
	}
	return 0, false
}

func dirOf(r rune) (geom.Dir, bool) {
	switch unicode.ToLower(r) {
	case 'w':
		return geom.N, true
	case 'a':
		return geom.W, true
	case 's':
		return geom.S, true
	case 'd':
		return geom.E, true
	}
	return 0, false
}

// Handle maps a key event to a command (nil if none) and whether to quit. v is
// the view the player is looking at; the gear screen reads its choices from it.
func (in *Input) Handle(ev *tcell.EventKey, v *game.View) (cmd game.Command, quit bool) {
	if ev.Key() == tcell.KeyCtrlC {
		return nil, true
	}
	if in.gear.open {
		return in.handleGear(ev, &v.Gear), false
	}
	if ev.Key() == tcell.KeyEscape {
		if in.prompt != promptNone { // Esc cancels a pending command first
			in.prompt = promptNone
			return nil, false
		}
		return nil, true
	}
	d, isDir := dirOfKey(ev)
	if p := in.prompt; p != promptNone {
		in.prompt = promptNone
		switch {
		case !isDir:
			return nil, false // any other key cancels
		case p == promptLight:
			return game.ToggleLightCmd{Dir: d}, false
		}
		return game.AttackCmd{Dir: d}, false
	}
	if isDir {
		return game.MoveCmd{Dir: d}, false
	}
	if ev.Key() != tcell.KeyRune {
		return nil, false
	}
	switch unicode.ToLower(ev.Rune()) {
	case 'f':
		in.prompt = promptAttack
	case 'l':
		in.prompt = promptLight
	case 'e':
		return game.UseStairsCmd{}, false
	case 'i':
		in.gear = gearMenu{open: true}
	case 'q':
		return nil, true
	}
	return nil, false
}
