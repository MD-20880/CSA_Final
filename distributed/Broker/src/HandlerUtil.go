package BrokerService

import (
	"uk.ac.bris.cs/gameoflife/stubs"
)

func newCheckFlipCells(world [][]byte, newWorld [][]byte) []stubs.Cell {

	flipCells := make([]stubs.Cell, 0)
	for i := range world {
		for j := range world[i] {
			if world[i][j] != newWorld[i][j] {
				flipCells = append(flipCells, stubs.Cell{X: i, Y: j})
			}
		}
	}
	return flipCells
}
