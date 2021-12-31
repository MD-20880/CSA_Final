package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"uk.ac.bris.cs/gameoflife/stubs"
)

type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

var Threads int

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func calculateNextState(p Params, world [][]byte, startX int, startY int, endX int, endY int, resultChan chan [][]byte) {
	x_scan_map := [3]int{-1, 0, 1}
	y_scan_map := [3]int{-1, 0, 1}

	newWorld := make([][]byte, endX-startX)
	for i := range newWorld {
		newWorld[i] = make([]byte, endY-startY)
	}

	for i := startX; i < endX; i++ {
		for j := startY; j < endY; j++ {
			c := make(chan byte, 10)
			calculateHelper(i, j, &world, x_scan_map, y_scan_map, p, c)
			result := <-c
			//fmt.Printf("startX : %d\n",startX)
			//fmt.Printf("startY : %d\n",startY)
			newWorld[i-startX][j-startY] = result
		}
	}
	resultChan <- newWorld
}

func calculateHelper(x int, y int, oldWorld *[][]byte, xmap [3]int, ymap [3]int, p Params, c chan byte) {
	d_oldWorld := *oldWorld
	alive := 0
	check := func(x_cor int, y_cor int) int {
		if d_oldWorld[x_cor][y_cor] == 255 {
			return 1
		}
		return 0
	}

	for _, x_scan := range xmap {
		xcal := x
		if x+x_scan > p.ImageWidth-1 {
			xcal = 0
		} else if x+x_scan < 0 {
			xcal = p.ImageWidth - 1
		} else {
			xcal = xcal + x_scan
		}
		for _, y_scan := range ymap {
			if x_scan == 0 && y_scan == 0 {
				continue
			}
			if y+y_scan > p.ImageHeight-1 {
				alive += check(xcal, 0)
			} else if y+y_scan < 0 {
				alive += check(xcal, p.ImageHeight-1)
			} else {
				alive += check(xcal, y+y_scan)
			}
		}
	}

	if d_oldWorld[x][y] == 255 && (alive < 2 || alive > 3) {
		c <- 0

	} else if d_oldWorld[x][y] == 0 && alive == 3 {
		c <- 255
	} else {
		c <- d_oldWorld[x][y]
	}
}

func StartWorker(p Params, world [][]byte, startX int, startY int, endX int, endY int, resultChan chan [][]byte) {

	if Threads == 1 {
		calculateNextState(p, world, startX, startY, endX, endY, resultChan)
	} else {
		chans := make([]chan [][]byte, Threads)
		for i := 0; i < Threads; i++ {
			chans[i] = make(chan [][]byte)
			go calculateNextState(p, world, startX, i*(endY-startY)/Threads, endX, (i+1)*(endY-startY)/Threads, chans[i])
		}
		newWorld := make([][][]byte, Threads)
		for i := 0; i < Threads; i++ {
			newWorld[i] = <-chans[i]
		}

		resultWorld := make([][]byte, endX-startX)
		for i := 0; i < endX-startX; i++ {
			for j := 0; j < Threads; j++ {
				resultWorld[i] = append(resultWorld[i], newWorld[j][i]...)
			}
		}

		resultChan <- resultWorld
	}
}

type Worker struct {
}

func (w *Worker) Calculate(request stubs.Work, response *stubs.GolResultReport) (err error) {
	fmt.Printf("Request received\n")
	p := Params{
		Turns:       request.Turns,
		Threads:     1,
		ImageWidth:  request.ImageWidth,
		ImageHeight: request.ImageHeight,
	}
	r := request
	resultMap := make(chan [][]byte, 1)
	StartWorker(p, r.CalculateMap, r.StartX, r.StartY, r.EndX, r.EndY, resultMap)
	resultWorld := <-resultMap
	response.StartX = request.StartX
	response.EndX = request.EndX
	response.StartY = request.StartY
	response.ResultMap = resultWorld
	response.CompleteTurn = request.Turns
	response.EndY = request.EndY
	fmt.Println("Request Finish")
	return
}

func (w *Worker) Kill(req stubs.Kill, rsp *stubs.StatusReport) (err error) {
	os.Exit(10)
	return
}

func subscribeBroker(bAddr string, pAddr string) {
	conn, _ := rpc.Dial("tcp", bAddr)
	ip := GetOutboundIP()
	fmt.Println(ip.String())
	addr := ip.String() + ":" + pAddr
	req := stubs.Subscribe{
		WorkerAddr: addr,
		Callback:   stubs.WorkerCalculate,
	}
	res := new(stubs.StatusReport)
	conn.Call(stubs.WorkerSubscribe, req, res)
	fmt.Printf(res.Msg)
	conn.Close()
}

func main() {
	pAddr := flag.String("port", "8030", "Port to listen on")
	bAddr := flag.String("broker", "127.0.0.1:8030", "Port to listen on")
	threads := flag.String("core", "4", "Threads you want to run on this server")
	//bAddr := flag.String("broker", "3.82.148.15:8030", "Port to listen on")
	flag.Parse()
	thread, err := strconv.Atoi(*threads)
	if err != nil {
		fmt.Println("Threads must be a number")
		os.Exit(20)
	}
	Threads = thread
	rpc.Register(&Worker{})
	listener, err := net.Listen("tcp", ":"+*pAddr)
	for err != nil {
		result, _ := strconv.Atoi(*pAddr)
		*pAddr = strconv.Itoa(result + 10)
		listener, err = net.Listen("tcp", ":"+*pAddr)
	}
	subscribeBroker(*bAddr, *pAddr)
	defer listener.Close()
	fmt.Println("Listining ")
	rpc.Accept(listener)

}
