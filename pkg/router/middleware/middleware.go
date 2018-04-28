package middleware

import (
	"net/textproto"

	"git.containerum.net/ch/user-manager/pkg/server"
	"git.containerum.net/ch/user-manager/pkg/umErrors"
	"github.com/containerum/cherry/adaptors/gonic"
	headers "github.com/containerum/utils/httputil"
	"github.com/gin-gonic/gin"
)

// RequireAdminRole
func RequireAdminRole(ctx *gin.Context) {
	if ctx.GetHeader(textproto.CanonicalMIMEHeaderKey(headers.UserRoleXHeader)) != "admin" {
		gonic.Gonic(umErrors.ErrAdminRequired(), ctx)
		return
	}

	um := ctx.MustGet(UMServices).(server.UserManager)
	err := um.CheckAdmin(ctx.Request.Context())
	if err != nil {
		gonic.Gonic(umErrors.ErrAdminRequired(), ctx)
	}
}

func RequireUserExist(ctx *gin.Context) {
	um := ctx.MustGet(UMServices).(server.UserManager)
	err := um.CheckUserExist(ctx.Request.Context())
	if err != nil {
		gonic.Gonic(umErrors.ErrUserNotExist(), ctx)
	}
}
