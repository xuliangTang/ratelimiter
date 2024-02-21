package ratelimiter

import (
	"context"
	"github.com/gin-gonic/gin"
	"sync"
	"time"
)

// Limiter 限流装饰器
func Limiter(cap, limiter int64) func(handler gin.HandlerFunc) gin.HandlerFunc {
	rl := NewBucket(cap, limiter)
	return func(handler gin.HandlerFunc) gin.HandlerFunc {
		return func(c *gin.Context) {
			ctx, _ := context.WithTimeout(c, time.Second*5)
			err := rl.Wait(ctx)
			if err != nil {
				c.AbortWithStatusJSON(400, gin.H{"err": err.Error()})
				return
			}
			handler(c)
			/*if rl.Allow() {
				handler(c)
			} else {
				c.AbortWithStatusJSON(429, gin.H{"message": "too many requests"})
			}*/
		}
	}
}

// ParamLimiter 参数限流
func ParamLimiter(cap, limiter int64, param string) func(handler gin.HandlerFunc) gin.HandlerFunc {
	rl := NewBucket(cap, limiter)
	return func(handler gin.HandlerFunc) gin.HandlerFunc {
		return func(c *gin.Context) {
			if c.Query(param) != "" && !rl.Allow() {
				c.AbortWithStatusJSON(429, gin.H{"message": "too many requests"})
				return
			}

			handler(c)
		}
	}
}

var IpLimiterCache *LimiterCache

type LimiterCache struct {
	data sync.Map // key:ip value:*Bucket
}

func init() {
	IpLimiterCache = &LimiterCache{}
}

// FindOrCreate 根据ip获取bucket对象，没有则创建
func (this *LimiterCache) FindOrCreate(ip string, cap, limiter int64) *Bucket {
	if getBucket, ok := this.data.Load(ip); ok {
		return getBucket.(*Bucket)
	}

	bucket := NewBucket(cap, limiter)
	this.data.Store(ip, bucket)
	return bucket
}

// IPLimiter 根据IP限流
func IPLimiter(cap, limiter int64) func(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(handler gin.HandlerFunc) gin.HandlerFunc {
		return func(c *gin.Context) {
			rl := IpLimiterCache.FindOrCreate(c.Request.RemoteAddr, cap, limiter)
			if !rl.Allow() {
				c.AbortWithStatusJSON(429, gin.H{"message": "too many requests"})
				return
			}

			handler(c)
		}
	}
}
