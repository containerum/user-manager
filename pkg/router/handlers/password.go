package handlers

import (
	"net/http"

	"git.containerum.net/ch/user-manager/pkg/models"
	m "git.containerum.net/ch/user-manager/pkg/router/middleware"
	"git.containerum.net/ch/user-manager/pkg/server"
	"git.containerum.net/ch/user-manager/pkg/umErrors"
	"git.containerum.net/ch/user-manager/pkg/validation"
	"github.com/containerum/cherry"
	"github.com/containerum/cherry/adaptors/gonic"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// swagger:operation POST /password/change Password PasswordChangeHandler
// Change password.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/PasswordChangeRequest'
// responses:
//  '200':
//    description: password changed
//    schema:
//      $ref: '#/definitions/CreateTokenResponse'
//  default:
//    $ref: '#/responses/error'
func PasswordChangeHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request models.PasswordChangeRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	errs := validation.ValidatePasswordChangeRequest(request)
	if errs != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	tokens, err := um.ChangePassword(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableChangePassword(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

// swagger:operation POST /password/reset Password PasswordResetHandler
// Reset password.
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
//    description: password reset link sent
//  default:
//    $ref: '#/responses/error'
func PasswordResetHandler(ctx *gin.Context) {
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

	err := um.ResetPassword(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableResetPassword(), ctx)
		}
		return
	}

	ctx.Status(http.StatusAccepted)
}

// swagger:operation POST /password/restore Password PasswordRestoreHandler
// Change password with token from email.
//
// ---
// x-method-visibility: public
// parameters:
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/PasswordRestoreRequest'
// responses:
//  '200':
//    description: password changed
//    schema:
//      $ref: '#/definitions/CreateTokenResponse'
//  default:
//    $ref: '#/responses/error'
func PasswordRestoreHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request models.PasswordRestoreRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	errs := validation.ValidatePasswordRestoreRequest(request)
	if errs != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	tokens, err := um.RestorePassword(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableResetPassword(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}
