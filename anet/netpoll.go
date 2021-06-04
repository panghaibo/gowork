package anet

import (
	"errors"
	"unsafe"
)

const (
	NONE_EVENT   = 0
	READ_EVENT   = 1 << iota
	WRITE_EVENT  = 1 << iota
	TIME_EVENT   = 1 << iota

	// default poll event size warning select support 1024 default
	DEFAULT_SET_SIZE = 1024
)

var (
    OverflowError = errors.New("overflow net poll size")
)

//event callback function

type Callback func(fd int, event uint8, private unsafe.Pointer)

type Event struct {
	Fd int
	Events uint8
	Private unsafe.Pointer
	RCall Callback
	WCall Callback
}

//fired event
//events is
type FiredEvent struct {
	Fd int
	Events uint8
}

type PollInterface interface {
	AddEvent(eventLoopApi *EventLoopApi, fd int, mask uint8) error
	DeleteEvent(eventLoopApi *EventLoopApi, fd int, delMask uint8) error
	Loop(eventLoopApi *EventLoopApi, msec *int) (int, error)
	Close()
	GetApiName() string
}

type EventLoopApi struct {
	MaxFd int
	Size int
	api PollInterface
	Events []Event
	FiredEvent []FiredEvent
}

func GetEventApi(size int) (*EventLoopApi, error) {
	eventApi := new(EventLoopApi)
	var err error
	eventApi.Size = size
	eventApi.api, err = GetNetPollApi(eventApi)
	if err != nil {
		return nil, err
	}
	eventApi.Events = make([]Event, size)
	eventApi.FiredEvent = make([]FiredEvent, size)
	return eventApi, err
}

func (p *EventLoopApi) AddEvent(fd int, event uint8, rCall Callback, wCall Callback, private unsafe.Pointer) error {
	if fd >= p.Size {
		return OverflowError
	}
	err := p.api.AddEvent(p, fd, event)
	if err != nil {
		return err
	}
	ev := &p.Events[fd]
	if event & READ_EVENT > 0 {
		ev.Events |= READ_EVENT
		ev.RCall = rCall
	}
	if event & WRITE_EVENT > 0 {
		ev.Events |= WRITE_EVENT
		ev.WCall = wCall
	}
	ev.Private = private
	if fd > p.MaxFd {
		p.MaxFd = fd
	}
    return nil
}

func (p *EventLoopApi) DeleteEvent(fd int, event uint8) {
	if fd >= p.Size || p.Events[fd].Events == NONE_EVENT {
		return
	}
	p.api.DeleteEvent(p, fd, event)
	ev := &p.Events[fd]
	ev.Events &= ^event
	if ev.Events == NONE_EVENT {
		ev.Private = nil
	}
	if fd != p.MaxFd || fd == p.MaxFd && ev.Events == NONE_EVENT {
		return
	}
	for p.MaxFd > 0 {
		if p.Events[p.MaxFd].Events == NONE_EVENT {
			p.MaxFd--
		} else {
			break
		}
	}
	return
}

func (p *EventLoopApi) Loop() error {
	for {
		n, err := p.api.Loop(p, nil)
		if err != nil {
			panic(err)
			return err
		}
		for i := 0; i < n; i++ {
			ev := p.FiredEvent[i]
			event := p.Events[ev.Fd]
			if ev.Events & READ_EVENT > 0 {
				event.RCall(ev.Fd, READ_EVENT, event.Private)
			}
			if ev.Events & WRITE_EVENT > 0 {
				event.RCall(ev.Fd, WRITE_EVENT, event.Private)
			}
		}
	}
	return nil
}





