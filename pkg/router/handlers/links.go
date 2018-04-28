package handlers

import (
	"net/http"

	"git.containerum.net/ch/user-manager/pkg/models"
	m "git.containerum.net/ch/user-manager/pkg/router/middleware"
	"git.containerum.net/ch/user-manager/pkg/server"
	"github.com/containerum/cherry"
	"github.com/containerum/cherry/adaptors/gonic"
	"github.com/gin-gonic/gin"

	"git.containerum.net/ch/user-manager/pkg/umErrors"
	"git.containerum.net/ch/user-manager/pkg/validation"
	"github.com/gin-gonic/gin/binding"
)

// swagger:operation POST /user/sign_up/resend Links LinkResendHandler
// Resend activation link.
// https://ch.pages.containerum.net/api-docs/modules/user-manager/index.html#resend-activation-link
//
// ---
// x-method-visibility: public
// parameters:
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/UserLogin'
// responses:
//  '202':
//    description: link sent
//  default:
//    $ref: '#/responses/error'
func LinkResendHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request models.UserLogin
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	errs := validation.ValidateUserLogin(request)
	if errs != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	err := um.LinkResend(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableResendLink(), ctx)
		}
		return
	}

	ctx.Status(http.StatusAccepted)
}

// swagger:operation GET /user/links/{user_id} Links LinksGetHandler
// Get user links.
// https://ch.pages.containerum.net/api-docs/modules/user-manager/index.html#get-user-links
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
//  - name: user_id
//    in: path
//    type: string
//    required: true
// responses:
//  '200':
//    description: link sent
//    schema:
//      $ref: '#/definitions/Links'
//  default:
//    $ref: '#/responses/error'
func LinksGetHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	resp, err := um.GetUserLinks(ctx.Request.Context(), ctx.Param("user_id"))
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableGetUserInfo(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
