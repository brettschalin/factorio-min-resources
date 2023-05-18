// package geo defines simple geometric shapes.
// It is adapted from the shapes in the standard library's "image" package,
// with added support for floating point values
package geo

import (
	"encoding/json"
	"math"
)

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (p Point) In(r Rectangle) bool {
	return r.TopLeft.X <= p.X && p.X <= r.BottomRight.X &&
		r.TopLeft.Y <= p.Y && p.Y <= r.BottomRight.Y
}

func (p Point) Add(o Point) Point {
	return Point{
		X: p.X + o.X,
		Y: p.Y + o.Y,
	}
}

func (p Point) Sub(o Point) Point {
	return Point{
		X: p.X - o.X,
		Y: p.Y - o.Y,
	}
}

func (p Point) Mul(x float64) Point {
	return Point{
		X: p.X * x,
		Y: p.Y * x,
	}
}

// Distance returns the straight-line distance between two points
func (p Point) Distance(o Point) float64 {
	d := o.Sub(p)
	return d.Magnitude()
}

// Magnitude returns the stright-line distance between p and the origin
func (p Point) Magnitude() float64 {
	return math.Hypot(p.X, p.Y)
}

// ClosestWithin returns the closest point that is
// - within d units of p
// - in the direction of o
// - outside the box (if any) defined by placing colBox[0] at p
func (p Point) ClosestWithin(o Point, d float64, colBox ...Rectangle) Point {

	// sanity check
	if d == 0 {
		return p
	}

	if len(colBox) > 0 {
		box := colBox[1].Add(p)
		other := box.ClosestTo(o)
		if o.In(box) {
			// go to the edge and a bit further beyond
			return other.Add(other.Sub(o).normalize(d))
		}
		o = other
	}

	// vector from p to o, normalized to the right distance
	v := o.Sub(p).normalize(d)

	// and add it to p
	return p.Add(v)
}

// treats p as a vector from the origin and ensures it has magnitude l
func (p Point) normalize(l float64) Point {
	return p.Mul(math.Sqrt(l) / p.Magnitude())
}

// PathDistance returns the distance between two points by
// walking between them. This uses the algorithm from control.lua
// in the mod
func (p Point) PathDistance(o Point) float64 {

	distance := float64(0)

	delta := o.Sub(p)

	// gets us close enough without having to deal with how floating point numbers
	// are actually handled
	const threshhold = 1e-5

	// the mod cares about the direction traveled, this calculation does not.
	// Keep us in one quadrant
	x, y := math.Abs(delta.X), math.Abs(delta.Y)

	switch {
	case x > threshhold && y > threshhold:
		// travel diagonally until we're at a horizontal or vertical line from the destination,
		// then travel along that line
		m := math.Min(x, y)
		x -= m
		y -= m
		distance = (m * math.Sqrt2) + x + y
	case x < threshhold && y > threshhold:
		// vertical
		distance = y
	case x > threshhold && y < threshhold:
		// horizontal
		distance = x
	default:
		// no distance to travel
		distance = 0
	}

	return distance
}

type Rectangle struct {
	TopLeft     Point
	BottomRight Point
}

// Add translates the Rectangle r by the distance in Point p
func (r Rectangle) Add(p Point) Rectangle {
	return Rectangle{
		TopLeft:     r.TopLeft.Add(p),
		BottomRight: r.BottomRight.Add(p),
	}
}

// Dx returns the width of the Rectangle
func (r Rectangle) Dx() float64 {
	return r.BottomRight.X - r.TopLeft.X
}

// Dy returns the height of the Rectangle
func (r Rectangle) Dy() float64 {
	return r.BottomRight.Y - r.TopLeft.Y
}

// Overlap determines if two Rectangles overlap
func (r Rectangle) Overlap(s Rectangle) bool {
	return r.TopLeft.X < s.BottomRight.X && s.TopLeft.X < r.BottomRight.X &&
		r.TopLeft.Y < s.BottomRight.Y && s.BottomRight.Y < r.BottomRight.Y
}

// ClosestTo returns a Point on the Rectangle's edge that is closest to the provided Point
func (r Rectangle) ClosestTo(p Point) Point {

	var retX, retY = p.X, p.Y
	if p.In(r) {

		// TODO: these checks will ensure that one of r's corners is returned.
		// This is obviously not correct

		if math.Abs(p.Y-r.TopLeft.Y) > math.Abs(p.Y-r.BottomRight.Y) {
			retY = r.TopLeft.Y
		} else {
			retY = r.BottomRight.Y
		}

		if math.Abs(p.X-r.TopLeft.X) > math.Abs(p.X-r.BottomRight.X) {
			retX = r.TopLeft.X
		} else {
			retX = r.BottomRight.X
		}

		return Point{
			X: retX,
			Y: retY,
		}
	}

	if p.Y <= r.TopLeft.Y {
		retY = r.TopLeft.Y
	} else if p.Y <= r.BottomRight.Y {
		retY = r.BottomRight.Y
	}

	if p.X <= r.TopLeft.X {
		retX = r.TopLeft.X
	} else if p.X >= r.BottomRight.X {
		retX = r.BottomRight.X
	}

	return Point{
		X: retX,
		Y: retY,
	}
}

// Unmarshal implements the json.Unmarshaler interface
func (r *Rectangle) UnmarshalJSON(b []byte) error {
	var rect [][]float64
	if err := json.Unmarshal(b, &rect); err != nil {
		return err
	}

	r.TopLeft = Point{
		X: rect[0][0],
		Y: rect[0][1],
	}
	r.BottomRight = Point{
		X: rect[1][0],
		Y: rect[1][1],
	}

	return nil
}
