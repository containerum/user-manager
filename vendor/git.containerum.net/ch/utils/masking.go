package utils

import (
	"github.com/gin-gonic/gin"

	"git.containerum.net/ch/json-types/misc"
	umtypes "git.containerum.net/ch/json-types/user-manager"
)

func MaskForNonAdmin(ctx *gin.Context, m misc.Masker) {
	if ctx.GetHeader(umtypes.UserRoleHeader) != "admin" {
		m.Mask()
	}
}
