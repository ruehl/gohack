// Package term is the tcell-based terminal frontend. It only talks to the
// game through Commands and Views, so it can be replaced by another renderer.
package term

import (
	"github.com/gdamore/tcell/v2"

	"gohack/internal/entity"
	"gohack/internal/game"
)

// Run plays the game for the given player until they quit or die.
func Run(g *game.Game, player entity.EntityID) error {
	s, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := s.Init(); err != nil {
		return err
	}
	defer s.Fini()

	var in Input
	for {
		w, h := s.Size()
		mw, mh := viewportSize(w, h)
		v := g.BuildView(player, mw, mh)
		if v == nil {
			return nil
		}
		draw(s, v, &in)

		switch ev := s.PollEvent().(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			if v.Dead {
				return nil
			}
			cmd, quit := in.Handle(ev, v)
			if quit {
				return nil
			}
			if cmd != nil {
				g.Submit(player, cmd)
			}
		}
	}
}
