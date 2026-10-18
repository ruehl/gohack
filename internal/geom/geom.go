// Package geom holds basic 2D integer geometry. Y grows downwards.
package geom

// Point is a tile coordinate.
type Point struct{ X, Y int }

// Add returns p translated by q.
func (p Point) Add(q Point) Point { return Point{p.X + q.X, p.Y + q.Y} }

// Dir is one of eight compass directions.
type Dir uint8

const (
	N Dir = iota
	E
	S
	W
	NE
	SE
	SW
	NW
)

var deltas = [...]Point{
	N: {0, -1}, E: {1, 0}, S: {0, 1}, W: {-1, 0},
	NE: {1, -1}, SE: {1, 1}, SW: {-1, 1}, NW: {-1, -1},
}

// Delta returns the unit offset of d.
func (d Dir) Delta() Point { return deltas[d] }

// AllDirs lists all eight directions.
var AllDirs = []Dir{N, E, S, W, NE, SE, SW, NW}
