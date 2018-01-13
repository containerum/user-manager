package routes

import (
	"net/http"

	"git.containerum.net/ch/json-types/errors"
	umtypes "git.containerum.net/ch/json-types/user-manager"
	"github.com/gin-gonic/gin"
)

type ReCaptchaRequest struct {
	ReCaptcha string `json:"recaptcha" binding:"required"`
}

const (
	reCaptchaFailed = "reCaptcha failed"
	adminRequired   = "you don`t have access to do this"
)

func reCaptchaMiddleware(ctx *gin.Context) {
	/*var request ReCaptchaRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	checkResp, err := svc.ReCaptchaClient.Check(ctx, ctx.ClientIP(), request.ReCaptcha)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, reCaptchaRequestFailed)
		return
	}

	if !checkResp.Success {
		ctx.AbortWithStatusJSON(http.StatusForbidden, errors.New(reCaptchaFailed))
	}*/
}

func adminAccessMiddleware(ctx *gin.Context) {
	userID := ctx.GetHeader(umtypes.UserIDHeader)
	user, err := svc.DB.GetUserByID(ctx, userID)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}
	if user == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, errors.Format(userWithIDNotFound, userID))
		return
	}
	if user.Role != umtypes.RoleAdmin {
		ctx.AbortWithStatusJSON(http.StatusForbidden, errors.Format(adminRequired))
		return
	}
}
