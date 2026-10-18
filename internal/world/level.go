package world

import "gohack/internal/geom"

// MaxLevelSize is the maximum width and height of a level in tiles.
const MaxLevelSize = 1024

// LevelID uniquely names a level.
type LevelID string

// PortalKind is the type of a portal. Only stairs exist for now.
type PortalKind uint8

const (
	PortalStairs PortalKind = iota
)

// Portal connects a tile of one level to a tile of another.
type Portal struct {
	Kind      PortalKind
	Target    LevelID
	TargetPos geom.Point
}

// Spawn marks where an NPC of the given template is placed at world creation.
type Spawn struct {
	Template string
	Pos      geom.Point
}

// Level is a rectangular tile map.
type Level struct {
	ID       LevelID
	Name     string
	W, H     int
	Portals  map[geom.Point]Portal
	Spawns   []Spawn
	Start    geom.Point // player start; valid only if HasStart
	HasStart bool

	tiles []TileType
}

// NewLevel creates a level filled with Void.
func NewLevel(id LevelID, name string, w, h int) *Level {
	return &Level{
		ID: id, Name: name, W: w, H: h,
		Portals: map[geom.Point]Portal{},
		tiles:   make([]TileType, w*h),
	}
}

func (l *Level) InBounds(p geom.Point) bool {
	return p.X >= 0 && p.Y >= 0 && p.X < l.W && p.Y < l.H
}

// At returns the tile at p; anything outside the level is Void.
func (l *Level) At(p geom.Point) TileType {
	if !l.InBounds(p) {
		return Void
	}
	return l.tiles[p.Y*l.W+p.X]
}

func (l *Level) Set(p geom.Point, t TileType) {
	if l.InBounds(p) {
		l.tiles[p.Y*l.W+p.X] = t
	}
}

func (l *Level) Walkable(p geom.Point) bool    { return l.At(p).Walkable() }
func (l *Level) Transparent(p geom.Point) bool { return l.At(p).Transparent() }
