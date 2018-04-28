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

// swagger:operation POST /admin/user/sign_up Admin AdminUserCreateHandler
// Create user.
// https://ch.pages.containerum.net/api-docs/modules/user-manager/index.html#admin-create-user
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/UserLogin'
// responses:
//  '201':
//    description: account created
//    schema:
//      $ref: '#/definitions/UserLogin'
//  default:
//    $ref: '#/responses/error'
func AdminUserCreateHandler(ctx *gin.Context) {
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

	resp, err := um.AdminCreateUser(ctx.Request.Context(), request)
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

// swagger:operation POST /admin/user/activation Admin AdminUserActivate
// Activate user.
// https://ch.pages.containerum.net/api-docs/modules/user-manager/index.html#admin-user-activation
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/UserLogin'
// responses:
//  '202':
//    description: user activated
//  default:
//    $ref: '#/responses/error'
func AdminUserActivate(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request models.UserLogin
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	if err := um.AdminActivateUser(ctx.Request.Context(), request); err != nil {
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

// swagger:operation POST /admin/user/deactivation Admin AdminUserDeactivate
// Deactivate user.
// https://ch.pages.containerum.net/api-docs/modules/user-manager/index.html#admin-user-deactivation
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/UserLogin'
// responses:
//  '202':
//    description: user deactivated
//  default:
//    $ref: '#/responses/error'
func AdminUserDeactivate(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request models.UserLogin

	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	err := um.AdminDeactivateUser(ctx.Request.Context(), request)
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

// swagger:operation POST /admin Admin AdminSetAdmin
// Make user admin.
// https://ch.pages.containerum.net/api-docs/modules/user-manager/index.html#set-admin
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/UserLogin'
// responses:
//  '202':
//    description: user becomes admin
//  default:
//    $ref: '#/responses/error'
func AdminSetAdmin(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request models.UserLogin

	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	err := um.AdminSetAdmin(ctx.Request.Context(), request)
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

// swagger:operation DELETE /admin Admin AdminSetAdmin
// Make admin user.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/UserLogin'
// responses:
//  '202':
//    description: admin becomes user
//  default:
//    $ref: '#/responses/error'
func AdminUnsetAdmin(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request models.UserLogin

	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	err := um.AdminUnsetAdmin(ctx.Request.Context(), request)
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

// swagger:operation POST /admin/user/password/reset Admin AdminResetPassword
// Make admin user.
//https://ch.pages.containerum.net/api-docs/modules/user-manager/index.html#admin-reset-password
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/UserLogin'
// responses:
//  '202':
//    description: user new credentials
//    schema:
//      $ref: '#/definitions/UserLogin'
//  default:
//    $ref: '#/responses/error'
func AdminResetPassword(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request models.UserLogin

	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	resp, err := um.AdminResetPassword(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableDeleteUser(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusAccepted, resp)
}
