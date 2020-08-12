package lib

import (
	"crypto/md5"
	"encoding/hex"
	"sync"
	"time"
)

type (
	Online struct {
		mt            *sync.RWMutex
		connections   map[string]ChannelConnections
		lastFlushedAt int64
		duration      int64
	}

	OnlineConfig struct {
		// 5 минутные интевалы, настраиваются из вне
		OnlineDuration int64
	}

	ChannelConnections struct {
		connections map[string]bool
	}

	// уникальные пользователи на конкретный канал
	// определяются из хеша комбинаций ip и user-agent
	UniqueIdentity struct {
		Channel   string
		Ip        string
		UserAgent string
	}
)

func NewOnline(config OnlineConfig) Online {
	o := Online{}
	o.mt = &sync.RWMutex{}
	o.setFlushedAt()
	o.connections = map[string]ChannelConnections{}
	o.duration = config.OnlineDuration

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
	o.connections[i.Channel].connections[o.hash(i.Ip, i.UserAgent)] = true
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
	return len(o.Connections())
}

func (o Online) Contains(i UniqueIdentity) bool {
	exist := false

	o.mt.RLock()
	if c, channelExist := o.connections[i.Channel]; channelExist {
		_, exist = c.connections[o.hash(i.Ip, i.UserAgent)]
	}
	o.mt.RUnlock()

	return exist
}

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
func (o Online) hash(ip, userAgent string) string {
	ipAgent := ip + userAgent
	hasher := md5.Sum([]byte(ipAgent))
	return hex.EncodeToString(hasher[:])
}

//

func (c ChannelConnections) Count() int {
	return len(c.connections)
}
