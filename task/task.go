package task

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/brettschalin/factorio-min-resources/calc"
	"github.com/brettschalin/factorio-min-resources/constants"
	"github.com/brettschalin/factorio-min-resources/data"
	"github.com/brettschalin/factorio-min-resources/geo"
)

type TaskType int

const (
	TaskUnknown TaskType = iota // for spacing
	TaskWalk
	TaskWait
	TaskCraft
	TaskHandcraft
	TaskBuild
	TaskTake
	TaskPut
	TaskTech
	TaskMine
	TaskLaunch
	TaskMeta // isn't actually a task, just groups them together
)

func (t TaskType) String() string {
	switch t {
	case TaskWalk:
		return "walk"
	case TaskWait:
		return "wait"
	case TaskCraft:
		return "craft"
	case TaskHandcraft:
		return "craft"
	case TaskBuild:
		return "build"
	case TaskTake:
		return "take"
	case TaskPut:
		return "put"
	case TaskTech:
		return "tech"
	case TaskMine:
		return "mine"
	case TaskLaunch:
		return "launch"
	case TaskMeta:
		return "meta"
	default:
		return "unknown"
	}
}

func (t *TaskType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

type Task struct {
	// tasks.lua needs IDs to properly do prerequisite checks.
	// Store them here
	ID string `json:"id,omitempty"`

	Parent *Task `json:"-" diff:"-"`
	Index  int   `json:"index" diff:"-"`

	Type          TaskType `json:"type"`
	Prerequisites []*Task  `json:"prerequisites,omitempty"`

	Tech string `json:"tech,omitempty"`

	// for crafting / mining / transferring
	Item string `json:"item,omitempty"`
	// for handcrafting, this is the number of recipes.
	// in all other cases it's the number of items.
	Amount int `json:"amount,omitempty"`

	Entity    string              `json:"entity,omitempty"`
	Location  *geo.Point          `json:"location,omitempty"`
	Direction constants.Direction `json:"direction,omitempty"`

	Slot constants.Inventory `json:"slot,omitempty"` // inventory slot

	WaitCondition string `json:"wait_condition,omitempty"`
}

func (t *Task) AddPrereq(p *Task) {
	p.Parent = t
	p.Index = len(t.Prerequisites)
	t.Prerequisites = append(t.Prerequisites, p)
}

func (t *Task) Prev() *Task {
	p := t.Parent
	if p == nil {
		return nil
	}
	if t.Index == 0 {
		return p.Prev()
	}
	p = p.Prerequisites[t.Index-1]
	for {
		if len(p.Prerequisites) == 0 {
			return p
		}
		n := p.Prerequisites[len(p.Prerequisites)-1]
		if n == nil {
			return p
		}
		p = n
	}
}

func (t *Task) Prune() {
	if t == nil {
		return
	}

	prereqs := make([]*Task, 0, len(t.Prerequisites))
	i := 0
	for _, p := range t.Prerequisites {
		if p == nil ||
			(p.Type == TaskCraft && p.Amount <= 0) ||
			(p.Type == TaskMine && p.Amount <= 0) ||
			(p.Type == TaskTech && p.Tech == "") {
			continue
		}

		p.Prune()
		p.Index = i
		prereqs = append(prereqs, p)
		i++
	}
	t.Prerequisites = prereqs
}

var ids = map[TaskType]int{}

func (t *Task) GetID() string {
	if t.Type == TaskMeta {
		prev := t.Prev()
		if prev != nil {
			return prev.GetID()
		}
	}
	if t.ID != "" {
		return t.ID
	}

	n := ids[t.Type]
	ids[t.Type]++

	t.ID = fmt.Sprintf("task_%s_%d", t.Type.String(), n)
	return t.ID
}

func NewWalk(location geo.Point) *Task {
	return &Task{
		Type:     TaskWalk,
		Location: &location,
	}
}

func NewCraft(items map[string]int) *Task {

	if len(items) == 0 {
		return &Task{
			Type: TaskUnknown,
		}
	}

	steps := data.Ingredients{}

	// Sort the names so their order is predictable
	itemNames := make([]string, 0, len(items))
	for item := range items {
		itemNames = append(itemNames, item)
	}
	sort.Strings(itemNames)

	for _, item := range itemNames {
		n := items[item]
		s, err := calc.RecipeAllIngredients(item, n)
		if err != nil {
			panic(err)
		}
		steps = append(steps, s...)
	}

	steps.MergeDuplicates()

	var t *Task

	if len(items) == 1 {
		t = &Task{
			Type: TaskCraft,
		}
		for k, v := range items {
			t.Item = k
			t.Amount = v
		}
		steps = steps[:len(steps)-1]

	} else {
		t = &Task{
			Type: TaskMeta,
		}
	}

	for _, s := range steps {
		if calc.BaseItems[s.Name] {
			t.AddPrereq(NewMine(s.Name, s.Amount))
		} else {
			t.AddPrereq(&Task{
				Type:   TaskCraft,
				Item:   s.Name,
				Amount: s.Amount,
			})
		}
	}

	return t
}

func NewWait(entity string, slot constants.Inventory, item string, amount int) *Task {
	return &Task{
		Type:          TaskWait,
		Entity:        entity,
		Slot:          slot,
		Item:          item,
		Amount:        amount,
		WaitCondition: "has_inventory",
	}
}

func NewBuild(entity string, direction constants.Direction) *Task {
	return &Task{
		Type:      TaskBuild,
		Entity:    entity,
		Direction: direction,
	}
}

func NewTransfer(entity string, slot constants.Inventory, item string, amount int, take bool) *Task {
	t := &Task{
		Slot:   slot,
		Item:   item,
		Amount: amount,
		Entity: entity,
	}
	if take {
		t.Type = TaskTake
	} else {
		t.Type = TaskPut
	}
	return t
}

func NewTech(techName string) *Task {

	task := &Task{
		Type: TaskTech,
		Tech: techName,
	}

	tech := data.GetTech(techName)
	for _, p := range tech.Prerequisites {
		task.AddPrereq(NewTech(p))
	}

	cost := map[string]int{}
	for _, ing := range tech.Unit.Ingredients {
		cost[ing.Name] += ing.Amount * tech.Unit.Count
	}

	task.AddPrereq(NewCraft(cost))

	return task
}

func NewMine(resource string, amount int) *Task {

	t := &Task{
		Type:   TaskMine,
		Amount: amount,
	}

	if calc.BaseItems[resource] {
		t.Item = resource
	} else {
		t.Entity = resource
	}

	return t
}

func NewLaunch() *Task {
	return &Task{
		Type: TaskLaunch,
	}
}
