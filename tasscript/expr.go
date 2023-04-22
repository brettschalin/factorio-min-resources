package tasscript

import (
	"errors"
	"fmt"
	"strings"
)

type Expr interface {
	Eval(p *Prog) (string, error)
}

type ExprStart struct {
	Location Location
}

func (e *ExprStart) Eval(p *Prog) (string, error) {
	p.Start = e.Location.Values
	p.taskNum++
	return fmt.Sprintf(`task[%d] = {"move", {%f, %f}}`, p.taskNum, p.Start.X, p.Start.Y), nil
}

type LocationExpr struct {
	Name     string
	Location Location
}

func (e *LocationExpr) Eval(p *Prog) (string, error) {
	v := e.Location.Values
	if _, ok := p.Locations[e.Name]; ok {
		return "", errors.New("duplicate location " + e.Name)
	}
	p.Locations[e.Name] = v
	return fmt.Sprintf(`local %s = {%f, %f}`, e.Name, v.X, v.Y), nil
}

type BuildExpr struct {
	Location  Location
	Item      string
	Direction direction
}

func (e *BuildExpr) Eval(p *Prog) (string, error) {
	p.buildingDirections[e.Location] = e.Direction
	p.taskNum++
	return fmt.Sprintf(`task[%d] = {"build", %s, %q, %s}`, p.taskNum, e.Location.String(), e.Item, e.Direction), nil
}

type MineExpr struct {
	Location Location
	Amount   int
}

func (e *MineExpr) Eval(p *Prog) (string, error) {
	delete(p.buildingDirections, e.Location)
	l := e.Location.String()
	s := strings.Builder{}
	for i := e.Amount; i > 0; i-- {
		p.taskNum++
		s.WriteString(fmt.Sprintf(`task[%d] = {"mine", %s}`, p.taskNum, l))
		if i != 1 {
			s.WriteRune('\n')
		}
	}
	return s.String(), nil
}

type SpeedExpr struct {
	Value float64
}

func (e *SpeedExpr) Eval(p *Prog) (string, error) {
	p.taskNum++
	return fmt.Sprintf(`task[%d] = {"speed", %f}`, p.taskNum, e.Value), nil
}

type RecipeExpr struct {
	Location Location
	Recipe   string
}

func (e *RecipeExpr) Eval(p *Prog) (string, error) {
	p.taskNum++
	return fmt.Sprintf(`task[%d] = {"recipe", %s, %s}`, p.taskNum, e.Location.String(), e.Recipe), nil
}

type RotateExpr struct {
	Location  Location
	Direction direction
}

var rotations = map[direction]map[direction]string{
	North: {
		East:  "cw",
		South: "180",
		West:  "ccw",
	},
	East: {
		North: "ccw",
		South: "cw",
		West:  "180",
	},
	South: {
		North: "180",
		East:  "ccw",
		West:  "cw",
	},
	West: {
		North: "cw",
		East:  "180",
		South: "ccw",
	},
}

func (e *RotateExpr) Eval(p *Prog) (string, error) {

	dir, ok := p.buildingDirections[e.Location]
	if !ok {
		return "", errors.New("no building at location " + e.Location.String())
	}

	rot := rotations[dir][e.Direction]
	if rot == "" {
		// building is already facing this direction!
		return fmt.Sprintf(`-- NOOP ROTATE %s %q`, e.Location, e.Direction.String()), nil
	}

	p.taskNum++
	return fmt.Sprintf(`task[%d] = {"rotate", %s, %q}`, p.taskNum, e.Location, rot), nil
}

type PutExpr struct {
	Location  Location
	Item      string
	Inventory string
	Amount    int
}

func (e *PutExpr) Eval(p *Prog) (string, error) {
	p.taskNum++
	return fmt.Sprintf(`task[%d] = {"put", %s, %q, %d, defines.inventory.%s}`, p.taskNum, e.Location, e.Item, e.Amount, e.Inventory), nil
}

type TakeExpr struct {
	Location  Location
	Item      string
	Inventory string
	Amount    int
}

func (e *TakeExpr) Eval(p *Prog) (string, error) {
	p.taskNum++
	return fmt.Sprintf(`task[%d] = {"take", %s, %q, %d, defines.inventory.%s}`, p.taskNum, e.Location, e.Item, e.Amount, e.Inventory), nil
}

type CraftExpr struct {
	Item   string
	Amount int
}

func (e *CraftExpr) Eval(p *Prog) (string, error) {
	p.taskNum++
	return fmt.Sprintf(`task[%d] = {"craft", %d, %q}`, p.taskNum, e.Amount, e.Item), nil
}

type TechExpr struct {
	Tech string
}

func (e *TechExpr) Eval(p *Prog) (string, error) {
	p.taskNum++
	return fmt.Sprintf(`task[%d] = {"tech", %q}`, p.taskNum, e.Tech), nil
}

type LaunchExpr struct {
	Location Location
}

func (e *LaunchExpr) Eval(p *Prog) (string, error) {
	p.taskNum++
	return fmt.Sprintf(`task[%d] = {"launch", %s}`, p.taskNum, e.Location.String()), nil
}

type IdleExpr struct {
	N int
}

func (e *IdleExpr) Eval(p *Prog) (string, error) {
	// TODO: AnyPctTAS uses `idle` and `nop`. I'm not sure what the difference is yet
	s := strings.Builder{}
	for i := e.N; i > 0; i-- {
		p.taskNum++
		s.WriteString(fmt.Sprintf(`task[%d] = {"nop"}`, p.taskNum))
		if i != 1 {
			s.WriteRune('\n')
		}
	}
	return s.String(), nil
}

type WalkExpr struct {
	Location Location
}

func (e *WalkExpr) Eval(p *Prog) (string, error) {
	// TODO: AnyPctTAS uses `walk` and `move`. I'm not sure what the difference is yet
	p.taskNum++
	return fmt.Sprintf(`task[%d] = {"walk", %s}`, p.taskNum, e.Location.String()), nil
}

type LoopExpr struct {
	N     int
	Exprs []Expr
}

func (e *LoopExpr) Eval(p *Prog) (string, error) {
	s := strings.Builder{}
	for n := 0; n < e.N; n++ {
		for _, sub := range e.Exprs {
			line, err := sub.Eval(p)
			if err != nil {
				return "", err
			}
			s.WriteString(line)
			s.WriteRune('\n')
		}
	}
	return s.String(), nil
}
