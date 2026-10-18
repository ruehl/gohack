package game

import (
	"slices"

	"gohack/internal/entity"
	"gohack/internal/geom"
)

// npcPhase lets every NPC act once. NPCs never move; a hostile NPC attacks
// an adjacent living player.
func (g *Game) npcPhase() []Event {
	var evs []Event
	for _, id := range slices.Clone(g.order) {
		npc := g.entities[id]
		if npc == nil || npc.IsPlayer || npc.IsDead() || npc.Disposition != entity.Hostile {
			continue
		}
		for _, d := range geom.AllDirs {
			t := g.EntityAt(npc.Level, npc.Pos.Add(d.Delta()))
			if t != nil && t.IsPlayer && !t.IsDead() {
				evs = append(evs, g.attack(npc, t)...)
				break
			}
		}
	}
	return evs
}
