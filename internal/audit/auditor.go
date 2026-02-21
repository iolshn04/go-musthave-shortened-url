package audit

import "sync"

type Auditor struct {
	observers []Observer
	mu        sync.RWMutex
}

func NewAuditor() *Auditor {
	return &Auditor{}
}

func (a *Auditor) Register(o Observer) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.observers = append(a.observers, o)
}

func (a *Auditor) Notify(e Event) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, o := range a.observers {
		go o.Notify(e)
	}
}
