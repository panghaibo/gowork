// +build linux

package anet

import (
	"syscall"
)

type NetPollApi struct {
	eFd int
	Size int
	Events []syscall.EpollEvent
}

func GetNetPollApi(eventLoopApi *EventLoopApi) (*NetPollApi, error) {
	eFd, err := syscall.EpollCreate(eventLoopApi.Size)
	if err != nil {
		return nil, err
	}
    api := new(NetPollApi)
    api.eFd = eFd
    api.Events = make([]syscall.EpollEvent, eventLoopApi.Size)
	return api, nil
}


func (p *NetPollApi) AddEvent(eventLoopApi *EventLoopApi, fd int, mask uint8) error {
	event := eventLoopApi.Events[fd]
	var op int
	if event.Events == NONE_EVENT {
		op = syscall.EPOLL_CTL_ADD
	} else {
		op = syscall.EPOLL_CTL_MOD
	}
	ev := syscall.EpollEvent{
		Fd : int32(fd),
	}
	mask |= event.Events
	if mask & READ_EVENT > 0 {
		ev.Events |= syscall.EPOLLIN
	}
	if mask & WRITE_EVENT > 0 {
		ev.Events |= syscall.EPOLLOUT
	}
    return syscall.EpollCtl(p.eFd, op, fd, &ev)
}

func (p *NetPollApi) DeleteEvent(eventLoopApi *EventLoopApi, fd int, delMask uint8) error {
	event := eventLoopApi.Events[fd]
	mask := event.Events & (^delMask)
	ev := syscall.EpollEvent{
		Fd : int32(fd),
	}
	if mask & READ_EVENT > 0 {
		ev.Events |= syscall.EPOLLIN
	}
	if mask & WRITE_EVENT > 0 {
		ev.Events |= syscall.EPOLLOUT
	}
	if mask == NONE_EVENT {
		return syscall.EpollCtl(p.eFd, syscall.EPOLL_CTL_DEL, fd, &ev)
	} else {
		return syscall.EpollCtl(p.eFd, syscall.EPOLL_CTL_MOD, fd, &ev)
	}
}

func (p *NetPollApi) Loop(eventLoopApi *EventLoopApi, msec *int) (int, error) {
	mseconds := -1
	if msec != nil {
		mseconds = *msec
	}
	REDO:
	n, err := syscall.EpollWait(p.eFd, p.Events, mseconds)
    if err != nil {
    	if err == syscall.EINTR {
    		goto REDO
		}
    	return n, err
	}
	for i:=0; i < n; i++ {
		ev := p.Events[i]
		var mask uint8 = 0
		if ev.Events & syscall.EPOLLIN > 0 {
			mask |= READ_EVENT
		}
		if ev.Events & syscall.EPOLLOUT > 0 {
			mask |= WRITE_EVENT
		}
		if ev.Events & syscall.EPOLLERR > 0 {
			mask |= READ_EVENT | WRITE_EVENT
		}
		if ev.Events & syscall.EPOLLHUP > 0 {
			mask |= READ_EVENT | WRITE_EVENT
		}
		eventLoopApi.FiredEvent[i].Events = mask
		eventLoopApi.FiredEvent[i].Fd = int(ev.Fd)
	}
	return n, nil
}

func (p *NetPollApi) Close() {
	syscall.Close(p.eFd)
}


func (p *NetPollApi) GetApiName() string {
	return "ePoll"
}
