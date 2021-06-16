package anet

import (
	"fmt"
	"net"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	InvalidFileDescription = -1
	ReadMaxLineLen = 64 * 1024
	GeneralIoBuf   = 16 * 1024
	RequestTypeInline = 1
	RequestTypeMultiBuck = 2
)

//pool of io buffer
var (
	bufPool = sync.Pool {
		New: func() interface{} {
			return make([]byte, GeneralIoBuf)
		},
    }
)

//user
type Client struct {
	Fd int
	RBuffer []byte
	WBuffer []byte
	Argc int8
	Argv [][]byte
	rdPos int //当前缓冲的数据
	multiBulkLen int
	bulkLen int
	LastInteraction time.Time //上次交互时间
	requestType int8
}


//读取客户端的请求数据
//protocol designed similar to redis
//*5\r\n$12\r\n
func readQueryFromClient(fd int, event uint8, private unsafe.Pointer) {
    client := (*Client)(private)
    //multi reserved
    buf := bufPool.Get().([]byte)
    n, err := syscall.Read(fd, buf[0:])
    if err != nil {
		if er, ok := err.(syscall.Errno); ok {
			if er.Temporary() || er.Timeout() {
				fmt.Println("tmp")
				return
			}
		}
		UnlinkClient(client)
		return
	}
	if n == 0 {
		fmt.Println("n==0")
		bufPool.Put(buf[0:0])
		return
	}
	//读取的内容append到用户的缓冲区
	client.RBuffer = append(client.RBuffer, buf[0:n]...)
	//归还缓冲区
	bufPool.Put(buf[0:0])
	if client.requestType == -1 {
		if client.RBuffer[client.rdPos] == '*' {
			client.requestType = RequestTypeMultiBuck
		} else {
			client.requestType = RequestTypeInline
		}
	}
	fmt.Println(string(client.RBuffer))
}

//free client
func UnlinkClient(client *Client) {
    if client.Fd != InvalidFileDescription {
    	syscall.Close(client.Fd)
		MainPollApi.DeleteEvent(client.Fd, READ_EVENT)
	}
}

func CreateClient(fd int) {
	client := new(Client)
	client.Fd = fd
	client.RBuffer = make([]byte, 0, 2048)
	client.WBuffer = make([]byte, 0, 2048)

	//设置文件描述符非阻塞 开启探活
	if syscall.SetNonblock(fd, false) != nil || syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1) != nil {
		syscall.Close(fd)
		return
	}

	err := MainPollApi.AddEvent(fd, READ_EVENT, readQueryFromClient, nil, unsafe.Pointer(client))
	if err != nil {
		syscall.Close(fd)
		return
	}
}

//创建一个TCP服务器
func CreateTcpServer(host string, port int, backlog int) (fd int, err error) {
	fd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_IP)
	if err != nil {
		return
	}
	//服务端开启keepalive
	addr := syscall.SockaddrInet4{Port: port}
	copy(addr.Addr[:], net.ParseIP(host).To4())
	if err = syscall.SetNonblock(fd, true); err != nil {
		goto ERR
	}
	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		goto ERR
	}
	err = syscall.Bind(fd, &addr)
	if  err != nil{
		goto ERR
	}
	err = syscall.Listen(fd, backlog)
	if err == nil {
		return
	}
ERR:
	syscall.Close(fd)
	return
}

//general accept
//frequency means how many times wo loop the requests
//func(fd int, event uint8, private unsafe.Pointer)
func GeneralAccept(sfd int, event uint8, private unsafe.Pointer) {
	var frequency int = 10
	for frequency > 0 {
		fd, _, err := syscall.Accept(sfd)
		if err != nil {
			//区分由于信号引起的中断 EINTR 还有由于请求队列没有请求的情况
			if err == syscall.EINTR {
				continue
			} else if err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
				break
			} else {
				panic(err)
			}
		}
		CreateClient(fd)
		frequency--
	}
}
