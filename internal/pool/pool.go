package pool

import "sync"

type Resettable interface {
	Reset()
}

type Pool[T Resettable] struct {
	pool sync.Pool
}

func New[T Resettable](newFn func() T) *Pool[T] {
	p := &Pool[T]{}

	p.pool.New = func() any {
		return newFn()
	}

	return p
}

func (p *Pool[T]) Get() T {
	v := p.pool.Get()

	return v.(T)
}

func (p *Pool[T]) Put(obj T) {
	if any(obj) == nil {
		return
	}

	obj.Reset()
	p.pool.Put(obj)
}
