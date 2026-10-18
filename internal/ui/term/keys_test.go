package term

import (
	"testing"

	"github.com/gdamore/tcell/v2"

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
		cmd, quit := in.Handle(c.ev)
		if m, ok := cmd.(game.MoveCmd); !ok || m.Dir != c.want || quit {
			t.Errorf("%v: got %v quit=%v, want move %v", c.ev.Name(), cmd, quit, c.want)
		}
	}
}

func TestAttackWithArrow(t *testing.T) {
	var in Input
	in.Handle(tcell.NewEventKey(tcell.KeyRune, 'f', 0))
	if !in.Attacking() {
		t.Fatal("F should enter attack mode")
	}
	cmd, _ := in.Handle(tcell.NewEventKey(tcell.KeyLeft, 0, 0))
	if a, ok := cmd.(game.AttackCmd); !ok || a.Dir != geom.W || in.Attacking() {
		t.Fatalf("got %v attacking=%v", cmd, in.Attacking())
	}
	in.Handle(tcell.NewEventKey(tcell.KeyRune, 'f', 0))
	if cmd, _ := in.Handle(tcell.NewEventKey(tcell.KeyRune, 'x', 0)); cmd != nil || in.Attacking() {
		t.Fatal("other key must cancel attack mode")
	}
}
