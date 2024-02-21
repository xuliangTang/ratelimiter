package ratelimiter

import (
	"context"
	"github.com/gin-gonic/gin"
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
