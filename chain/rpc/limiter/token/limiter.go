package token

import (
	"sync"
)

type Limiter struct {
	mu     sync.Mutex
	max    uint64
	tokens uint64
}

func NewLimiter(b uint64) *Limiter {
	return &Limiter{
		max:    b,
		tokens: b,
	}
}

func (l *Limiter) Allow() bool {
	return l.AllowN(1)
}

func (l *Limiter) AllowN(n uint64) bool {
	return l.reserveN(n)
}

func (l *Limiter) reserveN(n uint64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.tokens <= 0 || l.tokens < n {
		return false
	}

	l.tokens -= n

	return true
}

func (l *Limiter) Recover() bool {
	return l.RecoverN(1)
}

func (l *Limiter) RecoverN(n uint64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if n <= 0 || l.tokens+n > l.max {
		return false
	}

	l.tokens += n

	return true
}

func (l *Limiter) Len() uint64 {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.max - l.tokens
}

func (l *Limiter) Surplus() uint64 {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.tokens
}
