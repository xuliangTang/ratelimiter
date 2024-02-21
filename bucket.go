package ratelimiter

import (
	"sync"
	"time"
)

type Bucket struct {
	mu      sync.Mutex
	tokens  int
	cap     int
	limiter int
}

func NewBucket(cap, limiter int) *Bucket {
	if cap <= 0 {
		panic("error cap")
	}
	bucket := &Bucket{cap: cap, tokens: cap, limiter: limiter}
	bucket.start()
	return bucket
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

func (this *Bucket) addToken() {
	this.mu.Lock()
	defer this.mu.Unlock()
	if this.tokens+this.limiter <= this.cap {
		this.tokens += this.limiter
	} else {
		this.tokens = this.cap
	}
}

func (this *Bucket) start() {
	go func() {
		for {
			time.Sleep(time.Second)
			this.addToken()
		}
	}()
}
