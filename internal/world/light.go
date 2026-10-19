package world

import "gohack/internal/geom"

// RGB is a light color. It is independent of any frontend's color type.
type RGB struct{ R, G, B uint8 }

// Animation is how a light's color changes over time.
type Animation uint8

const (
	AnimNone    Animation = iota
	AnimFlicker           // irregular brightness changes, like a torch flame
)

// LightDef describes a kind of light source.
type LightDef struct {
	Name   string
	Color  RGB
	Radius int // tiles; light fades out towards this distance
	Anim   Animation
	On     bool // whether lights of this kind start lit
}

// LightTemplates maps template IDs (see lightGlyphs) to light definitions.
var LightTemplates = map[string]LightDef{
	"torch": {Name: "torch", Color: RGB{255, 160, 60}, Radius: 6, Anim: AnimFlicker, On: true},
}

// Light is a light source placed on a level. On changes during play, like a
// door's tile does.
type Light struct {
	Def *LightDef
	Pos geom.Point
	On  bool
}
