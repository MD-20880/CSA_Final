package gol

import "fmt"

// Params provides the details of how to run the Game of Life and which image to load.
type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

// Cell is used as the return type for the testing framework

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func Run(p Params, events chan<- Event, keyPresses <-chan rune) {

	//	TODO: Put the missing channels in here.

	ioCommand := make(chan ioCommand)
	ioIdle := make(chan bool)
	filename := make(chan string)
	output := make(chan byte)
	input := make(chan byte)

	ioChannels := ioChannels{
		command:  ioCommand,
		idle:     ioIdle,
		filename: filename,
		output:   output,
		input:    input,
	}
	go startIo(p, ioChannels)

	distributorChannels := distributorChannels{
		events:     events,
		ioCommand:  ioCommand,
		ioIdle:     ioIdle,
		ioFilename: filename,
		ioOutput:   output,
		ioInput:    input,
	}

	channelStatus := channelAvailibility{
		events:     true,
		ioCommand:  true,
		ioIdle:     true,
		ioFilename: true,
		ioOutput:   true,
		ioInput:    true,
	}
	distributor(p, distributorChannels, &channelStatus, keyPresses)
	fmt.Printf("running Here")
}
