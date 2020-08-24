package lib

import (
	"gopkg.in/mcuadros/go-syslog.v2"
	"gopkg.in/mcuadros/go-syslog.v2/format"
)

type (
	Pool struct {
		pool     chan Receiver
		taskPool chan func() (Receiver, error)
		listener func(q Receiver)
		receiver func(p format.LogParts) (Receiver, error)
		workers  int
		senders  int
		workerFn func(pool *Pool, channel syslog.LogPartsChannel)
	}
	PoolConfig struct {
		ListenerCallback func(q Receiver)
		ReceiverCallback func(p format.LogParts) (Receiver, error)
		PoolSize         int
		WorkerPoolSize   int
		WorkersCount     int
		SenderCount      int
		WorkerFn         func(pool *Pool, channel syslog.LogPartsChannel)
	}
	Receiver struct {
		Parser Log
		Finder *GeoFinderResult
	}
)

func NewPool(c PoolConfig) *Pool {
	p := new(Pool)
	p.pool = make(chan Receiver, c.WorkerPoolSize)
	p.taskPool = make(chan func() (Receiver, error), c.PoolSize)
	p.listener = c.ListenerCallback
	p.receiver = c.ReceiverCallback
	p.workers = c.WorkersCount
	p.senders = c.SenderCount
	p.workerFn = c.WorkerFn

	p.listen()

	return p
}

func (p *Pool) SetReceiverCallback(c func(p format.LogParts) (Receiver, error)) {
	p.receiver = c
}

func (p *Pool) SetListenerCallback(c func(q Receiver)) {
	p.listener = c
}

func (p Pool) Run(channel syslog.LogPartsChannel, parallel int) {
	for i := 0; i < parallel; i++ {
		go p.workerFn(&p, channel)
	}
}

func (p Pool) Task(parts format.LogParts) {
	p.taskPool <- func() (Receiver, error) {
		return p.receiver(parts)
	}
}

func (p Pool) listen() {
	for i := 0; i < p.workers; i++ {
		go p.taskManager()
	}
	for i := 0; i < p.workers; i++ {
		go p.sendManager()
	}
}

func (p Pool) send(r Receiver) {
	p.pool <- r
}

func (p Pool) sendManager() {
	for log := range p.pool {
		p.listener(log)
	}
}

func (p Pool) taskManager() {
	for task := range p.taskPool {
		if receive, err := task(); err == nil {
			p.send(receive)
		}
	}
}
