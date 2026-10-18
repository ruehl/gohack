package game

import (
	"fmt"

	"gohack/internal/entity"
)

const maxLogLines = 100

// record turns events into message log lines for every player who perceives them.
func (g *Game) record(evs []Event) {
	for pid, p := range g.views {
		viewer := g.entities[pid]
		if viewer == nil {
			continue
		}
		for _, ev := range evs {
			involved := ev.Actor == pid || ev.Target == pid
			if !involved && (ev.Level != viewer.Level || !p.visible.Has(ev.Pos.X, ev.Pos.Y)) {
				continue
			}
			if line := describe(ev, pid); line != "" {
				g.logs[pid] = append(g.logs[pid], line)
			}
		}
		if n := len(g.logs[pid]); n > maxLogLines {
			g.logs[pid] = g.logs[pid][n-maxLogLines:]
		}
	}
}

// describe renders an event from the viewer's point of view; "" means not worth a message.
func describe(ev Event, viewer entity.EntityID) string {
	actor, target := "The "+ev.ActorName, "the "+ev.TargetName
	you := ev.Actor == viewer
	switch ev.Kind {
	case EvDoorOpened:
		if you {
			return "You open the door."
		}
		return actor + " opens a door."
	case EvBlocked:
		return "The " + ev.TargetName + " is in your way."
	case EvAttack:
		switch {
		case you && ev.Hit:
			return fmt.Sprintf("You hit %s for %d.", target, ev.Damage)
		case you:
			return fmt.Sprintf("You miss %s.", target)
		case ev.Target == viewer && ev.Hit:
			return fmt.Sprintf("%s hits you for %d.", actor, ev.Damage)
		case ev.Target == viewer:
			return fmt.Sprintf("%s misses you.", actor)
		case ev.Hit:
			return fmt.Sprintf("%s hits %s for %d.", actor, target, ev.Damage)
		default:
			return fmt.Sprintf("%s misses %s.", actor, target)
		}
	case EvDied:
		if you {
			return "You die."
		}
		return actor + " dies."
	case EvLevelChanged:
		return "You take the stairs to " + ev.Text + "."
	case EvEquipped:
		if you {
			return "You equip the " + ev.Text + "."
		}
	case EvUnequipped:
		if you {
			return "You take off the " + ev.Text + "."
		}
	case EvMessage:
		if you {
			return ev.Text
		}
	}
	return ""
}
