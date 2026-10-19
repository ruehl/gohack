// Package term is the tcell-based terminal frontend. It only talks to the
// game through Commands and Views, so it can be replaced by another renderer.
package term

import (
	"time"

	"github.com/gdamore/tcell/v2"

	"gohack/internal/entity"
	"gohack/internal/game"
)

// redrawInterval is how often the screen is redrawn to animate lights.
const redrawInterval = 100 * time.Millisecond

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

	// Light animation needs redraws even when no key is pressed.
	start := time.Now()
	done := make(chan struct{})
	defer close(done)
	go func() {
		tick := time.NewTicker(redrawInterval)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				_ = s.PostEvent(tcell.NewEventInterrupt(nil))
			case <-done:
				return
			}
		}
	}()

	var in Input
	for {
		w, h := s.Size()
		mw, mh := viewportSize(w, h)
		v := g.BuildView(player, mw, mh, time.Since(start))
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
