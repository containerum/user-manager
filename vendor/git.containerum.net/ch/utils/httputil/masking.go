package httputil

import (
	"github.com/gin-gonic/gin"

	umtypes "git.containerum.net/ch/json-types/user-manager"
)

type Masker interface {
	Mask()
}

func MaskForNonAdmin(ctx *gin.Context, m Masker) {
	if ctx.GetHeader(umtypes.UserRoleHeader) != "admin" {
		m.Mask()
	}
}
