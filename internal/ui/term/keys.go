package term

import (
	"unicode"

	"github.com/gdamore/tcell/v2"

	"gohack/internal/game"
	"gohack/internal/geom"
)

// Input translates key presses into game commands. It holds the one piece of
// UI state needed for the two-key attack ("F" then a direction).
type Input struct {
	attacking bool
}

// Attacking reports whether the next direction key will attack.
func (in *Input) Attacking() bool { return in.attacking }

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

// Handle maps a key event to a command (nil if none) and whether to quit.
func (in *Input) Handle(ev *tcell.EventKey) (cmd game.Command, quit bool) {
	if ev.Key() == tcell.KeyCtrlC {
		return nil, true
	}
	if ev.Key() == tcell.KeyEscape {
		if in.attacking { // Esc cancels attack mode first
			in.attacking = false
			return nil, false
		}
		return nil, true
	}
	d, isDir := dirOfKey(ev)
	if in.attacking {
		in.attacking = false
		if isDir {
			return game.AttackCmd{Dir: d}, false
		}
		return nil, false // any other key cancels
	}
	if isDir {
		return game.MoveCmd{Dir: d}, false
	}
	if ev.Key() != tcell.KeyRune {
		return nil, false
	}
	switch unicode.ToLower(ev.Rune()) {
	case 'f':
		in.attacking = true
	case 'e':
		return game.UseStairsCmd{}, false
	case 'q':
		return nil, true
	}
	return nil, false
}
