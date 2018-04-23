package middleware

import (
	"context"

	"git.containerum.net/ch/kube-client/pkg/cherry/adaptors/gonic"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/user-manager"
	umtypes "git.containerum.net/ch/user-manager/pkg/models"
	"git.containerum.net/ch/user-manager/pkg/server"
	"github.com/gin-gonic/gin"
)

var hdrToKey = map[string]interface{}{
	umtypes.UserIDHeader:      server.UserIDContextKey,
	umtypes.UserAgentHeader:   server.UserAgentContextKey,
	umtypes.FingerprintHeader: server.FingerPrintContextKey,
	umtypes.SessionIDHeader:   server.SessionIDContextKey,
	umtypes.TokenIDHeader:     server.TokenIDContextKey,
	umtypes.ClientIPHeader:    server.ClientIPContextKey,
	umtypes.PartTokenIDHeader: server.PartTokenIDContextKey,
}

func RequireHeaders(headers ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var notFoundHeaders []string
		for _, v := range headers {
			if ctx.GetHeader(v) == "" {
				notFoundHeaders = append(notFoundHeaders, v)
			}
		}
		if len(notFoundHeaders) > 0 {
			gonic.Gonic(cherry.ErrRequiredHeadersNotProvided().AddDetails(notFoundHeaders...), ctx)
		}
	}
}

func PrepareContext(ctx *gin.Context) {
	for hn, ck := range hdrToKey {
		if hv := ctx.GetHeader(hn); hv != "" {
			rctx := context.WithValue(ctx.Request.Context(), ck, hv)
			ctx.Request = ctx.Request.WithContext(rctx)
		}
	}
}

// needs role header
func RequireAdminRole(ctx *gin.Context) {
	if ctx.GetHeader(umtypes.UserRoleHeader) != "admin" {
		gonic.Gonic(cherry.ErrAdminRequired(), ctx)
		return
	}

	ump := ctx.MustGet(UMServices).(*server.UserManager)
	um := *ump

	err := um.CheckAdmin(ctx.Request.Context())
	if err != nil {
		gonic.Gonic(cherry.ErrAdminRequired(), ctx)
	}
}

func RequireUserExist(ctx *gin.Context) {
	ump := ctx.MustGet(UMServices).(*server.UserManager)
	um := *ump

	err := um.CheckUserExist(ctx.Request.Context())
	if err != nil {
		gonic.Gonic(cherry.ErrUserNotExist(), ctx)
	}
}
