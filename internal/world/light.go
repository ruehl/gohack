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
	Glyph  rune // used in map files and on screen
	Color  RGB
	Radius int // tiles; light fades out towards this distance
	Anim   Animation
	On     bool // whether lights of this kind start lit
}

// LightTemplates maps template IDs to light definitions. Glyphs must be unique
// and must not clash with other map glyphs.
var LightTemplates = map[string]LightDef{
	"torch":   {Name: "torch", Glyph: '*', Color: RGB{255, 160, 60}, Radius: 6, Anim: AnimFlicker, On: true},
	"crystal": {Name: "glowing crystal", Glyph: '&', Color: RGB{90, 160, 255}, Radius: 5, Anim: AnimNone, On: true},
}

// lightForGlyph returns the light template whose map glyph is c.
func lightForGlyph(c rune) (LightDef, bool) {
	for _, def := range LightTemplates {
		if def.Glyph == c {
			return def, true
		}
	}
	return LightDef{}, false
}

// Light is a light source placed on a level. On changes during play, like a
// door's tile does.
type Light struct {
	Def *LightDef
	Pos geom.Point
	On  bool
}
