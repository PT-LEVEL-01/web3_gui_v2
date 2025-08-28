package token2

import (
	"math"
	"sync"
	"time"
	"web3_gui/chain/config"
)

type Limit float64

const Inf = Limit(math.MaxFloat64)

type Limiter struct {
	mu        sync.Mutex
	limit     Limit
	burst     int     //总量
	tokens    float64 //当前剩余令牌数
	last      time.Time
	lastEvent time.Time
}

func NewLimiter(r Limit, b int) *Limiter {
	return &Limiter{
		limit:  r,
		burst:  b,
		tokens: float64(b),
	}
}

func (lim *Limiter) Recover() bool {
	return lim.RecoverN(1)
}
func (lim *Limiter) RecoverN(n uint64) bool {
	if int(lim.tokens)+int(n) > lim.burst {
		lim.tokens = float64(lim.burst)
	} else {
		lim.tokens += float64(n)
	}
	return true
}

func (lim *Limiter) Allow() bool {
	return lim.AllowN(1)
}

func (lim *Limiter) AllowN(n uint64) bool {
	return lim.reserveN(config.TimeNow(), int(n), 0).ok
}
func (lim *Limiter) advance(t time.Time) (newT time.Time, newTokens float64) {
	last := lim.last
	if t.Before(last) {
		last = t
	}

	//根据时间计算令牌数量
	elapsed := t.Sub(last)
	delta := lim.limit.tokensFromDuration(elapsed)
	tokens := lim.tokens + delta
	if burst := float64(lim.burst); tokens > burst {
		tokens = burst
	}
	return t, tokens
}

func (lim *Limiter) reserveN(t time.Time, n int, maxFutureReserve time.Duration) Reservation {
	lim.mu.Lock()
	defer lim.mu.Unlock()

	if lim.limit == Inf {
		return Reservation{
			ok:        true,
			lim:       lim,
			tokens:    n,
			timeToAct: t,
		}
	} else if lim.limit == 0 {
		var ok bool
		if lim.burst >= n {
			ok = true
			lim.burst -= n
		}
		return Reservation{
			ok:        ok,
			lim:       lim,
			tokens:    lim.burst,
			timeToAct: t,
		}
	}

	t, tokens := lim.advance(t)

	// 计算请求产生的令牌的剩余数量。
	tokens -= float64(n)

	// 计算等待时间
	var waitDuration time.Duration
	if tokens < 0 {
		waitDuration = lim.limit.durationFromTokens(-tokens)
	}

	ok := n <= lim.burst && waitDuration <= maxFutureReserve

	r := Reservation{
		ok:    ok,
		lim:   lim,
		limit: lim.limit,
	}
	if ok {
		r.tokens = n
		r.timeToAct = t.Add(waitDuration)

		lim.last = t
		lim.tokens = tokens
		lim.lastEvent = r.timeToAct
	}

	return r
}

type Reservation struct {
	ok        bool
	lim       *Limiter
	tokens    int
	timeToAct time.Time
	// This is the Limit at reservation time, it can change later.
	limit Limit
}

func (r *Reservation) DelayFrom(t time.Time) time.Duration {
	if !r.ok {
		return InfDuration
	}
	delay := r.timeToAct.Sub(t)
	if delay < 0 {
		return 0
	}
	return delay
}

func (r *Reservation) CancelAt(t time.Time) {
	if !r.ok {
		return
	}

	r.lim.mu.Lock()
	defer r.lim.mu.Unlock()

	if r.lim.limit == Inf || r.tokens == 0 || r.timeToAct.Before(t) {
		return
	}

	// calculate tokens to restore
	// The duration between lim.lastEvent and r.timeToAct tells us how many tokens were reserved
	// after r was obtained. These tokens should not be restored.
	restoreTokens := float64(r.tokens) - r.limit.tokensFromDuration(r.lim.lastEvent.Sub(r.timeToAct))
	if restoreTokens <= 0 {
		return
	}
	// advance time to now
	t, tokens := r.lim.advance(t)
	// calculate new number of tokens
	tokens += restoreTokens
	if burst := float64(r.lim.burst); tokens > burst {
		tokens = burst
	}
	// update state
	r.lim.last = t
	r.lim.tokens = tokens
	if r.timeToAct == r.lim.lastEvent {
		prevEvent := r.timeToAct.Add(r.limit.durationFromTokens(float64(-r.tokens)))
		if !prevEvent.Before(t) {
			r.lim.lastEvent = prevEvent
		}
	}
}

func (r *Reservation) Cancel() {
	r.CancelAt(config.TimeNow())
}

const InfDuration = time.Duration(math.MaxInt64)

func (limit Limit) durationFromTokens(tokens float64) time.Duration {
	if limit <= 0 {
		return InfDuration
	}
	seconds := tokens / float64(limit)
	return time.Duration(float64(time.Second) * seconds)
}

func (limit Limit) tokensFromDuration(d time.Duration) float64 {
	if limit <= 0 {
		return 0
	}
	return d.Seconds() * float64(limit)
}

func (lim *Limiter) Len() uint64 {
	lim.mu.Lock()
	defer lim.mu.Unlock()

	return uint64(lim.burst - int(lim.tokens))
}

func (lim *Limiter) Surplus() uint64 {
	lim.mu.Lock()
	defer lim.mu.Unlock()

	return uint64(lim.tokens)
}
