package main

import (
	"flag"
	"fmt"
	"github.com/ChrisGora/semaphore"
	"net"
	"net/rpc"
	"os"
	"sync"
	BrokerService "uk.ac.bris.cs/gameoflife/Broker/src"
	"uk.ac.bris.cs/gameoflife/stubs"
)

type Broker struct {
}

func (b *Broker) HandleTask(req stubs.PublishTask, res *stubs.GolResultReport) (err error) {
	id := req.ID
	fmt.Println(id)

	if _, ok := BrokerService.Buffers[id]; ok {
		return &BrokerService.ChannelExist{}
	}

	////Initialize Topics
	//BrokerService.TopicsMx.Lock()
	//BrokerService.Topics[id] = make(chan stubs.Work, 1)
	//BrokerService.TopicsMx.Unlock()

	//Initialize Buffers
	BrokerService.BufferMx.Lock()
	BrokerService.Buffers[id] = make(chan *stubs.GolResultReport, 1)
	BrokerService.BufferMx.Unlock()
	//
	////Initialize WorkSemaList
	//BrokerService.WorkSemaListMx.Lock()
	//BrokerService.WorkSemaList[id] = semaphore.Init(1, 0)
	//BrokerService.WorkSemaListMx.Unlock()

	//Initialize EventChannel
	BrokerService.EventChannelsMx.Lock()
	BrokerService.EventChannels[id] = make(chan BrokerService.EventRequest)
	BrokerService.EventChannelsMx.Unlock()

	//Start Handler
	BrokerService.HandleTask(req, res, id)
	return

}

func (b *Broker) Kill(req stubs.Kill, res *stubs.StatusReport) (err error) {
	go quitBroker()
	return

}

func (b *Broker) StopWork(req stubs.WorkStop, res *stubs.StatusReport) (err error) {
	BrokerService.EventChannelsMx.RLock()
	BrokerService.EventChannels[req.Id] <- BrokerService.HandlerStopEvent{Cmd: BrokerService.HandlerStop}
	BrokerService.EventChannelsMx.RUnlock()
	return
}

func (b *Broker) Subscribe(req stubs.Subscribe, res *stubs.StatusReport) (err error) {
	fmt.Println("Receve Subscribe")
	BrokerService.Subscribe(req, res)
	res.Msg = "Got it"
	return
}

func (b *Broker) GetCells(req stubs.RequestCurrentCells, res *stubs.RespondCurrentCell) (err error) {
	fmt.Println("Receive Request")
	fmt.Println(req.ID)

	if _, ok := BrokerService.Buffers[req.ID]; !ok {
		return
	}

	resultChan := make(chan BrokerService.CurrentWorld, 1)
	BrokerService.EventChannelsMx.RLock()
	tempChan := BrokerService.EventChannels[req.ID]
	BrokerService.EventChannelsMx.RUnlock()
	tempChan <- BrokerService.GetMapEvent{BrokerService.GetMap, resultChan}

	result := <-resultChan

	res.Cells = BrokerService.CalculateAliveCells(result.World)
	res.Turn = result.Turn
	return

}

func (b *Broker) Getmap(req stubs.RequestCurrentWorld, res *stubs.RespondCurrentWorld) (err error) {
	fmt.Println("Receive Request")
	fmt.Println(req.ID)

	if _, ok := BrokerService.Buffers[req.ID]; !ok {
		return
	}

	resultChan := make(chan BrokerService.CurrentWorld, 1)
	BrokerService.EventChannelsMx.RLock()
	tempChan := BrokerService.EventChannels[req.ID]
	BrokerService.EventChannelsMx.RUnlock()
	tempChan <- BrokerService.GetMapEvent{BrokerService.GetMap, resultChan}

	result := <-resultChan

	res.World = result.World
	res.Turn = result.Turn
	return

}

//Broker initialization
func initializeBroker() {
	//BrokerService.Topics = map[string]chan stubs.Work{}
	//BrokerService.TopicsMx = sync.RWMutex{}

	BrokerService.Buffers = map[string]chan *stubs.GolResultReport{}
	BrokerService.BufferMx = sync.RWMutex{}

	//BrokerService.WorkSemaList = map[string]semaphore.Semaphore{}
	//BrokerService.WorkSemaListMx = sync.RWMutex{}

	BrokerService.EventChannels = map[string]chan BrokerService.EventRequest{}
	//BrokerService.WorkSemaListMx = sync.RWMutex{}

	BrokerService.Subscribers = map[string]*rpc.Client{}

	BrokerService.WorkChan = make(chan stubs.Work, 1)

	BrokerService.WorkSema = semaphore.Init(999, 0)

	BrokerService.Counter = 0

	BrokerService.TestChan = make(chan BrokerService.EventRequest)

}

func quitBroker() {
	BrokerService.SubscribersMx.Lock()
	BrokerService.EventChannelsMx.Lock()

	for keys := range BrokerService.EventChannels {
		BrokerService.EventChannels[keys] <- BrokerService.HandlerStopEvent{Cmd: BrokerService.HandlerStop}
	}
	for keys := range BrokerService.Subscribers {
		newReq := stubs.Kill{Msg: "kill"}
		newRes := new(stubs.StatusReport)
		BrokerService.Subscribers[keys].Go(stubs.KillWorker, newReq, newRes, nil)
	}
	os.Exit(10)
}

func main() {

	initializeBroker()
	//go BrokerService.WorkDistributor()
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rpc.Register(&Broker{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()
	rpc.Accept(listener)

}
