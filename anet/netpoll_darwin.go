// +build darwin

package anet

import (
	"syscall"
	"unsafe"
)

//func Select(nfd int, r *FdSet, w *FdSet, e *FdSet, timeout *Timeval) (n int, err error)
//we define some tool function

const BYTES = unsafe.Sizeof(syscall.FdSet{})

//go:inline
func FdIsSet(fd int, fdSet *syscall.FdSet) bool {
	pos := fd / 32
	idx := fd % 32
    if fdSet.Bits[pos] & (1 << idx) > 0 {
    	return true
	}
    return false
}

//go:inline
func FdClear(fd int, fdSet *syscall.FdSet) {
	pos := fd / 32
	idx := fd % 32
	fdSet.Bits[pos] &= ^(1 << idx)
}

//go:inline
func FdSet(fd int, fdSet *syscall.FdSet) {
	pos := fd / 32
	idx := fd % 32
	fdSet.Bits[pos] |= 1 << idx
}

type NetPollApi struct {
	R syscall.FdSet
	W syscall.FdSet
}

//initializing the platform supported net poll
func GetNetPollApi(eventLoopApi *EventLoopApi) (*NetPollApi, error) {
	return new(NetPollApi), nil
}

func (p *NetPollApi) Resize(size int) error {
	if size >= syscall.FD_SETSIZE {
		return OverflowError
	}
	return nil
}

func (p *NetPollApi) AddEvent(eventLoopApi *EventLoopApi, fd int, mask uint8) error {
    if fd >= syscall.FD_SETSIZE {
    	return OverflowError
	}
	if mask & READ_EVENT > 0 {
		FdSet(fd, &p.R)
	}
	if mask & WRITE_EVENT > 0 {
		FdSet(fd, &p.W)
	}
	return nil
}

func (p *NetPollApi) DeleteEvent(eventLoopApi *EventLoopApi, fd int, delMask uint8) error {
	if fd >= syscall.FD_SETSIZE {
		return OverflowError
	}
	if delMask & READ_EVENT > 0 {
		FdClear(fd, &p.R)
	}
	if delMask & WRITE_EVENT > 0 {
		FdClear(fd, &p.W)
	}
	return nil
}

func (p *NetPollApi) Loop(eventLoopApi *EventLoopApi, msec *int) (int, error) {
	var r syscall.FdSet
	var w syscall.FdSet
	var firedEventNumber int
	r = p.R
	w = p.W
	var timeout *syscall.Timeval = nil
	if msec != nil {
		timeout = &syscall.Timeval{
			Sec: int64((*msec)/1000),
			Usec: int32(((*msec)%1000) * 1000),
		}
	}

	err := syscall.Select(eventLoopApi.MaxFd+1, &r, &w, nil, timeout)
    if err != nil {
    	//注意 系统信号对系统调用的影响
		if er, ok := err.(syscall.Errno); ok {
			if er.Temporary() || er.Timeout() {
				return 0, nil
			}
		} else {
			//不可恢复的系统错误,记录系统日志
			panic(err)
		}
	}
	for i := 0; i <= eventLoopApi.MaxFd; i++ {
		var event uint8
		if FdIsSet(i, &r) {
			event |= READ_EVENT
		}
		if FdIsSet(i, &w) {
			event |= WRITE_EVENT
		}
		if event > 0 {
			eventLoopApi.FiredEvent[firedEventNumber].Fd = i
			eventLoopApi.FiredEvent[firedEventNumber].Events = event
			firedEventNumber++
		}
	}
	return firedEventNumber, nil
}

func (p *NetPollApi) Close() {
	//nothing to do
}

func (p *NetPollApi) GetApiName() string {
	return "select"
}










