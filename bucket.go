package ratelimiter

import "sync"

type Bucket struct {
	mu     sync.Mutex
	tokens int
	cap    int
}

func NewBucket(cap int) *Bucket {
	if cap <= 0 {
		panic("error cap")
	}
	return &Bucket{cap: cap, tokens: cap}
}

func (this *Bucket) Allow() bool {
	this.mu.Lock()
	defer this.mu.Unlock()
	if this.tokens > 0 {
		this.tokens--
		return true
	}
	return false
}
