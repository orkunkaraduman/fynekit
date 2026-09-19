package fynekit

import (
	"sync"
)

type namedLock struct {
	mutex   sync.Mutex
	holders map[string]*_namedLockHolder
}

type namedLocker interface {
	Lock()
	TryLock() bool
	Unlock()
	RLock()
	TryRLock() bool
	RUnlock()
}

func newNamedLock() *namedLock {
	return &namedLock{
		holders: make(map[string]*_namedLockHolder),
	}
}

func (l *namedLock) Locker(name string) namedLocker {
	return &_namedLocker{
		owner: l,
		name:  name,
	}
}

func (l *namedLock) holder(name string) *_namedLockHolder {
	l.mutex.Lock()
	h := l.holders[name]
	l.mutex.Unlock()
	return h
}

func (l *namedLock) increase(name string) *_namedLockHolder {
	l.mutex.Lock()
	h := l.holders[name]
	if h == nil {
		h = &_namedLockHolder{
			owner: l,
			name:  name,
		}
		l.holders[name] = h
	}
	h.count++
	l.mutex.Unlock()
	return h
}

type _namedLockHolder struct {
	owner *namedLock
	name  string
	count int
	mutex sync.RWMutex
}

func (h *_namedLockHolder) decrease() {
	h.owner.mutex.Lock()
	h.count--
	if h.count <= 0 {
		delete(h.owner.holders, h.name)
	}
	h.owner.mutex.Unlock()
}

type _namedLocker struct {
	owner *namedLock
	name  string
}

func (l *_namedLocker) Lock() {
	// increase count, get holder
	h := l.owner.increase(l.name)
	// lock
	h.mutex.Lock()
}

func (l *_namedLocker) TryLock() bool {
	// increase count, get holder
	h := l.owner.increase(l.name)
	// try  lock
	ok := h.mutex.TryLock()
	if !ok {
		// decrease count, delete holder if needed
		h.decrease()
	}
	return ok
}

func (l *_namedLocker) Unlock() {
	// get holder
	h := l.owner.holder(l.name)
	if h == nil {
		panic("unlock of unlocked mutex")
	}
	// unlock
	h.mutex.Unlock()
	// decrease count, delete holder
	h.decrease()
}

func (l *_namedLocker) RLock() {
	// increase count, get holder
	h := l.owner.increase(l.name)
	// rlock
	h.mutex.RLock()
}

func (l *_namedLocker) TryRLock() bool {
	// increase count, get holder
	h := l.owner.increase(l.name)
	// try  rlock
	ok := h.mutex.TryRLock()
	if !ok {
		// decrease count, delete holder if needed
		h.decrease()
	}
	return ok
}

func (l *_namedLocker) RUnlock() {
	// get holder
	h := l.owner.holder(l.name)
	if h == nil {
		panic("runlock of unlocked mutex")
	}
	// runlock
	h.mutex.RUnlock()
	// decrease count, delete holder
	h.decrease()
}
