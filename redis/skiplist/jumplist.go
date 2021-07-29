package skiplist

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

type pointerColumn struct {
	next []*Column
}

type Column struct {
	pointerColumn
	key   float64
	value interface{}
}

type SkipList struct {
	startPointers pointerColumn
	maxLevel      int
	randomSeed    rand.Source
	probabilities []float64
	mutex         sync.RWMutex
	levelCursors  []*pointerColumn
}

func (list *SkipList) moveCursors(key float64) {
	pointerColumn := &list.startPointers

	for i := list.maxLevel - 1; i >= 0; i-- { //move from the top
		nextColumn := pointerColumn.next[i]

		for nextColumn != nil && key > nextColumn.key {
			pointerColumn = &nextColumn.pointerColumn //result if it is the end
			nextColumn = nextColumn.next[i]           //keep move to the right
		}
		list.levelCursors[i] = pointerColumn //this is to save fingers
	}
}

func (list *SkipList) Set(key float64, value interface{}) *Column {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	list.moveCursors(key)
	column := list.levelCursors[0].next[0]  //bottom layer
	if column != nil && column.key <= key { //check if successfully get
		column.value = value
		return column
	}

	//not exists, so create a column
	column = &Column{pointerColumn{make([]*Column, list.randLevel())}, key, value}

	//set column next and previous column next
	for i := range column.next { //remember that resultPointers[i].next[i] is the previous column
		column.next[i] = list.levelCursors[i].next[i] // resultPointers[i].next[i] is the future next
		list.levelCursors[i].next[i] = column         //update resultPointers[i].next[i] to new column
	}

	return column
}

func (list *SkipList) Get(key float64) *Column {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	pointers := &list.startPointers
	var next *Column

	for i := list.maxLevel - 1; i >= 0; i-- { // from the top level to the bottom level
		next = pointers.next[i]

		for next != nil && key > next.key { // if key > storage of this column, move to the right
			pointers = &next.pointerColumn
			next = next.next[i]
		} //move down again
	}

	if next != nil && next.key == key {
		return next
	}

	return nil
}

func (list *SkipList) Del(key float64) *Column {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	list.moveCursors(key)
	column := list.levelCursors[0].next[0]

	if column != nil && column.key <= key { //value found
		for k, v := range column.next { //found next column (which is results[0].next[0].next[k])
			list.levelCursors[k].next[k] = v //modify current column to next-next
		}

		return column
	}

	return nil
}

func (list *SkipList) randLevel() int {
	r := float64(list.randomSeed.Int63()) / (1 << 63) // https://golang.org/src/math/rand/rand.go#L178

	for level, prob := range list.probabilities {
		if r > prob {
			return level + 1
		}
	}
	return list.maxLevel
}

func NewWithLevel(level int) *SkipList {
	if level < 1 || level > 64 {
		panic("level must be 1~64")
	}
	probabilities := []float64{}
	prob := 1.0
	for i := 1; i <= level; i++ {
		prob /= math.E
		probabilities = append(probabilities, prob)
	}
	return &SkipList{
		startPointers: pointerColumn{next: make([]*Column, level)},
		levelCursors:  make([]*pointerColumn, level),
		maxLevel:      level,
		randomSeed:    rand.New(rand.NewSource(time.Now().UnixNano())),
		probabilities: probabilities,
	}
}

func New() *SkipList {
	return NewWithLevel(18) //e^18 = 65659969
}
