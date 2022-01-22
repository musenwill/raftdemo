package common

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type RWMutex struct {
	sync.RWMutex
	logger  *zap.Logger
	lock    chan string
	unlock  chan bool
	rlock   chan bool
	runlock chan bool
}

func NewRWMutex(logger *zap.Logger) *RWMutex {
	mutex := &RWMutex{
		logger:  logger,
		lock:    make(chan string),
		unlock:  make(chan bool),
		rlock:   make(chan bool),
		runlock: make(chan bool),
	}

	go func() {
		for {
			tag := <-mutex.lock
			timer := time.NewTimer(time.Second * 2)
			select {
			case <-mutex.unlock:
			case <-timer.C:
				mutex.logger.Warn("lock unrealeased after 2 seconds", zap.String("tag", tag))
			}
		}
	}()

	go func() {
		for {
			<-mutex.rlock
			timer := time.NewTimer(time.Second * 2)
			select {
			case <-mutex.runlock:
			case <-timer.C:
				mutex.logger.Warn("rlock unrealeased after 2 seconds")
			}
		}
	}()

	return mutex
}

func (m *RWMutex) Lock(tag string) {
	m.RWMutex.Lock()
	m.lock <- tag
}

func (m *RWMutex) Unlock() {
	m.RWMutex.Unlock()
	m.unlock <- true
}

func (m *RWMutex) RLock() {
	m.RWMutex.RLock()
	m.rlock <- true
}

func (m *RWMutex) RUnlock() {
	m.RWMutex.RUnlock()
	m.runlock <- true
}

type Mutex struct {
	sync.Mutex
	logger *zap.Logger
	lock   chan bool
	unlock chan bool
}

func NewMutex(logger *zap.Logger) *Mutex {
	mutex := &Mutex{
		logger: logger,
		lock:   make(chan bool),
		unlock: make(chan bool),
	}

	go func() {
		for {
			<-mutex.lock
			timer := time.NewTimer(time.Second * 2)
			select {
			case <-mutex.unlock:
			case <-timer.C:
				mutex.logger.Warn("lock unrealeased after 2 seconds")
			}
		}
	}()

	return mutex
}

func (m *Mutex) Lock() {
	m.Mutex.Lock()
	m.lock <- true
}

func (m *Mutex) Unlock() {
	m.Mutex.Unlock()
	m.unlock <- true
}
