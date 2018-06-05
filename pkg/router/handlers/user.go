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

// swagger:operation POST /user/sign_up User UserCreateHandler
// Create user.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserAgentHeader'
//  - $ref: '#/parameters/FingerprintHeader'
//  - $ref: '#/parameters/ClientIPHeader'
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/RegisterRequest'
// responses:
//  '200':
//    description: user created
//    schema:
//      $ref: '#/definitions/UserLogin'
//  default:
//    $ref: '#/responses/error'
func UserCreateHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request models.RegisterRequest

	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	errs := validation.ValidateUserCreateRequest(request)
	if errs != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	resp, err := um.CreateUser(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableCreateUser(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// swagger:operation POST /user/activation User ActivateHandler
// Activate user.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/Link'
// responses:
//  '200':
//    description: user activated
//    schema:
//      $ref: '#/definitions/CreateTokenResponse'
//  default:
//    $ref: '#/responses/error'
func ActivateHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request models.Link
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	errs := validation.ValidateLink(request)
	if errs != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	tokens, err := um.ActivateUser(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableActivate(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

// swagger:operation POST /user/delete/partial User PartialDeleteHandler
// Mark user as deleted.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
// responses:
//  '202':
//    description: user deleted
//  default:
//    $ref: '#/responses/error'
func PartialDeleteHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	err := um.PartiallyDeleteUser(ctx.Request.Context())
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableDeleteUser(), ctx)
		}
		return
	}

	ctx.Status(http.StatusAccepted)
}

// swagger:operation POST /user/delete/complete User CompleteDeleteHandler
// Delete user completely (almost).
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserAgentHeader'
//  - $ref: '#/parameters/FingerprintHeader'
//  - $ref: '#/parameters/ClientIPHeader'
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/UserLogin'
// responses:
//  '202':
//    description: user deleted
//  default:
//    $ref: '#/responses/error'
func CompleteDeleteHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request models.UserLogin
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	errs := validation.ValidateUserID(request)
	if errs != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	err := um.CompletelyDeleteUser(ctx.Request.Context(), request.ID)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableDeleteUser(), ctx)
		}
		return
	}

	ctx.Status(http.StatusAccepted)
}
