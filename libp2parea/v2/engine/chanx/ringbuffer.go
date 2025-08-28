package chanx

import (
	"errors"
)

var ErrIsEmpty = errors.New("ringbuffer is empty")

// RingBuffer is a ring buffer for common types.
// It never is full and always grows if it will be full.
// It is not thread-safe(goroutine-safe) so you must use the lock-like synchronization primitive to use it in multiple writers and multiple readers.
type RingBuffer[T any] struct {
	buf         []T
	initialSize int
	size        int
	r           int // read pointer
	w           int // write pointer
	maxQueue    int
}

func NewRingBuffer[T any](initialSize int, maxQueue int) *RingBuffer[T] {
	if initialSize <= 0 {
		panic("initial size must be great than zero")
	}
	// initial size must >= 2
	if initialSize == 1 {
		initialSize = 2
	}

	return &RingBuffer[T]{
		buf:         make([]T, initialSize),
		initialSize: initialSize,
		size:        initialSize,
		maxQueue:    maxQueue,
	}
}

func (r *RingBuffer[T]) Read() (T, error) {
	var t T
	if r.r == r.w {
		return t, ErrIsEmpty
	}
	p := r.r % r.size
	v := r.buf[p]
	r.r++
	// if r.r >= r.size {
	// 	r.r = 0
	// }

	return v, nil
}

func (r *RingBuffer[T]) Pop() T {
	v, err := r.Read()
	if err == ErrIsEmpty { // Empty
		panic(ErrIsEmpty.Error())
	}

	return v
}

func (r *RingBuffer[T]) Peek() T {
	if r.r == r.w { // Empty
		panic(ErrIsEmpty.Error())
	}
	p := r.r % r.size
	v := r.buf[p]
	return v
}

func (r *RingBuffer[T]) Write(v T) {
	p := r.w % r.size
	r.buf[p] = v
	r.w++
	// if r.w >= r.size {
	// 	r.w = 0
	// }

	if r.w == r.size && r.size != r.maxQueue { // full
		r.grow()
	}
}

func (r *RingBuffer[T]) grow() {
	var size int
	if r.size < 1024 {
		size = r.size * 2
	} else {
		size = r.size + r.size/4
	}

	//控制chanx最大buf
	if size > r.maxQueue {
		size = r.maxQueue
	}

	buf := make([]T, size)

	copy(buf[0:], r.buf[r.r:r.w])

	r.w = r.w - r.r
	r.r = 0
	r.size = size
	r.buf = buf
}

func (r *RingBuffer[T]) IsEmpty() bool {
	return r.r == r.w
}

// Capacity returns the size of the underlying buffer.
func (r *RingBuffer[T]) Capacity() int {
	return r.size
}

func (r *RingBuffer[T]) Len() int {
	if r.r == r.w {
		return 0
	}

	if r.w > r.r {
		return r.w - r.r
	}

	return r.size - r.r + r.w
}

func (r *RingBuffer[T]) Reset() {
	r.r = 0
	r.w = 0
	r.size = r.initialSize
	r.buf = make([]T, r.initialSize)
}
