package lib

import (
	"sync"
	"time"
)

type (
	StreamQueue struct {
		mt               sync.RWMutex
		internal         []InfluxRequestParams
		scheduleCallback func(s *StreamQueue)
	}
)

func NewStream() *StreamQueue {
	s := new(StreamQueue)
	s.internal = []InfluxRequestParams{}

	return s
}

func (s *StreamQueue) SetScheduleHandler(handler func(s *StreamQueue)) {
	s.scheduleCallback = handler
}

func (s *StreamQueue) Add(item InfluxRequestParams) {
	s.mt.Lock()
	s.internal = append(s.internal, item)
	s.mt.Unlock()
}

func (s *StreamQueue) All() []InfluxRequestParams {
	s.mt.Lock()
	defer s.mt.Unlock()
	return s.internal
}

func (s *StreamQueue) Flush() {
	s.internal = []InfluxRequestParams{}
}

func (s *StreamQueue) Scheduler(duration int) {
schedule:
	time.Sleep(time.Second * time.Duration(duration))

	s.scheduleCallback(s)

	goto schedule
}
