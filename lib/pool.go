package lib

import "gopkg.in/mcuadros/go-syslog.v2/format"

type WorkerPool struct {
	pool chan Queue

	listenerHandler func(q Queue)
	queueHandler    func(p format.LogParts) Queue
}

type Queue struct {
	ParserResult Log
	FinderResult *GeoFinderResult
}

type WorkerConfig struct {
	// слушает очередь готовых логов к отправке в influx
	ListenerHandler func(q Queue)
	// обрабатывает логи в новых очередях
	QueueHandler func(p format.LogParts) Queue
}

func NewWorkerPool(config WorkerConfig) WorkerPool {
	wp := WorkerPool{}
	wp.pool = make(chan Queue, 1)

	wp.listenerHandler = config.ListenerHandler
	wp.queueHandler = config.QueueHandler

	return wp
}

func (w WorkerPool) Queue(parts format.LogParts) {
	l := w.queueHandler(parts)
	w.pool <- l
}

func (w WorkerPool) Listen() {
	for log := range w.pool {
		w.listenerHandler(log)
	}
}
