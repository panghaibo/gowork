/**
 * copyright redis all right reserved @2021
 * we use golang to rewrite redis service
 */
package main

import (
	"github.com/haibo/gowork/anet"
)


func main() {
    sfd, err := anet.CreateTcpServer("127.0.0.1", 3000, 128)
    if err != nil {
    	panic(err)
	}
	anet.MainPollApi.AddEvent(sfd, anet.READ_EVENT, anet.GeneralAccept, nil, nil)
    anet.MainPollApi.Loop()
}