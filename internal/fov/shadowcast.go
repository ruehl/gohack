// Package fov computes line of sight with symmetric shadowcasting
// (after Albert Ford, "Symmetric Shadowcasting"). It works in exact integer
// arithmetic, touches only cells inside the radius, and allocates nothing
// beyond the closures passed in.
package fov

// slope is the fraction n/d with d > 0.
type slope struct{ n, d int }

// Compute marks every cell visible from (ox, oy) within radius.
// isOpaque must return true for blocking cells and for anything out of bounds.
// Opaque cells that border visible floor are marked too, so walls show up.
func Compute(ox, oy, radius int, isOpaque func(x, y int) bool, mark func(x, y int)) {
	mark(ox, oy)
	r2 := radius * radius
	for q := 0; q < 4; q++ {
		s := scanner{ox: ox, oy: oy, q: q, radius: radius, r2: r2, isOpaque: isOpaque, mark: mark}
		s.scan(1, slope{-1, 1}, slope{1, 1})
	}
}

type scanner struct {
	ox, oy   int
	q        int // 0 north, 1 east, 2 south, 3 west
	radius   int
	r2       int
	isOpaque func(x, y int) bool
	mark     func(x, y int)
}

// transform maps quadrant-local (depth, col) to map coordinates.
func (s *scanner) transform(depth, col int) (int, int) {
	switch s.q {
	case 0:
		return s.ox + col, s.oy - depth
	case 1:
		return s.ox + depth, s.oy + col
	case 2:
		return s.ox + col, s.oy + depth
	default:
		return s.ox - depth, s.oy + col
	}
}

func floorDiv(a, b int) int { // b > 0
	q := a / b
	if a%b != 0 && a < 0 {
		q--
	}
	return q
}

func ceilDiv(a, b int) int { return -floorDiv(-a, b) }

// scan processes one row (distance `depth` from the origin) between the two slopes.
func (s *scanner) scan(depth int, start, end slope) {
	if depth > s.radius {
		return
	}
	// round_ties_up(depth*start), round_ties_down(depth*end)
	minCol := floorDiv(2*depth*start.n+start.d, 2*start.d)
	maxCol := ceilDiv(2*depth*end.n-end.d, 2*end.d)

	prevSet, prevWall := false, false
	for col := minCol; col <= maxCol; col++ {
		x, y := s.transform(depth, col)
		wall := s.isOpaque(x, y)

		// symmetric: the cell center lies within [start, end] of this row
		sym := col*start.d >= depth*start.n && col*end.d <= depth*end.n
		if (wall || sym) && col*col+depth*depth <= s.r2 {
			s.mark(x, y)
		}
		if prevSet && prevWall && !wall {
			start = slope{2*col - 1, 2 * depth}
		}
		if prevSet && !prevWall && wall {
			s.scan(depth+1, start, slope{2*col - 1, 2 * depth})
		}
		prevSet, prevWall = true, wall
	}
	if prevSet && !prevWall {
		s.scan(depth+1, start, end)
	}
}
