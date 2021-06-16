// +build darwin

package anet

import (
	"fmt"
	anetx "github.com/panghaibo/gowork/anet"
	"syscall"
	"testing"
)

//测试select是否正常运行
//我们使用pipe测试文件描述符的读写是否在横传
func TestFdSet(t *testing.T) {
	fdSetVal := syscall.FdSet{}
	anetx.FdSet(1, &fdSetVal)
    if anetx.FdIsSet(1, &fdSetVal) {
   	   t.Log("fd set success")
    } else {
   	   t.Fatal("fd set error")
    }
	anetx.FdClear(1,  &fdSetVal)
    if anetx.FdIsSet(1, &fdSetVal) {
	   t.Fatal("after fd clear failed")
    } else {
	   t.Log("after fd clear success")
    }
}

//测试
func TestSelectPoll(t *testing.T) {
	t.Log("test select start!!")
	api, err := anetx.GetEventApi(20)
	if err != nil {
		t.Fatal(err)
	}
	err = api.AddEvent(syscall.Stdin, anetx.READ_EVENT,nil,nil,nil)
	if err != nil {
		t.Fatal(err)
	}
	api.DeleteEvent(syscall.Stdin, anetx.READ_EVENT)
}

var name map[string]int

func TestNi(t *testing.T) {
   buf := make([]byte, 0, 10)
   buf = append(buf, "alice"...)
   fmt.Println(len(buf), cap(buf))
   buf = buf[0:0]
	fmt.Println(len(buf), cap(buf))
}
