package world

import (
	"embed"
	"fmt"
)

// World is the set of all levels. Entities live in the game, not here.
type World struct {
	Levels map[LevelID]*Level
	Order  []LevelID // sorted IDs, for deterministic iteration
	Start  LevelID   // level containing the player start
}

func (w *World) Level(id LevelID) *Level { return w.Levels[id] }

//go:embed levels/*.txt
var embedded embed.FS

// LoadEmbedded loads the levels compiled into the binary.
func LoadEmbedded() (*World, error) {
	w, err := Load(embedded, "levels")
	if err != nil {
		return nil, fmt.Errorf("load embedded levels: %w", err)
	}
	return w, nil
}
