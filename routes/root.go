package routes

import (
	"net/http"

	"git.containerum.net/ch/auth/storages"
	"git.containerum.net/ch/grpc-proto-files/auth"
	"git.containerum.net/ch/grpc-proto-files/common"
	"git.containerum.net/ch/utils"
	"github.com/gin-gonic/gin"
)

func logoutHandler(ctx *gin.Context) {
	tokenID := ctx.Param("token_id")
	_, err := svc.AuthClient.DeleteToken(ctx, &auth.DeleteTokenRequest{
		TokenId: &common.UUID{Value: tokenID},
		UserId:  &common.UUID{Value: ctx.GetHeader("X-User-ID")},
	})
	switch err {
	case nil:
	case storages.ErrTokenNotOwnedBySender:
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusForbidden, utils.Error{Text: err.Error()})
		return
	case storages.ErrInvalidToken:
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, utils.Error{Text: err.Error()})
		return
	default:
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	oneTimeToken, err := svc.DB.GetTokenBySessionID(ctx.GetHeader("X-Session-ID")) // TODO: may be other header name
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if oneTimeToken != nil {
		if oneTimeToken.User.ID != ctx.GetHeader("X-User-ID") {
			ctx.AbortWithStatusJSON(http.StatusForbidden, utils.Error{Text: "token not belongs to user"})
			return
		}
		if err := svc.DB.DeleteToken(oneTimeToken.Token); err != nil {
			ctx.Error(err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	ctx.Status(http.StatusOK)
}
