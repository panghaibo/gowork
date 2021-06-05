package anet

import (
	"net"
	"syscall"
	"unsafe"
)

//user
type Client struct {
	Fd int
	RBuffer []byte
	WBuffer []byte
}

//读取客户端的请求数据
func readQueryFromClient(fd int, event uint8, private unsafe.Pointer) {
    client := (*Client)(private)
    _, err := syscall.Read(fd, client.RBuffer[:])
    if err != nil {
    	if err == syscall.EINTR || err == syscall.EWOULDBLOCK {
    		return
		}
		syscall.Close(fd)
	}
    syscall.Write(fd, []byte("alice"))
}

func CreateClient(fd int) {
	client := new(Client)
	client.Fd = fd
	client.RBuffer = make([]byte, 8096)
	client.WBuffer = make([]byte, 8096)

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
	if err = syscall.SetNonblock(fd, true); err != nil {
		syscall.Close(fd)
		return
	}
	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		syscall.Close(fd)
		return
	}
	//服务端开启keepalive
	addr := syscall.SockaddrInet4{Port: 3000}
	copy(addr.Addr[:], net.ParseIP("127.0.0.1").To4())

	if syscall.Bind(fd, &addr) != nil || syscall.Listen(fd, backlog) != nil {
		syscall.Close(fd)
		return
	}
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
			if err == syscall.EINTR || err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
				break
			} else {
				panic(err)
			}
		}
		CreateClient(fd)
		frequency--
	}
}
