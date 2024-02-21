package ratelimiter

import "github.com/gin-gonic/gin"

// Limiter 限流装饰器
func Limiter(cap int) func(handler gin.HandlerFunc) gin.HandlerFunc {
	rl := NewBucket(cap)
	return func(handler gin.HandlerFunc) gin.HandlerFunc {
		return func(c *gin.Context) {
			if rl.Allow() {
				handler(c)
			} else {
				c.AbortWithStatusJSON(429, gin.H{"message": "too many requests"})
			}
		}
	}
}
