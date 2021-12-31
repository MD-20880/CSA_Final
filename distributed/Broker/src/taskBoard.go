package BrokerService

import (
	"github.com/ChrisGora/semaphore"
	"net/rpc"
	"sync"
	"uk.ac.bris.cs/gameoflife/stubs"
)


//Worker Subscriber List
var Subscribers map[string]*rpc.Client
var SubscribersMx sync.RWMutex

////Topics : {id : work sending channel} contain channels for handler sending work to subscribers
//var Topics map[string]chan stubs.Work
//var TopicsMx sync.RWMutex

//Buffers : { id : result receiving channel } contain channels for handler receiving working result
var Buffers map[string]chan *stubs.GolResultReport
var BufferMx sync.RWMutex

////WorkSemaList : { id : whether work exist in channel } Use to identify whether channel has work left
//var WorkSemaList map[string]semaphore.Semaphore
//var WorkSemaListMx sync.RWMutex

var EventChannels map[string]chan EventRequest
var EventChannelsMx sync.RWMutex

var WorkSema semaphore.Semaphore

var WorkChan chan stubs.Work

var TestChan chan EventRequest

//func WorkDistributor() {
//	for {
//		WorkSema.Wait()
//		WorkSemaListMx.RLock()
//		for key := range WorkSemaList {
//			if _, ok := WorkSemaList[key]; !ok {
//				continue
//			}
//			sema := WorkSemaList[key]
//
//			if sema.GetValue() == 0 {
//				continue
//			}
//			sema.Wait()
//
//			TopicsMx.RLock()
//			topicChan := Topics[key]
//			TopicsMx.RUnlock()
//			work := <-topicChan
//			WorkChan <- work
//
//			break
//		}
//		WorkSemaListMx.RUnlock()
//	}
//}



