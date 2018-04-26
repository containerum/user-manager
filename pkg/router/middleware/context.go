package middleware

import (
	"git.containerum.net/ch/user-manager/pkg/server"
	"github.com/gin-gonic/gin"
)

const (
	//UMServices is key for services
	UMServices = "um-service"
)

// RegisterServices adds services to context
func RegisterServices(svc *server.UserManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(UMServices, *svc)
	}
}
