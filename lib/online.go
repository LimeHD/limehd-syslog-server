package lib

import (
	"crypto/md5"
	"encoding/hex"
	"sync"
	"time"
)

type (
	Online struct {
		mt               *sync.RWMutex
		connections      map[string]ChannelConnections
		lastFlushedAt    int64
		duration         int64
		scheduleCallback func(*Online)
	}

	OnlineConfig struct {
		// 5 минутные интевалы, настраиваются из вне
		OnlineDuration   int64
		ScheduleCallback func(*Online)
	}

	ChannelConnections struct {
		connections map[string]bool
	}

	UniqueCombination struct {
		Ip        string
		UserAgent string
	}

	// уникальные пользователи на конкретный канал
	// определяются из хеша комбинаций ip и user-agent
	UniqueIdentity struct {
		Channel string
		UniqueCombination
	}
)

func NewOnline(config OnlineConfig) Online {
	o := Online{}
	o.mt = &sync.RWMutex{}
	o.setFlushedAt()
	o.connections = map[string]ChannelConnections{}
	o.duration = config.OnlineDuration
	o.scheduleCallback = config.ScheduleCallback

	return o
}

func (o *Online) Add(i UniqueIdentity) {
	o.mt.Lock()
	// смотрим существует ли канал в мапе, т.к. мы не знаем о канал ничего
	if _, ok := o.connections[i.Channel]; !ok {
		o.connections[i.Channel] = ChannelConnections{
			connections: map[string]bool{},
		}
	}
	o.connections[i.Channel].connections[i.hash()] = true
	o.mt.Unlock()
}

// сбрасываем аккумулированные данные
func (o *Online) Flush() {
	o.mt.Lock()
	o.connections = make(map[string]ChannelConnections)
	o.setFlushedAt()
	o.mt.Unlock()
}

func (o Online) Connections() map[string]ChannelConnections {
	o.mt.RLock()
	defer o.mt.RUnlock()
	return o.connections
}

func (o Online) Count() int {
	o.mt.RLock()
	defer o.mt.RUnlock()
	return len(o.Connections())
}

// существует ли данный пользователь для данного канала
func (o Online) Contains(i UniqueIdentity) bool {
	exist := false

	o.mt.RLock()
	// если канал не существует, то и пользователя не существует
	if c, channelExist := o.connections[i.Channel]; channelExist {
		_, exist = c.connections[i.hash()]
	}
	o.mt.RUnlock()

	return exist
}

// внутренний планировщик для отправки данны в influx
// можно было бы и циклом
func (o *Online) Scheduler() {
schedule:
	time.Sleep(time.Second * time.Duration(o.duration))

	o.scheduleCallback(o)

	goto schedule
}

// смотрит пользователя, если нет - добавляет нового уникального
func (o Online) Peek(i UniqueIdentity) {
	if !o.Contains(i) {
		o.Add(i)
	}
}

// deprecated в пользу планировщика
// настало ли время сбросить данные
// todo можно сделать как scheduler в отдельной гоурутине
// todo с интевалом равным аргументу -online-duration
func (o Online) IsExpiredFlush() bool {
	return time.Now().Unix() >= o.lastFlushedAt+o.duration
}

// private
// метка последнего сброса данных
func (o *Online) setFlushedAt() {
	o.lastFlushedAt = time.Now().Unix()
}

// определяет хеш для определения уникальности поступившего запроса
func (u UniqueIdentity) hash() string {
	ipAgent := u.Ip + u.UserAgent

	hasher := md5.Sum([]byte(ipAgent))
	return hex.EncodeToString(hasher[:])
}

// количество активных соединений
func (c ChannelConnections) Count() int {
	return len(c.connections)
}
