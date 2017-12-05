package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type ReCaptchaRequest struct {
	ReCaptcha string `json:"recaptcha" binding:"required"`
}

func reCaptchaMiddleware(ctx *gin.Context) {
	var request ReCaptchaRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	checkResp, err := svc.ReCaptchaClient.Check(ctx.ClientIP(), request.ReCaptcha)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if !checkResp.Success {
		ctx.AbortWithStatusJSON(http.StatusForbidden, Error{Error: "ReCaptcha failed"})
	}
}
