package lib

import "gopkg.in/mcuadros/go-syslog.v2/format"

type (
	Pool struct {
		pool     chan Receiver
		taskPool chan func() (Receiver, error)
		listener func(q Receiver)
		receiver func(p format.LogParts) (Receiver, error)
	}
	PoolConfig struct {
		ListenerCallback func(q Receiver)
		ReceiverCallback func(p format.LogParts) (Receiver, error)
		MaxParallel      int
	}
	Receiver struct {
		Parser Log
		Finder *GeoFinderResult
	}
)

func NewPool(c PoolConfig) *Pool {
	p := new(Pool)
	p.pool = make(chan Receiver, c.MaxParallel)
	p.taskPool = make(chan func() (Receiver, error), c.MaxParallel)
	p.listener = c.ListenerCallback
	p.receiver = c.ReceiverCallback

	return p
}

func (p Pool) Task(parts format.LogParts) {
	p.taskPool <- func() (Receiver, error) {
		return p.receiver(parts)
	}
}

func (p Pool) Listen() {
	go p.worker()
	go p.sender()
}

func (p Pool) send(r Receiver) {
	p.pool <- r
}

func (p Pool) sender() {
	for log := range p.pool {
		p.listener(log)
	}
}

func (p Pool) worker() {
	for task := range p.taskPool {
		if receive, err := task(); err == nil {
			p.send(receive)
		}
	}
}
