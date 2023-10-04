package main

import (
	"math"
	"sync/atomic"
	"time"
)

type proxyInitParams struct {
	readTimeout         time.Duration
	writeTimeout        time.Duration
	maxIdleConnDuration time.Duration
}

// lock-free array
type lfa[T any] struct {
	idx  uint32
	buf  []T
	len_ uint32
}

func (a *lfa[T]) reserve(len_ int) {
	a.len_ = uint32(len_)
	a.idx = math.MaxUint32
	a.buf = make([]T, len_)
}

func (a *lfa[T]) add(x T) {
	a.buf[atomic.AddUint32(&a.idx, 1)%a.len_] = x
}

func (a *lfa[T]) reset() {
	a.idx = math.MaxUint32
}

func (a *lfa[T]) len() int {
	return int(a.len_)
}

func (a *lfa[T]) get(i int) T {
	return a.buf[i]
}
