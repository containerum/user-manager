package routes

import (
	"net/http"

	"git.containerum.net/ch/auth/storages"
	"git.containerum.net/ch/grpc-proto-files/auth"
	"git.containerum.net/ch/grpc-proto-files/common"
	"git.containerum.net/ch/json-types/errors"
	umtypes "git.containerum.net/ch/json-types/user-manager"
	"github.com/gin-gonic/gin"
)

const (
	tokenNotOwnedByUser = "token %s not owned by user %s"
)

func logoutHandler(ctx *gin.Context) {
	tokenID := ctx.Param("token_id")
	userID := ctx.GetHeader(umtypes.UserIDHeader)
	_, err := svc.AuthClient.DeleteToken(ctx, &auth.DeleteTokenRequest{
		TokenId: &common.UUID{Value: tokenID},
		UserId:  &common.UUID{Value: userID},
	})

	switch err {
	case nil:
	case storages.ErrTokenNotOwnedBySender:
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusForbidden, errors.New(err.Error()))
		return
	case storages.ErrInvalidToken:
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	default:
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, deleteTokenFailed)
		return
	}

	oneTimeToken, err := svc.DB.GetTokenBySessionID(ctx.GetHeader(umtypes.SessionIDHeader))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, getTokenFailed)
		return
	}
	if oneTimeToken != nil {
		if oneTimeToken.User.ID != userID {
			ctx.AbortWithStatusJSON(http.StatusForbidden, errors.Format(tokenNotOwnedByUser, oneTimeToken.Token, userID))
			return
		}
		if err := svc.DB.DeleteToken(oneTimeToken.Token); err != nil {
			ctx.Error(err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, oneTimeTokenDeleteFailed)
			return
		}
	}

	ctx.Status(http.StatusOK)
}
