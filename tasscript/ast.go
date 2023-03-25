package tasscript

import (
	"fmt"
	"io"
	"os"
	"strings"
)

//go:generate goyacc -o tas.go -p "tas" tas.y

// Uncomment for better debugging
// func init() {
// 	tasDebug = 3
// 	tasErrorVerbose = true
// }

type AST struct {
	Statements Statements
}

func (a *AST) Clone() *AST {

	other := AST{
		Statements: a.Statements.Clone(),
	}

	return &other
}

func (a *AST) Add(s *Statement) {
	a.Statements = append(a.Statements, s.Clone())
}

func (a *AST) String() string {
	return a.Statements.String()
}

type Statements []*Statement

func (st Statements) Clone() Statements {
	other := make(Statements, len(st))
	for i, s := range st {
		other[i] = s.Clone()
	}
	return other
}

func (st Statements) String() string {
	return st.string(0)
}

func (st Statements) string(depth int) string {
	var str string
	for _, s := range st {
		if s == nil || s.Type == "" {
			continue
		}
		var b strings.Builder
		b.WriteString(strings.Repeat("    ", depth))
		b.WriteString(s.Type)
		if len(s.IntVals) > 0 {
			b.WriteString(", ints: ")
			b.WriteString(fmt.Sprintf("%v", s.IntVals))
		}
		if len(s.FloatVals) > 0 {
			b.WriteString(", floats: ")
			b.WriteString(fmt.Sprintf("%v", s.FloatVals))
		}
		if len(s.StrVals) > 0 {
			b.WriteString(", strings: ")
			b.WriteString(fmt.Sprintf("%v", s.StrVals))
		}
		b.WriteString(", location: ")
		b.WriteString(fmt.Sprintf("%#v", s.Location))

		if s.Direction > 0 {
			b.WriteString(", direction: ")
			b.WriteString(s.Direction.String())
		}
		if len(s.Body) > 0 {
			b.WriteRune('\n')
			b.WriteString(s.Body.string(depth + 1))
			str += b.String()
			continue
		}
		b.WriteRune('\n')
		str += b.String()

	}

	return str
}

type direction int

const (
	North direction = iota + 1
	South
	East
	West
)

func (d direction) String() string {
	switch d {
	case North:
		return "NORTH"
	case South:
		return "SOUTH"
	case East:
		return "EAST"
	case West:
		return "WEST"
	}
	return "UNKNOWN"
}

type Statement struct {
	Type      string
	IntVals   []int
	FloatVals []float32
	StrVals   []string
	Location  Location
	Direction direction
	Body      Statements
}

func (s *Statement) Clone() *Statement {
	if s == nil {
		return nil
	}

	other := Statement{
		Type:      s.Type,
		IntVals:   make([]int, len(s.IntVals)),
		FloatVals: make([]float32, len(s.FloatVals)),
		StrVals:   make([]string, len(s.StrVals)),
		Location:  s.Location,
		Direction: s.Direction,
		Body:      s.Body.Clone(),
	}

	copy(other.IntVals, s.IntVals)
	copy(other.FloatVals, s.FloatVals)
	copy(other.StrVals, s.StrVals)

	return &other
}

type Floatpair struct {
	X, Y float32
}

type Location struct {
	Named  string
	Values Floatpair
}

func (l Location) String() string {
	if l.Named != "" {
		return l.Named
	} else {
		return fmt.Sprintf(`{%f, %f}`, l.Values.X, l.Values.Y)
	}
}

func (l Location) Clone() Location {
	return Location{
		Named: l.Named,
		Values: Floatpair{
			X: l.Values.X,
			Y: l.Values.Y,
		},
	}
}

type loopBegin struct {
	n int
}

var keywords = map[string]int{
	"BUILD":    BUILD,
	"CRAFT":    CRAFT,
	"IDLE":     IDLE,
	"LAUNCH":   LAUNCH,
	"MINE":     MINE,
	"PUT":      PUT,
	"RECIPE":   RECIPE,
	"ROTATE":   ROTATE,
	"SPEED":    SPEED,
	"TAKE":     TAKE,
	"TECH":     TECH,
	"WALK":     WALK,
	"START":    START,
	"LOCATION": LOCATION,
	"HALT":     HALT,
	"LOOP":     LOOP,
	"ENDLOOP":  ENDLOOP,
	"NORTH":    NORTH,
	"SOUTH":    SOUTH,
	"EAST":     EAST,
	"WEST":     WEST,
}

// Holds the result of the parse. Absolutely not threadsafe in the slightest but I don't need it to be so I don't care
var tasParseResult *AST

func Read(r io.Reader) (*AST, error) {
	in, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	n := tasParse(&exprLex{input: in})

	if n != 0 {
		return nil, fmt.Errorf(`parse returned code: %d`, n)
	}

	return tasParseResult.Clone(), nil
}

func ReadFromFile(f string) (*AST, error) {
	file, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return Read(file)
}
