package io

import (
	"log"
	"sync/atomic"
)

type PCSSink interface {
	SetReactivePower(q_mvar float64)
	LastQ() float64
}

type LocalPCS struct {
	lastQ atomic.Value // float64
}

func NewLocalPCS() *LocalPCS {
	l := &LocalPCS{}
	l.lastQ.Store(0.0)
	return l
}

func (p *LocalPCS) SetReactivePower(q_mvar float64) {
	p.lastQ.Store(q_mvar)
	// In production, this would call vendor API. Here we log once in a while.
	log.Printf("[PCS] Q_set = %.3f MVAr\n", q_mvar)
}

func (p *LocalPCS) LastQ() float64 {
	v := p.lastQ.Load()
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}
