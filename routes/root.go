package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func logoutHandler(ctx *gin.Context) {
	err := srv.Logout(ctx.Request.Context())
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.Status(http.StatusOK)
}
