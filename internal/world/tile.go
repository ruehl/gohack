// Package world contains the static map data: tiles, levels and portals.
package world

// TileType identifies a kind of tile. The zero value is Void.
type TileType uint8

const (
	Void TileType = iota
	Floor
	Wall
	DoorClosed
	DoorOpen
	Stairs
	LightSource
	numTiles
)

// TileDef describes the static properties of a tile type.
type TileDef struct {
	Name        string
	Glyph       rune
	Walkable    bool
	Transparent bool
}

// Defs is the tile registry, indexed by TileType.
var Defs = [numTiles]TileDef{
	Void:       {Name: "void", Glyph: ' '},
	Floor:      {Name: "floor", Glyph: '.', Walkable: true, Transparent: true},
	Wall:       {Name: "wall", Glyph: '#'},
	DoorClosed: {Name: "closed door", Glyph: '+', Transparent: false},
	DoorOpen:   {Name: "open door", Glyph: '\'', Walkable: true, Transparent: true},
	Stairs:     {Name: "stairs", Glyph: '>', Walkable: true, Transparent: true},
	// LightSource is a free-standing lamp or brazier: solid, but you can see past it.
	LightSource: {Name: "light source", Glyph: '*', Transparent: true},
}

func (t TileType) Def() TileDef      { return Defs[t] }
func (t TileType) Walkable() bool    { return Defs[t].Walkable }
func (t TileType) Transparent() bool { return Defs[t].Transparent }
