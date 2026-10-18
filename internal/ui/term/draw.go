package term

import (
	"fmt"

	"github.com/gdamore/tcell/v2"

	"gohack/internal/entity"
	"gohack/internal/game"
	"gohack/internal/world"
)

const hudWidth = 24

var (
	styleText     = tcell.StyleDefault.Foreground(tcell.ColorSilver)
	styleHeading  = tcell.StyleDefault.Foreground(tcell.ColorWhite).Bold(true)
	styleDim      = tcell.StyleDefault.Foreground(tcell.ColorGray)
	styleRemember = tcell.StyleDefault.Foreground(tcell.NewRGBColor(70, 70, 80))
	styleWarn     = tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)

	tileStyle = map[world.TileType]tcell.Style{
		world.Floor:      tcell.StyleDefault.Foreground(tcell.NewRGBColor(130, 130, 130)),
		world.Wall:       tcell.StyleDefault.Foreground(tcell.NewRGBColor(200, 200, 200)),
		world.DoorClosed: tcell.StyleDefault.Foreground(tcell.NewRGBColor(190, 130, 60)),
		world.DoorOpen:   tcell.StyleDefault.Foreground(tcell.NewRGBColor(190, 130, 60)),
		world.Stairs:     tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true),
	}
)

func actorStyle(a *game.ActorView) tcell.Style {
	switch {
	case a.IsPlayer:
		return tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
	case a.Disposition == entity.Hostile:
		return tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)
	case a.Disposition == entity.Friendly:
		return tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true)
	}
	return tcell.StyleDefault.Foreground(tcell.ColorOlive).Bold(true)
}

// viewportSize returns the map area size for a screen of w×h.
func viewportSize(w, h int) (int, int) {
	return max(0, w-hudWidth-1), max(0, h-game.MessageLines)
}

func putString(s tcell.Screen, x, y, maxW int, style tcell.Style, text string) {
	n := 0
	for _, r := range text {
		if n >= maxW {
			return
		}
		s.SetContent(x+n, y, r, nil, style)
		n++
	}
}

func draw(s tcell.Screen, v *game.View, in *Input) {
	s.Clear()
	sw, sh := s.Size()
	mapW, mapH := viewportSize(sw, sh)

	for y := 0; y < v.H; y++ {
		for x := 0; x < v.W; x++ {
			c := v.At(x, y)
			switch {
			case c.Actor != nil:
				s.SetContent(x, y, c.Actor.Glyph, nil, actorStyle(c.Actor))
			case c.Vis == game.Visible:
				s.SetContent(x, y, world.Defs[c.Tile].Glyph, nil, tileStyle[c.Tile])
			case c.Vis == game.Remembered:
				s.SetContent(x, y, world.Defs[c.Tile].Glyph, nil, styleRemember)
			}
		}
	}

	// HUD
	hx, h := mapW+1, v.HUD
	for y := 0; y < mapH; y++ {
		s.SetContent(mapW, y, '│', nil, styleDim)
	}
	line := 0
	put := func(style tcell.Style, format string, args ...any) {
		putString(s, hx, line, hudWidth, style, fmt.Sprintf(format, args...))
		line++
	}
	put(styleHeading, "%s", h.Name)
	put(styleText, "%s %s", h.Race, h.Class)
	put(styleDim, "%s", h.Level)
	put(styleDim, "Turn %d", h.Turn)
	line++
	hpStyle := styleText
	if h.HP*3 <= h.MaxHP {
		hpStyle = styleWarn
	}
	put(hpStyle, "HP  %d/%d", h.HP, h.MaxHP)
	put(styleText, "AC  %d", h.ArmorClass)
	line++
	put(styleText, "STA %d", h.Attrs.Stamina)
	put(styleText, "STR %d", h.Attrs.Strength)
	put(styleText, "DEX %d", h.Attrs.Dexterity)
	put(styleText, "INT %d", h.Attrs.Intelligence)
	line++
	if in.Attacking() {
		put(styleWarn, "Attack which direction?")
	} else {
		put(styleDim, "WASD/arrows move")
		put(styleDim, "F attack  E stairs")
		put(styleDim, "Q quit")
	}

	// message log
	for i, m := range v.Messages {
		putString(s, 0, mapH+i, sw, styleText, m)
	}

	if v.Dead {
		drawDeath(s, sw, sh)
	}
	s.Show()
}

func drawDeath(s tcell.Screen, sw, sh int) {
	lines := []string{"                      ", "       YOU DIED       ", "   Press any key...   ", "                      "}
	for i, l := range lines {
		putString(s, (sw-len(l))/2, sh/2-2+i, len(l), styleWarn.Reverse(true), l)
	}
}
