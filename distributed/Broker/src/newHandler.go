package BrokerService

import (
	"fmt"
	"os"
	"sync"
	"uk.ac.bris.cs/gameoflife/stubs"
)

type variables struct {
	id               string
	req              stubs.PublishTask
	res              *stubs.GolResultReport
	completeTurn     int
	CompleteWorld    [][]byte
	CalculatingWorld [][]byte
	WorkList         []stubs.Work
	WorkNum          int
	ResultChan       chan *stubs.GolResultReport
	worldMx          sync.Mutex
	eventChan        chan EventRequest
	closed           bool
}

func initVars(req stubs.PublishTask, res *stubs.GolResultReport, id string) (v variables) {

	WorkList := make([]stubs.Work, 0)

	BufferMx.RLock()
	ResultChan := Buffers[id]
	BufferMx.RUnlock()

	CompleteWorld := req.GolMap

	CalculatingWorld := make([][]byte, len(CompleteWorld))
	for i := 0; i < len(CompleteWorld); i++ {
		CalculatingWorld[i] = make([]byte, len(CompleteWorld[i]))
	}

	EventChannelsMx.RLock()
	eventChan := EventChannels[id]
	EventChannelsMx.RUnlock()

	return variables{
		id:               id,
		req:              req,
		res:              res,
		completeTurn:     0,
		CompleteWorld:    CompleteWorld,
		CalculatingWorld: CalculatingWorld,
		WorkList:         WorkList,
		WorkNum:          0,
		ResultChan:       ResultChan,
		worldMx:          sync.Mutex{},
		eventChan:        eventChan,
		closed:           false,
	}
}

func workSplit(v variables) []stubs.Work {
	splitResult := make([]stubs.Work, 0)
	noSubscribers := len(Subscribers)
	if noSubscribers == 0 {
		os.Exit(3)
	}
	for i := 0; i < noSubscribers-1; i++ {
		splitResult = append(splitResult, stubs.Work{
			Turns:        1,
			ImageWidth:   v.req.ImageWidth,
			ImageHeight:  v.req.ImageHeight,
			StartX:       i * v.req.ImageHeight / noSubscribers,
			StartY:       0,
			EndX:         (i + 1) * v.req.ImageWidth / noSubscribers,
			EndY:         v.req.ImageWidth,
			CalculateMap: v.CompleteWorld,
			Owner:        v.id,
		})
	}
	splitResult = append(splitResult, stubs.Work{
		Turns:        1,
		ImageWidth:   v.req.ImageWidth,
		ImageHeight:  v.req.ImageHeight,
		StartX:       (noSubscribers - 1) * v.req.ImageHeight / noSubscribers,
		StartY:       0,
		EndX:         v.req.ImageHeight,
		EndY:         v.req.ImageWidth,
		CalculateMap: v.CompleteWorld,
		Owner:        v.id,
	})
	return splitResult
}

func postWork(workList []stubs.Work, id string) {
	for _, work := range workList {

		workList = append(workList, work)
	}
}

func workSender(workList []stubs.Work, id string) {
	//WorkMutex.RLock()
	for _, work := range workList {
		//WorkSemaListMx.RLock()
		//currentWorkSema := WorkSemaList[id]
		//WorkSemaListMx.RUnlock()
		//currentWorkSema.Post()
		//
		//WorkSema.Post()
		//
		//TopicsMx.RLock()
		//topicChan := Topics[id]
		//TopicsMx.RUnlock()
		//topicChan <- work
		WorkChan <- work
	}
}

func checkWork(v variables) {
	for v.WorkNum > 0 {
		work := <-v.ResultChan
		if work.EndX > len(v.CalculatingWorld) {
			break
		}
		receive(work, v)
		for i := work.StartX; i < work.EndX; i++ {
			v.CalculatingWorld[i] = work.ResultMap[i-work.StartX]
		}
		v.WorkNum--
	}
}

func receive(jobResult *stubs.GolResultReport, v variables) {
	if len(v.WorkList) == 0 {
		return
	}
	for i, work := range v.WorkList {
		if work.StartX == jobResult.StartX {
			if len(v.WorkList) > 1 {
				v.WorkList = append(v.WorkList[:i], v.WorkList[i+1:]...)
			} else {
				v.WorkList = make([]stubs.Work, 0)
			}
			break
		}
	}
}

func reply(v variables) {
	v.res.ResultMap = v.CompleteWorld
	v.res.CompleteTurn = v.req.Turns
	v.res.StartX = 0
	v.res.StartY = 0
	v.res.EndY = len(v.req.GolMap)
	v.res.EndX = len(v.req.GolMap[0])
}

func closeHandler(id string) {
	//TopicsMx.Lock()
	//close(Topics[id])
	//delete(Topics, id)
	//TopicsMx.Unlock()

	BufferMx.Lock()
	close(Buffers[id])
	delete(Buffers, id)
	BufferMx.Unlock()

	//WorkSemaListMx.Lock()
	//delete(WorkSemaList, id)
	//WorkSemaListMx.Unlock()

	EventChannelsMx.Lock()
	delete(EventChannels, id)
	EventChannelsMx.Unlock()

}

//Event Handler
func eventHandler(v *variables) {
	//add a receiver here
LOOP:
	for {
		fmt.Println(v.id)
		event := <-v.eventChan
		switch event.Command() {
		case GetMap:
			resultChan := event.(GetMapEvent).SendBack
			sendNum := v.completeTurn
			send := CurrentWorld{
				World: v.CompleteWorld,
				Turn:  sendNum,
			}
			resultChan <- send

		case HandlerStop:
			v.closed = true

		case HandlerLoopStop:
			closeHandler(v.id)
			break LOOP
		}

	}
}

func HandleTask(req stubs.PublishTask, res *stubs.GolResultReport, id string) (err error) {

	//Initialize variables
	v := initVars(req, res, id)
	go eventHandler(&v)
	//Task Cycle
	for v.completeTurn = 0; !v.closed && v.completeTurn < req.Turns; {
		//Split One big task into several small tasks
		v.WorkList = workSplit(v)
		//Record the number of work been send
		v.WorkNum = len(v.WorkList)
		//Post Work
		postWork(v.WorkList, v.id)

		workSender(v.WorkList, v.id)
		checkWork(v)
		//res.ResultMap = CalculatingWorld
		v.worldMx.Lock()
		v.CompleteWorld = v.CalculatingWorld
		v.completeTurn++
		v.worldMx.Unlock()
		v.CalculatingWorld = make([][]byte, len(v.CompleteWorld))
		//for i := 0; i < len(v.CompleteWorld); i++ {
		//	v.CalculatingWorld[i] = make([]byte, len(v.CompleteWorld[i]))
		//}
	}
	//Response to Request
	reply(v)

	v.eventChan <- HandlerLoopStopEvent{Cmd: HandlerLoopStop}
	return
}
