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

	slotLabel = map[entity.Slot]string{
		entity.SlotHead:     "Head",
		entity.SlotNeck:     "Neck",
		entity.SlotRing1:    "Ring 1",
		entity.SlotRing2:    "Ring 2",
		entity.SlotBody:     "Body",
		entity.SlotHands:    "Hands",
		entity.SlotLegs:     "Legs",
		entity.SlotFeet:     "Feet",
		entity.SlotMainHand: "Main hand",
		entity.SlotOffHand:  "Off hand",
	}

	tileStyle = map[world.TileType]tcell.Style{
		world.Floor:      tcell.StyleDefault.Foreground(tcell.NewRGBColor(130, 130, 130)),
		world.Wall:       tcell.StyleDefault.Foreground(tcell.NewRGBColor(200, 200, 200)),
		world.DoorClosed: tcell.StyleDefault.Foreground(tcell.NewRGBColor(190, 130, 60)),
		world.DoorOpen:   tcell.StyleDefault.Foreground(tcell.NewRGBColor(190, 130, 60)),
		world.Stairs:     tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true),
	}
)

// cellGlyph and cellStyle say how a tile looks. A light source looks like its
// own kind of light: dull in its own color when off, glowing once litLight
// applies its light to it.
func cellGlyph(c game.Cell) rune {
	if c.Source != nil {
		return c.Source.Glyph
	}
	return world.Defs[c.Tile].Glyph
}

func cellStyle(c game.Cell) tcell.Style {
	if c.Source != nil {
		col := c.Source.Color
		dim := func(v uint8) int32 { return int32(v) * 6 / 10 }
		return tcell.StyleDefault.Foreground(tcell.NewRGBColor(dim(col.R), dim(col.G), dim(col.B))).Bold(true)
	}
	return tileStyle[c.Tile]
}

// ambient is how bright a visible tile is with no light on it, as a fraction of
// its normal color. Light adds to it per color channel.
const ambient = 0.7

// litStyle returns style with its foreground scaled by ambient light plus the
// tile's light color.
func litStyle(style tcell.Style, light world.RGB) tcell.Style {
	fg, _, _ := style.Decompose()
	r, g, b := fg.RGB()
	scale := func(c int32, l uint8) int32 {
		return int32(min(255, float64(c)*(ambient+float64(l)/255)))
	}
	return style.Foreground(tcell.NewRGBColor(scale(r, light.R), scale(g, light.G), scale(b, light.B)))
}

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
				s.SetContent(x, y, cellGlyph(c), nil, litStyle(cellStyle(c), c.Light))
			case c.Vis == game.Remembered:
				s.SetContent(x, y, cellGlyph(c), nil, styleRemember)
			}
		}
	}

	if in.GearOpen() {
		drawGear(s, v, in, mapW, mapH)
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
	switch {
	case in.Attacking():
		put(styleWarn, "Attack which direction?")
	case in.Toggling():
		put(styleWarn, "Light which direction?")
	default:
		put(styleDim, "WASD/arrows move")
		put(styleDim, "F attack  E stairs")
		put(styleDim, "L light  I gear")
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

// drawGear covers the map area with the gear screen: the slot list, or the
// items that fit the chosen slot.
func drawGear(s tcell.Screen, v *game.View, in *Input, mapW, mapH int) {
	for y := 0; y < mapH; y++ {
		for x := 0; x < mapW; x++ {
			s.SetContent(x, y, ' ', nil, styleText)
		}
	}
	m := &in.gear
	var title, hint string
	var rows []string
	if m.picking {
		title, hint = "Equip: "+slotLabel[m.slot], "Enter equip  Esc back  I close"
		for _, c := range choices(&v.Gear, m.slot) {
			row := c.Name
			if c.Item >= 0 {
				if w := v.Gear.Items[c.Item].Worn; w != entity.SlotNone {
					row += " (worn: " + slotLabel[w] + ")"
				}
			}
			rows = append(rows, row)
		}
		if len(rows) == 0 {
			rows = append(rows, "nothing fits here")
		}
	} else {
		title, hint = "Equipment", "Enter choose slot  Esc close"
		for _, sl := range v.Gear.Slots {
			item := sl.Item
			if item == "" {
				item = "-"
			}
			rows = append(rows, fmt.Sprintf("%-11s%s", slotLabel[sl.Slot], item))
		}
	}
	putString(s, 2, 1, mapW-2, styleHeading, title)
	for i, r := range rows {
		style := styleText
		if i == m.cursor {
			style = style.Reverse(true)
		}
		putString(s, 2, 3+i, mapW-2, style, r)
	}
	putString(s, 2, 4+len(rows), mapW-2, styleDim, hint)
}

func drawDeath(s tcell.Screen, sw, sh int) {
	lines := []string{"                      ", "       YOU DIED       ", "   Press any key...   ", "                      "}
	for i, l := range lines {
		putString(s, (sw-len(l))/2, sh/2-2+i, len(l), styleWarn.Reverse(true), l)
	}
}
