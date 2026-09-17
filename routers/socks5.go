package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func Socks5(r *gin.Engine) {
	group := r.Group("/api/v1/settings/socks5")
	group.Use(middlewares.AuthToken)
	{
		group.GET("", api.GetSocks5Settings)
		group.POST("", middlewares.DemoModeRestrict, api.UpdateSocks5Settings)
		group.POST("/stop", middlewares.DemoModeRestrict, api.StopSocks5)
		group.GET("/status", api.GetSocks5Status)
		group.GET("/routing", api.GetSocks5RoutingSnapshot)
		group.POST("/routing/reset", middlewares.DemoModeRestrict, api.ResetSocks5RuntimeStats)
		group.GET("/profiles", api.ListSocks5RoutingProfiles)
		group.POST("/profiles", middlewares.DemoModeRestrict, api.CreateSocks5RoutingProfile)
		group.PUT("/profiles/:id", middlewares.DemoModeRestrict, api.UpdateSocks5RoutingProfile)
		group.DELETE("/profiles/:id", middlewares.DemoModeRestrict, api.DeleteSocks5RoutingProfile)
		group.POST("/health/probe", middlewares.DemoModeRestrict, api.ProbeSocks5Health)
		group.GET("/connections", api.GetSocks5Connections)
		group.DELETE("/connections", middlewares.DemoModeRestrict, api.DeleteSocks5Connections)
		group.DELETE("/connections/:id", middlewares.DemoModeRestrict, api.DeleteSocks5Connection)
	}
}
