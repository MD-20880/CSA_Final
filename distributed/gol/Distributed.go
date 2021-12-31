package gol

import (
	"fmt"
	"os"
	"strconv"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
)

func requestMap() (world [][]byte, turn int) {
	req := stubs.RequestCurrentWorld{ID: id}
	res := new(stubs.RespondCurrentWorld)
	conn.Call("Broker.Getmap", req, res)
	world = res.World
	turn = res.Turn
	return
}

func dstorePgm() {
	world, turn = requestMap()
	c.ioCommand <- ioOutput
	filename := strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(turn)
	c.ioFilename <- filename
	for i := range world {
		for j := range world[i] {
			c.ioOutput <- world[i][j]
		}
	}
}

func dreportCount() {
	for {
		time.Sleep(2 * time.Second)
		currentWorld, currentTurn := requestMap()
		mutex.Lock()
		result := CalculateAliveCells(currentWorld)
		mutex.Unlock()
		if a.events == true {
			c.events <- AliveCellsCount{
				CompletedTurns: currentTurn,
				CellsCount:     len(result),
			}
			c.events <- TurnComplete{CompletedTurns: currentTurn}

		} else {
			return
		}
	}
}

func dcheckKeyPressed(keyPressed <-chan rune) {
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
			dstorePgm()
		case 'p':
			{
				key := <-keyPressed
				for key != 'p' {
					key = <-keyPressed
				}
				fmt.Printf("Continuing\n")
			}
		case 'q':
			conn.Go(stubs.KillHandler, stubs.WorkStop{Id: id}, new(stubs.StatusReport), nil)
			quit()
			os.Exit(1)
		}
		semaPhore.Post()

	}
}

func DistributedWorkFlow(keyPressed <-chan rune, id string) {

	go dreportCount()
	go dcheckKeyPressed(keyPressed)

	req := stubs.PublishTask{
		ID:          id,
		GolMap:      world,
		Turns:       p.Turns,
		ImageWidth:  p.ImageWidth,
		ImageHeight: p.ImageHeight,
	}

	res := new(stubs.GolResultReport)
	conn.Call(stubs.DistributorPublish, req, res)
	newWorld = res.ResultMap
	mutex.Lock()
	world = newWorld
	turn = res.CompleteTurn
	mutex.Unlock()

	// TODO: Report the final state using FinalTurnCompleteEvent.
	quit()
	close(c.events)
}
