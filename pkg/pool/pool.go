package pool

import "sync"

// Resetter ограничивает типы, имеющие метод Reset() для сброса состояния.
type Resetter interface {
	Reset()
}

// Pool — generic-контейнер для хранения и повторного использования объектов с методом Reset().
type Pool[T Resetter] struct {
	pool sync.Pool
}

// New создаёт и возвращает указатель на структуру Pool.
func New[T Resetter](new func() T) *Pool[T] {
	p := &Pool[T]{}
	p.pool.New = func() any {
		return new()
	}
	return p
}

// Get возвращает объект из пула. Если пул пуст, создаёт новый объект с помощью фабрики.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put помещает объект в пул, предварительно вызывая Reset() для сброса состояния.
func (p *Pool[T]) Put(x T) {
	x.Reset()
	p.pool.Put(x)
}
