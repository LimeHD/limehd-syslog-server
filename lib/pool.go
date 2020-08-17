package lib

import "gopkg.in/mcuadros/go-syslog.v2/format"

type (
	Pool struct {
		pool     chan Receiver
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
	p.listener = c.ListenerCallback
	p.receiver = c.ReceiverCallback

	return p
}

func (p Pool) Receive(parts format.LogParts) {
	if i, err := p.receiver(parts); err == nil {
		p.pool <- i
	}
}

func (p Pool) Listen() {
	for log := range p.pool {
		p.listener(log)
	}
}
