package gol

import (
	"github.com/ChrisGora/semaphore"
	"math/rand"
	"net/rpc"
	"strconv"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
)

func tempSave() {
	//p = params
	//c = channels
	//a = *avail
	semaPhore = semaphore.Init(1, 1)
	rand.Seed(time.Now().UnixNano())
	id := strconv.Itoa(rand.Int())
	conn, _ := rpc.Dial("tcp", "127.0.0.1:8030")
	defer conn.Close()
	//getServerList()

	// TODO: Create a 2D slice to store the world.
	world = make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}

	//Pass File name to IO part
	file := strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(p.ImageWidth)
	c.ioCommand <- ioInput
	c.ioFilename <- file

	//Receive image from IO Part
	for i := range world {
		for j := range world[i] {
			world[i][j] = <-c.ioInput
		}
	}

	turn = 0

	// TODO: Execute all turns of the Game of Life.
	chans := make([]chan [][]byte, p.Threads)
	for i := range chans {
		chans[i] = make(chan [][]byte)
	}
	//Task 3

	go dreportCount()
	//go reportCount()
	////go checkKeyPressed(keyPressed)

	//go reportCount()

	//for i := range world {
	//	for j := range world[i] {
	//		if world[i][j] == 255 {
	//			c.events <- CellFlipped{turn, util.Cell{i, j}}
	//		}
	//	}
	//}

	//c.events <- TurnComplete{CompletedTurns: turn}

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
	mutex.Unlock()

	//Run GOL implementation for TURN times.
	//for i := 1; i <= p.Turns; i++ {
	//	semaPhore.Wait()
	//
	//	//newWorld = updateTurn(chans)
	//	req := stubs.PublishTask{
	//		ID: 		 id,
	//		GolMap:      world,
	//		Turns:       1,
	//		ImageWidth:  p.ImageWidth,
	//		ImageHeight: p.ImageHeight,
	//	}
	//
	//	res := new(stubs.GolResultReport)
	//	conn.Call(stubs.DistributorPublish, req, res)
	//	newWorld = res.ResultMap
	//	//stupid function
	//	//flipCells := checkFlipCells(&world,&newWorld,p)
	//	//smart one
	//	flipCells := CheckFlipCells()
	//	for j := range flipCells {
	//		c.events <- CellFlipped{turn, flipCells[j]}
	//	}
	//	c.events <- TurnComplete{CompletedTurns: turn}
	//	//cell Flipped event
	//	mutex.Lock()
	//	world = newWorld
	//	turn = i
	//	mutex.Unlock()
	//	semaPhore.Post()
	//}

	// TODO: Report the final state using FinalTurnCompleteEvent.
	quit()
	close(c.events)
}

func temp2() {
	//func updateTurn(chans []chan [][]byte) [][]byte {
	//	var updatedWorld [][]byte
	//	for i := 0; i < p.Threads-1; i++ {
	//	go startWorker(i*p.ImageHeight/p.Threads, 0, (i+1)*p.ImageHeight/p.Threads, p.ImageWidth, chans[i], serverList[i%(len(serverList))])
	//}
	//	go startWorker((p.Threads-1)*p.ImageHeight/p.Threads, 0, p.ImageHeight, p.ImageWidth, chans[p.Threads-1], serverList[0])
	//
	//	for i := range chans {
	//	tempStore := <-chans[i]
	//	updatedWorld = append(updatedWorld, tempStore...)
	//}
	//
	//	return updatedWorld
	//
	//}
}

func getServerList() {
	connMap = map[string]*rpc.Client{}
	serverList = make([]string, 0)
	scanner := readfile("gol/serverList")
	for scanner.Scan() {
		serverList = append(serverList, scanner.Text())
	}
	for _, server := range serverList {
		connMap[server], _ = rpc.Dial("tcp", server)
	}
}
