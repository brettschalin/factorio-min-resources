package tasscript

import (
	"errors"
	"fmt"
	"io"
)

type Prog struct {
	Start     Floatpair
	Locations map[string]Floatpair
	Exprs     []Expr
	taskNum   int
}

func (p *Prog) Write(w io.Writer) error {

	_, err := w.Write([]byte(`local task = {}

local NORTH = defines.direction.north
local SOUTH = defines.direction.south
local EAST = defines.direction.east
local WEST = defines.direction.west

`))
	if err != nil {
		return err
	}

	for _, e := range p.Exprs {
		s, err := e.Eval(p)
		if err != nil {
			return err
		}
		if _, err = w.Write([]byte(s + "\n")); err != nil {
			return err
		}
	}

	_, err = w.Write([]byte(fmt.Sprintf(`task[%d] = {"break"}
return task
`, p.taskNum+1)))
	return err
}

func (p *Prog) Valid() error {
	return errors.New("not implemented")
}

func (a *AST) ToProg() (*Prog, error) {

	p := Prog{
		Locations: make(map[string]Floatpair),
		Exprs:     make([]Expr, 0),
	}

	if len(a.Statements) == 0 {
		return &p, nil
	}

	i := 0

	if a.Statements[0].Type == "START" {
		p.Exprs = append(p.Exprs, &ExprStart{
			Location: Location{
				Values: Floatpair{
					X: a.Statements[0].FloatVals[0],
					Y: a.Statements[0].FloatVals[1],
				},
			},
		})
		i = 1
	} else {
		p.Exprs = append(p.Exprs, &ExprStart{
			Location: Location{
				Values: Floatpair{
					X: 0,
					Y: 0,
				},
			},
		})
	}

	for i < len(a.Statements) {
		s := a.Statements[i]
		if s.Type == "HALT" {
			break
		}

		e, err := a.convExpr(s)
		if err != nil {
			return nil, err
		}
		p.Exprs = append(p.Exprs, e)

		i++
	}

	return &p, nil
}

func (a *AST) convExpr(s *Statement) (Expr, error) {
	switch s.Type {

	case "LOCATION":
		return &LocationExpr{
			Name: s.StrVals[0],
			Location: Location{
				Values: Floatpair{
					X: s.FloatVals[0],
					Y: s.FloatVals[1],
				},
			},
		}, nil
	case "BUILD":
		return &BuildExpr{
			Location:  s.Location.Clone(),
			Item:      s.StrVals[0],
			Direction: s.Direction,
		}, nil
	case "MINE":
		return &MineExpr{
			Location: s.Location.Clone(),
			Amount:   s.IntVals[0],
		}, nil
	case "SPEED":
		return &SpeedExpr{
			Value: s.FloatVals[0],
		}, nil
	case "RECIPE":
		return &RecipeExpr{
			Location: s.Location.Clone(),
			Recipe:   s.StrVals[0],
		}, nil
	case "ROTATE":
		return &RotateExpr{
			Location:  s.Location.Clone(),
			Direction: s.Direction,
		}, nil
	case "PUT":
		return &PutExpr{
			Location:  s.Location.Clone(),
			Item:      s.StrVals[0],
			Inventory: s.StrVals[1],
			Amount:    s.IntVals[0],
		}, nil
	case "TAKE":
		return &TakeExpr{
			Location:  s.Location.Clone(),
			Item:      s.StrVals[0],
			Inventory: s.StrVals[1],
			Amount:    s.IntVals[0],
		}, nil
	case "CRAFT":
		return &CraftExpr{
			Item:   s.StrVals[0],
			Amount: s.IntVals[0],
		}, nil
	case "TECH":
		return &TechExpr{
			Tech: s.StrVals[0],
		}, nil
	case "LAUNCH":
		return &LaunchExpr{
			Location: s.Location.Clone(),
		}, nil
	case "IDLE":
		return &IdleExpr{
			N: s.IntVals[0],
		}, nil
	case "WALK":
		return &WalkExpr{
			Location: s.Location.Clone(),
		}, nil
	case "LOOP":
		e := &LoopExpr{
			N: s.IntVals[0],
		}

		for _, sub := range s.Body {
			ex, err := a.convExpr(sub)
			if err != nil {
				return nil, err
			}
			e.Exprs = append(e.Exprs, ex)
		}

		return e, nil

	default:
		return nil, fmt.Errorf(`invalid or unknown command %q`, s.Type)
	}
}
