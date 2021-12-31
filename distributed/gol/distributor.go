package gol

import (
	"bufio"
	"fmt"
	"github.com/ChrisGora/semaphore"
	"math/rand"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

type channelAvailibility struct {
	events     bool
	ioCommand  bool
	ioIdle     bool
	ioFilename bool
	ioOutput   bool
	ioInput    bool
}

var p Params
var world [][]byte
var newWorld [][]byte
var mutex = sync.Mutex{}
var a channelAvailibility
var c distributorChannels
var turn int
var semaPhore semaphore.Semaphore
var serverList []string
var connMap map[string]*rpc.Client
var conn *rpc.Client
var id string

//Parallel Functions

func readfile(path string) bufio.Scanner {
	file, err := os.Open(path)
	//if err != nil{
	//	os.Exit(3)
	//}
	fmt.Println(err)
	scanner := bufio.NewScanner(file)
	return *scanner

}

func ReceiveMap(cells []stubs.Cell) [][]byte {
	resultWorld := make([][]byte, p.ImageHeight)
	for i := range resultWorld {
		resultWorld[i] = make([]byte, p.ImageWidth)
	}
	for _, i := range cells {
		resultWorld[i.X][i.Y] = 255
	}
	return resultWorld
}

//This function Work just well

func CalculateAliveCells(world [][]byte) []util.Cell {
	var cells = []util.Cell{}
	for j, _ := range world {
		for i, num := range world[j] {
			if num == 255 {
				cells = append(cells, util.Cell{i, j})
			}
		}
	}
	return cells
}

func CheckFlipCells() []util.Cell {

	flipCells := make([]util.Cell, 0)
	for i := range world {
		for j := range world[i] {
			if world[i][j] != newWorld[i][j] {
				flipCells = append(flipCells, util.Cell{X: i, Y: j})
			}
		}
	}
	return flipCells
}

func quit() {
	aliveCells := CalculateAliveCells(world)
	c.events <- FinalTurnComplete{
		CompletedTurns: turn,
		Alive:          aliveCells,
	}
	storePgm()
	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	for _, j := range connMap {
		j.Close()
	}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	a.events = false
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(params Params, channels distributorChannels, avail *channelAvailibility, keyPressed <-chan rune) {
	p = params
	c = channels
	a = *avail
	semaPhore = semaphore.Init(1, 1)
	rand.Seed(time.Now().UnixNano())
	id = strconv.Itoa(rand.Int())
	conn, _ = rpc.Dial("tcp", "127.0.0.1:8030")
	//conn, _ = rpc.Dial("tcp", "3.82.148.15:8030")
	defer conn.Close()

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
	SDLWorkFlow(keyPressed, id)
	//DistributedWorkFlow(keyPressed, id)
}
