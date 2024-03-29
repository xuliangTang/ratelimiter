package main

import (
	"github.com/gin-gonic/gin"
	"ratelimiter"
)

func main() {
	r := gin.New()

	r.GET("/", ratelimiter.ParamLimiter(3, 1, "limit")(ratelimiter.Limiter(10, 1)(func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"msg": "ok"})
	})))

	r.GET("/byip", ratelimiter.IPLimiter(3, 1)(func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"msg": "by ip"})
	}))

	r.Run(":8081")
}
