package lib

import (
	"gopkg.in/mcuadros/go-syslog.v2"
	"gopkg.in/mcuadros/go-syslog.v2/format"
)

type (
	Pool struct {
		// Пул для данных, готовых к отправлению
		pool chan Receiver
		// Пул асихронных задач для формирования пула готовых данных
		taskPool chan func() (Receiver, error)
		// Пул ошибок при формировании или отправке данных, обрабатываются в отдельных потоках
		errorPool chan error
		// Обработчик готовых данных
		listener func(q Receiver) error
		// Обработчик, формирующий данные
		receiver func(p format.LogParts) (Receiver, error)
		// Обработчик полученных ошибок
		errorHandler func(err error)
		// Колиечество паралельнно обрабатывающихся задач
		workers int
		// Колиечство паралельно работающих обработчиков данных (для отправки)
		senders int
		// Колиечство паралельных обработчиков ошибок
		errorHandlers int
		// Обработчик, принимающий данные из вне (по UDP)
		workerFn func(pool *Pool, channel syslog.LogPartsChannel)
	}
	PoolConfig struct {
		ListenerCallback    func(q Receiver) error
		ReceiverCallback    func(p format.LogParts) (Receiver, error)
		ErrorHandleCallback func(err error)
		PoolSize            int
		WorkerPoolSize      int
		ErrorPoolSize       int
		WorkersCount        int
		SenderCount         int
		ErrorHandlerCount   int
		WorkerFn            func(pool *Pool, channel syslog.LogPartsChannel)
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
	p.errorPool = make(chan error, c.ErrorPoolSize)
	p.listener = c.ListenerCallback
	p.receiver = c.ReceiverCallback
	p.errorHandler = c.ErrorHandleCallback
	p.errorHandlers = c.ErrorHandlerCount
	p.workers = c.WorkersCount
	p.senders = c.SenderCount
	p.workerFn = c.WorkerFn

	p.listen()

	return p
}

func (p *Pool) SetReceiverCallback(c func(p format.LogParts) (Receiver, error)) {
	p.receiver = c
}

func (p *Pool) SetListenerCallback(c func(q Receiver) error) {
	p.listener = c
}

// Запускает в несколько потоков обработку входящих данных
func (p Pool) Run(channel syslog.LogPartsChannel, parallel int) {
	for i := 0; i < parallel; i++ {
		go p.workerFn(&p, channel)
	}
}

// Создает новую асихронную задачу
func (p Pool) Task(parts format.LogParts) {
	p.taskPool <- func() (Receiver, error) {
		return p.receiver(parts)
	}
}

// Слушаем все наши потоки данных, задач и ошибок
func (p Pool) listen() {
	for i := 0; i < p.workers; i++ {
		go p.taskManager()
	}
	for i := 0; i < p.workers; i++ {
		go p.sendManager()
	}
	for i := 0; i < p.errorHandlers; i++ {
		go p.errorManager()
	}
}

func (p Pool) send(r Receiver) {
	p.pool <- r
}

func (p Pool) error(err error) {
	p.errorPool <- err
}

func (p Pool) sendManager() {
	for log := range p.pool {
		if err := p.listener(log); err != nil {
			p.error(err)
		}
	}
}

func (p Pool) taskManager() {
	for task := range p.taskPool {
		if receive, err := task(); err == nil {
			p.send(receive)
		} else {
			p.error(err)
		}
	}
}

func (p Pool) errorManager() {
	for err := range p.errorPool {
		p.errorHandler(err)
	}
}
