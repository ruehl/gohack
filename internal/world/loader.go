package world

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"gohack/internal/geom"
)

// spawnGlyphs maps map-file glyphs to NPC template IDs. The tile below is Floor.
var spawnGlyphs = map[rune]string{
	'g': "goblin",
}

// lightGlyphs maps map-file glyphs to light template IDs in LightTemplates.
var lightGlyphs = map[rune]string{
	'*': "torch",
}

const startGlyph = '@'

// portalDecl is a "portal:" header line, resolved after all levels are parsed.
type portalDecl struct {
	from     geom.Point
	toLevel  LevelID
	to       geom.Point
	srcLevel LevelID
}

// Load reads every *.txt level file in dir and links their portals.
func Load(fsys fs.FS, dir string) (*World, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}
	w := &World{Levels: map[LevelID]*Level{}}
	var decls []portalDecl
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".txt") {
			continue
		}
		src, err := fs.ReadFile(fsys, dir+"/"+e.Name())
		if err != nil {
			return nil, err
		}
		l, ds, err := ParseLevel(string(src))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		if _, dup := w.Levels[l.ID]; dup {
			return nil, fmt.Errorf("%s: duplicate level id %q", e.Name(), l.ID)
		}
		w.Levels[l.ID] = l
		decls = append(decls, ds...)
	}
	for id := range w.Levels {
		w.Order = append(w.Order, id)
	}
	sort.Slice(w.Order, func(i, j int) bool { return w.Order[i] < w.Order[j] })

	for _, d := range decls {
		if err := w.link(d); err != nil {
			return nil, err
		}
	}
	for _, id := range w.Order {
		if w.Levels[id].HasStart {
			if w.Start != "" {
				return nil, fmt.Errorf("player start defined in both %q and %q", w.Start, id)
			}
			w.Start = id
		}
	}
	if w.Start == "" {
		return nil, fmt.Errorf("no player start ('%c') in any level", startGlyph)
	}
	return w, nil
}

// link creates both ends of a bidirectional stair.
func (w *World) link(d portalDecl) error {
	a, b := w.Levels[d.srcLevel], w.Levels[d.toLevel]
	if b == nil {
		return fmt.Errorf("level %q: portal at %v targets unknown level %q", d.srcLevel, d.from, d.toLevel)
	}
	if a.At(d.from) != Stairs {
		return fmt.Errorf("level %q: portal at %v is not on a stairs tile", a.ID, d.from)
	}
	if b.At(d.to) != Stairs {
		return fmt.Errorf("level %q: portal target %v in %q is not a stairs tile", a.ID, d.to, b.ID)
	}
	if err := addPortal(a, d.from, Portal{PortalStairs, b.ID, d.to}); err != nil {
		return err
	}
	return addPortal(b, d.to, Portal{PortalStairs, a.ID, d.from})
}

func addPortal(l *Level, p geom.Point, portal Portal) error {
	if old, ok := l.Portals[p]; ok && old != portal {
		return fmt.Errorf("level %q: conflicting portals at %v", l.ID, p)
	}
	l.Portals[p] = portal
	return nil
}

// ParseLevel parses one level file: "key: value" headers, a "---" line, then the map.
func ParseLevel(src string) (*Level, []portalDecl, error) {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	head, body, ok := strings.Cut(src, "\n---\n")
	if !ok {
		return nil, nil, fmt.Errorf("missing '---' separator")
	}
	var id LevelID
	var name string
	var portalLines []string
	for _, line := range strings.Split(head, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			return nil, nil, fmt.Errorf("bad header line %q", line)
		}
		v = strings.TrimSpace(v)
		switch strings.TrimSpace(k) {
		case "id":
			id = LevelID(v)
		case "name":
			name = v
		case "portal":
			portalLines = append(portalLines, v)
		default:
			return nil, nil, fmt.Errorf("unknown header %q", k)
		}
	}
	if id == "" {
		return nil, nil, fmt.Errorf("missing id")
	}
	if name == "" {
		name = string(id)
	}

	rows := strings.Split(strings.TrimRight(body, "\n"), "\n")
	w := 0
	for _, r := range rows {
		if n := len([]rune(r)); n > w {
			w = n
		}
	}
	h := len(rows)
	if w == 0 || h == 0 {
		return nil, nil, fmt.Errorf("empty map")
	}
	if w > MaxLevelSize || h > MaxLevelSize {
		return nil, nil, fmt.Errorf("map is %dx%d, maximum is %dx%d", w, h, MaxLevelSize, MaxLevelSize)
	}

	l := NewLevel(id, name, w, h)
	for y, r := range rows {
		for x, c := range []rune(r) {
			p := geom.Point{X: x, Y: y}
			switch c {
			case ' ':
			case '#':
				l.Set(p, Wall)
			case '.':
				l.Set(p, Floor)
			case '+':
				l.Set(p, DoorClosed)
			case '\'':
				l.Set(p, DoorOpen)
			case '>':
				l.Set(p, Stairs)
			case startGlyph:
				if l.HasStart {
					return nil, nil, fmt.Errorf("more than one player start")
				}
				l.Set(p, Floor)
				l.Start, l.HasStart = p, true
			default:
				if id, ok := lightGlyphs[c]; ok {
					def, ok := LightTemplates[id]
					if !ok {
						return nil, nil, fmt.Errorf("unknown light template %q at %v", id, p)
					}
					l.Set(p, LightSource)
					l.Lights = append(l.Lights, &Light{Def: &def, Pos: p, On: def.On})
					continue
				}
				tpl, ok := spawnGlyphs[c]
				if !ok {
					return nil, nil, fmt.Errorf("unknown map glyph %q at %v", c, p)
				}
				l.Set(p, Floor)
				l.Spawns = append(l.Spawns, Spawn{Template: tpl, Pos: p})
			}
		}
	}

	var decls []portalDecl
	for _, pl := range portalLines {
		d, err := parsePortal(pl)
		if err != nil {
			return nil, nil, err
		}
		d.srcLevel = id
		decls = append(decls, d)
	}
	return l, decls, nil
}

// parsePortal parses "x,y -> level:x,y".
func parsePortal(s string) (portalDecl, error) {
	var d portalDecl
	from, to, ok := strings.Cut(s, "->")
	if !ok {
		return d, fmt.Errorf("bad portal %q", s)
	}
	if _, err := fmt.Sscanf(strings.TrimSpace(from), "%d,%d", &d.from.X, &d.from.Y); err != nil {
		return d, fmt.Errorf("bad portal source in %q", s)
	}
	lvl, pos, ok := strings.Cut(strings.TrimSpace(to), ":")
	if !ok {
		return d, fmt.Errorf("bad portal target in %q", s)
	}
	if _, err := fmt.Sscanf(pos, "%d,%d", &d.to.X, &d.to.Y); err != nil {
		return d, fmt.Errorf("bad portal target position in %q", s)
	}
	d.toLevel = LevelID(strings.TrimSpace(lvl))
	return d, nil
}
