package ratelimiter

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Bucket struct {
	mu      sync.Mutex
	tokens  int64     // 令牌数量
	cap     int64     // 桶容量
	limiter Limit     // 速率 (+n/s)
	last    time.Time // 最后一次请求的时间
}

type Reservation struct {
	ok        bool      // 是否允许请求
	timeToAct time.Time // 需要等待的时间
}

func NewBucket(cap, limiter int64) *Bucket {
	if cap <= 0 {
		panic("error cap")
	}
	bucket := &Bucket{cap: cap, tokens: cap, limiter: Limit(limiter)}
	return bucket
}

// InfDuration is the duration returned by Delay when a Reservation is not OK.
const InfDuration = time.Duration(1<<63 - 1)

type Limit int64

// 计算恢复指定token数量，所需要等待的时间
func (limit Limit) durationFromTokens(tokens int64) time.Duration {
	if limit <= 0 {
		return InfDuration
	}
	seconds := float64(tokens / int64(limit))
	return time.Duration(float64(time.Second) * seconds)
}

// Allow 检查是否允许请求
func (this *Bucket) Allow() bool {
	return this.reserveN(time.Now(), 1).ok
}

// 根据当前时间和请求令牌数量，构建请求结果对象
func (this *Bucket) reserveN(now time.Time, n int64) Reservation {
	this.mu.Lock()
	defer this.mu.Unlock()

	// 计算令牌数量并减去这次消耗的数量
	tokens := this.advance(now)
	tokens -= n

	// 计算请求需要等待的时间
	var waitDuration time.Duration
	if tokens < 0 {
		waitDuration = this.limiter.durationFromTokens(-tokens)
	}

	// 构建返回对象
	ok := n <= this.cap && waitDuration <= 0
	r := Reservation{
		ok:        ok,
		timeToAct: now.Add(waitDuration),
	}

	// 更新状态
	if ok {
		this.last = now
		this.tokens = tokens
	}

	return r
}

// Wait 等待拥有足够的令牌以执行操作
func (this *Bucket) Wait(ctx context.Context) (err error) {
	return this.waitN(ctx, 1)
}

func (this *Bucket) waitN(ctx context.Context, n int64) (err error) {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 计算等待时间
	now := time.Now()
	r := this.reserveN(now, n)
	delay := r.delayFrom(now)
	if delay == 0 {
		return nil
	}
	fmt.Println("延迟：", delay)

	// 创建定时器等待
	t := time.NewTimer(delay)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// 根据上次请求间隔计算新的令牌数量
func (this *Bucket) advance(now time.Time) (newTokens int64) {
	elapsed := now.Sub(this.last)
	delta := int64(elapsed.Seconds() * float64(this.limiter))
	tokens := this.tokens + delta
	if tokens > this.cap {
		tokens = this.cap
	}
	return tokens
}

// 计算请求需要等待的秒数
func (r *Reservation) delayFrom(now time.Time) time.Duration {
	delay := r.timeToAct.Sub(now)
	if delay < 0 {
		return 0
	}
	return delay
}
