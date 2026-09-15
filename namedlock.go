package fynekit

import (
	"sync"
)

type Namedlock struct {
	mutex   sync.Mutex
	holders map[string]*_NamedlockHolder
}

type Namedlocker interface {
	Lock()
	TryLock() bool
	Unlock()
	RLock()
	TryRLock() bool
	RUnlock()
}

func NewNamedlock() *Namedlock {
	return &Namedlock{
		holders: make(map[string]*_NamedlockHolder),
	}
}

func (n *Namedlock) Locker(name string) Namedlocker {
	return &_Namedlocker{
		owner: n,
		name:  name,
	}
}

func (n *Namedlock) holder(name string) *_NamedlockHolder {
	n.mutex.Lock()
	h := n.holders[name]
	n.mutex.Unlock()
	return h
}

func (n *Namedlock) increase(name string) *_NamedlockHolder {
	n.mutex.Lock()
	h := n.holders[name]
	if h == nil {
		h = &_NamedlockHolder{
			owner: n,
			name:  name,
		}
		n.holders[name] = h
	}
	h.count++
	n.mutex.Unlock()
	return h
}

type _NamedlockHolder struct {
	owner *Namedlock
	name  string
	count int
	mutex sync.RWMutex
}

func (h *_NamedlockHolder) decrease() {
	h.owner.mutex.Lock()
	h.count--
	if h.count <= 0 {
		delete(h.owner.holders, h.name)
	}
	h.owner.mutex.Unlock()
}

type _Namedlocker struct {
	owner *Namedlock
	name  string
}

func (l *_Namedlocker) Lock() {
	// increase count, get holder
	h := l.owner.increase(l.name)
	// lock
	h.mutex.Lock()
}

func (l *_Namedlocker) TryLock() bool {
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

func (l *_Namedlocker) Unlock() {
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

func (l *_Namedlocker) RLock() {
	// increase count, get holder
	h := l.owner.increase(l.name)
	// rlock
	h.mutex.RLock()
}

func (l *_Namedlocker) TryRLock() bool {
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

func (l *_Namedlocker) RUnlock() {
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
