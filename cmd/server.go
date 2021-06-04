/**
 * copyright redis all right reserved @2021
 * we use golang to rewrite redis service
 */
package main

import (
	poll "assem/anet"
	"unsafe"
)

var eventApi *poll.EventLoopApi

var err error = nil

func init() {
	eventApi, err = poll.GetEventApi(1000)
	if err != nil {
		panic(err)
	}
}

func main() {
    sfd, err := poll.CreateTcpServer("127.0.0.1", 3000, 128)
    if err != nil {
    	panic(err)
	}
	eventApi.AddEvent(sfd, poll.READ_EVENT, poll.GeneralAccept, nil, unsafe.Pointer(eventApi))
    eventApi.Loop()
}