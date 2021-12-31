package BrokerService

import (
	"fmt"
	"strconv"
	"sync"
	"uk.ac.bris.cs/gameoflife/stubs"
)

var Counter int
var assginMutex sync.Mutex

func errorHandler(err error) {
	fmt.Println(err)
}

func IdGenerator() (id string) {
	assginMutex.Lock()
	id = strconv.Itoa(Counter)
	Counter++
	assginMutex.Unlock()

	return
}

func CalculateAliveCells(world [][]byte) []stubs.Cell {
	var cells = []stubs.Cell{}
	for j, _ := range world {
		for i, num := range world[j] {
			if num == 255 {
				cells = append(cells, stubs.Cell{i, j})
			}
		}
	}
	return cells
}
