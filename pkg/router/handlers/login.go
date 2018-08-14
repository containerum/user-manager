package handlers

import (
	"net/http"

	"git.containerum.net/ch/user-manager/pkg/models"
	m "git.containerum.net/ch/user-manager/pkg/router/middleware"
	"git.containerum.net/ch/user-manager/pkg/server"
	"git.containerum.net/ch/user-manager/pkg/umerrors"
	"git.containerum.net/ch/user-manager/pkg/validation"
	"github.com/containerum/cherry"
	"github.com/containerum/cherry/adaptors/gonic"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// swagger:operation POST /login/basic Login BasicLoginHandler
// Basic login.
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
//      $ref: '#/definitions/LoginRequest'
// responses:
//  '200':
//    description: user logged in
//    schema:
//      $ref: '#/definitions/CreateTokenResponse'
//  default:
//    $ref: '#/responses/error'
func BasicLoginHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request models.LoginRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umerrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	if errs := validation.ValidateLoginRequest(request); errs != nil {
		gonic.Gonic(umerrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	tokens, err := um.BasicLogin(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umerrors.ErrLoginFailed(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

// swagger:operation POST /login/token Login OneTimeTokenLoginHandler
// Login with one-time token.
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
//      $ref: '#/definitions/OneTimeTokenLoginRequest'
// responses:
//  '200':
//    description: user logged in
//    schema:
//      $ref: '#/definitions/CreateTokenResponse'
//  default:
//    $ref: '#/responses/error'
func OneTimeTokenLoginHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request models.OneTimeTokenLoginRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umerrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	tokens, err := um.OneTimeTokenLogin(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umerrors.ErrLoginFailed(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

// swagger:operation POST /login/oauth Login OAuthLoginRequest
// Login using oauth service.
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
//      $ref: '#/definitions/OAuthLoginRequest'
// responses:
//  '200':
//    description: user logged in
//    schema:
//      $ref: '#/definitions/CreateTokenResponse'
//  default:
//    $ref: '#/responses/error'
func OAuthLoginHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request models.OAuthLoginRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umerrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	if errs := validation.ValidateOAuthLoginRequest(request); errs != nil {
		gonic.Gonic(umerrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	tokens, err := um.OAuthLogin(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umerrors.ErrLoginFailed(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

// swagger:operation POST /logout Login LogoutHandler
// Logout for users who used one-time token login.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/TokenIDHeader'
//  - $ref: '#/parameters/SessionIDHeader'
// responses:
//  '200':
//    description: user logged out
//  default:
//    $ref: '#/responses/error'
func LogoutHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	err := um.Logout(ctx.Request.Context())
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umerrors.ErrLogoutFailed(), ctx)
		}
		return
	}

	ctx.Status(http.StatusOK)
}
