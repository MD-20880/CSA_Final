package stubs

var WorkerCalculate = "Worker.Calculate"
var DistributorPublish = "Broker.HandleTask"
var WorkerSubscribe = "Broker.Subscribe"
var KillBroker = "Broker.Kill"
var KillWorker = "Worker.Kill"
var KillHandler = "Broker.StopWork"
var GetMap = "Broker.Getmap"
var GetCells = "Broker.GetCells"

type Cell struct {
	X, Y int
}

//Stubs Here are appear in pairs, if one request do not need response, use StatusReport as response.

//Request
// Distributor -> Broker ( publish task )
type PublishTask struct {
	ID          string
	GolMap      [][]byte
	Turns       int
	ImageWidth  int
	ImageHeight int
}

//Response
// response for Gol result request
type GolResultReport struct {
	StartX       int
	StartY       int
	EndX         int
	EndY         int
	ResultMap    [][]byte
	CompleteTurn int
}

//Request, responded by StatusReport
type Subscribe struct {
	WorkerAddr string
	Callback   string
}

//Request, responded by GolResultReport
// request for Gol result request
type Work struct {
	Turns        int
	ImageWidth   int
	ImageHeight  int
	StartX       int
	StartY       int
	EndX         int
	EndY         int
	CalculateMap [][]byte
	Owner        string
}

//Respond
//TODO : Still unimplemented
type SdlUpdate struct {
	TurnComplete int
	flipCells    []Cell
}

//Request, responede by RespondCurrentWorld
//Request for last calculated world by ID
type RequestCurrentWorld struct {
	ID string
}

//Response
type RespondCurrentWorld struct {
	World [][]byte
	Turn  int
}

//Request, responded by RespondCurrentCell
//Request for last calculated world's alive cells
type RequestCurrentCells struct {
	ID string
}

type RespondCurrentCell struct {
	Cells []Cell
	Turn  int
}

//Request
//Send when you want to close entire system
type Kill struct {
	Msg string
}

//Request
//Send if you want to quit before get result
//Recommended but not compulsory, save calculation power.
type WorkStop struct {
	Id string
}

//Response
// If you don't need Response, use this as response interface.
type StatusReport struct {
	Msg string
}
