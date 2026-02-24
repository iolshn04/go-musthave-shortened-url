package audit

import (
	"sync"

	"go.uber.org/zap"
)

type Auditor struct {
	observers []Observer
	mu        sync.RWMutex
	events    chan Event
	log       *zap.Logger
}

func NewAuditor(log *zap.Logger) *Auditor {
	a := &Auditor{
		events: make(chan Event, 100),
		log:    log,
	}
	go a.worker()
	return a
}

func (a *Auditor) worker() {
	for e := range a.events {
		a.mu.RLock()
		observers := make([]Observer, len(a.observers))
		copy(observers, a.observers)
		a.mu.RUnlock()

		for _, o := range observers {
			o.Notify(e)
		}
	}
}

func (a *Auditor) Register(o Observer) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.observers = append(a.observers, o)
}

func (a *Auditor) Notify(e Event) {
	select {
	case a.events <- e:
	default:
		a.log.Warn("audit event dropped: buffer full")
	}
}
