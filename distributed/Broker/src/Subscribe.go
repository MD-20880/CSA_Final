package BrokerService

import (
	"net/rpc"
	"os"
	"uk.ac.bris.cs/gameoflife/stubs"
)

func SubscriberLoop(req stubs.Subscribe, work chan stubs.Work) {
	conn, e := rpc.Dial("tcp", req.WorkerAddr)
	if e != nil {
		os.Exit(2)
	}

	id := IdGenerator()
	Subscribers[id] = conn
	for {
		//build connection
		currentWork := <-work
		workResult, err := working(conn, currentWork, req.Callback)
		//If Error Occur, Put current work back into work queue
		if err != nil {
			conn.Close()
			SubscribersMx.Lock()
			delete(Subscribers, id)
			SubscribersMx.Unlock()
			work <- currentWork
			break
		}
		BufferMx.RLock()
		bufferChan := Buffers[currentWork.Owner]
		BufferMx.RUnlock()

		bufferChan <- workResult
	}
}

func Subscribe(req stubs.Subscribe, res *stubs.StatusReport) (err error) {
	go SubscriberLoop(req, WorkChan)
	res.Msg = "Get It"
	return
}

func working(conn *rpc.Client, work stubs.Work, callback string) (res *stubs.GolResultReport, err error) {
	response := new(stubs.GolResultReport)
	err = conn.Call(callback, work, response)
	if err != nil {
		return
	}
	return response, nil
}
