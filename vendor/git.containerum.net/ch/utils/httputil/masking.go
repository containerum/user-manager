package httputil

import (
	"github.com/gin-gonic/gin"

	"git.containerum.net/ch/api-gateway/pkg/utils/headers"
)

type Masker interface {
	Mask()
}

func MaskForNonAdmin(ctx *gin.Context, m Masker) {
	if ctx.GetHeader(headers.UserRoleXHeader) != "admin" {
		m.Mask()
	}
}
