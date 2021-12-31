package gol

import (
	"fmt"
	"os"
	"strconv"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

func reportCount() {
	for {
		time.Sleep(2 * time.Second)
		mutex.Lock()
		result := CalculateAliveCells(world)
		currentTurn := turn
		mutex.Unlock()
		if a.events == true {
			c.events <- AliveCellsCount{
				CompletedTurns: currentTurn,
				CellsCount:     len(result),
			}

		} else {
			return
		}
	}

}

func storePgm() {
	c.ioCommand <- ioOutput
	filename := strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(turn)
	c.ioFilename <- filename
	for i := range world {
		for j := range world[i] {
			c.ioOutput <- world[i][j]
		}
	}
}

func checkKeyPressed(keyPressed <-chan rune) {
	for {
		i := <-keyPressed
		semaPhore.Wait()
		switch i {
		case 'k':
			{
				req := stubs.Kill{Msg: "kill"}
				res := new(stubs.StatusReport)
				conn.Go(stubs.KillBroker, req, res, nil)
				quit()
			}
		case 's':
			storePgm()
		case 'p':
			{
				key := <-keyPressed
				for key != 'p' {
					key = <-keyPressed
				}
				fmt.Printf("Continuing\n")
			}
		case 'q':
			quit()
			os.Exit(1)
		}
		semaPhore.Post()

	}
}

func SDLWorkFlow(keyPressed <-chan rune, id string) {

	go reportCount()
	go checkKeyPressed(keyPressed)

	for i := range world {
		for j := range world[i] {
			if world[i][j] == 255 {
				c.events <- CellFlipped{turn, util.Cell{i, j}}
			}
		}
	}

	//c.events <- TurnComplete{CompletedTurns: turn}

	//Run GOL implementation for TURN times.
	for i := 1; i <= p.Turns; i++ {
		semaPhore.Wait()

		//newWorld = updateTurn(chans)
		req := stubs.PublishTask{
			ID:          id,
			GolMap:      world,
			Turns:       1,
			ImageWidth:  p.ImageWidth,
			ImageHeight: p.ImageHeight,
		}

		res := new(stubs.GolResultReport)
		conn.Call(stubs.DistributorPublish, req, res)
		newWorld = res.ResultMap

		flipCells := CheckFlipCells()
		for j := range flipCells {
			c.events <- CellFlipped{turn, flipCells[j]}
		}
		c.events <- TurnComplete{CompletedTurns: turn}
		//cell Flipped event
		mutex.Lock()
		world = newWorld
		turn = i
		mutex.Unlock()
		semaPhore.Post()
	}

	quit()
	close(c.events)
}
