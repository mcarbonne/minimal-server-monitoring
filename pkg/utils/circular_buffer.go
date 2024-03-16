package utils

type CircularBuffer[T any] struct {
	data     []T
	index    int
	size     int
	capacity int
}

func MakeCircularBuffer[T any](capacity int) CircularBuffer[T] {
	if capacity < 1 {
		panic("circular buffer capacity must be strictly positive")
	}
	return CircularBuffer[T]{
		data:     make([]T, capacity),
		index:    0,
		size:     0,
		capacity: capacity,
	}
}

func (cb *CircularBuffer[T]) Push(value T) {
	cb.data[cb.index] = value
	cb.index = (cb.index + 1) % cb.capacity
	if cb.size < cb.capacity {
		cb.size++
	}
}

func (cb *CircularBuffer[T]) Size() int {
	return cb.size
}

func (cb *CircularBuffer[T]) Empty() bool {
	return cb.size == 0
}

func (cb *CircularBuffer[T]) Full() bool {
	return cb.size == cb.capacity
}

func (cb *CircularBuffer[T]) Back() T {
	if cb.Empty() {
		panic("circular buffer is empty")
	}
	return cb.data[PythonMod(cb.index-1, cb.capacity)]
}

func (cb *CircularBuffer[T]) Front() T {
	if cb.Empty() {
		panic("circular buffer is empty")
	}
	if cb.size < cb.capacity {
		return cb.data[0]
	} else {
		return cb.data[cb.index]
	}
}
