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

// swagger:operation GET /domain DomainBlacklist BlacklistDomainsListGetHandler
// Get blacklisted domains list.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
// responses:
//  '200':
//    description: blacklisted domains
//    schema:
//      $ref: '#/definitions/DomainListResponse'
//  default:
//    $ref: '#/responses/error'
func BlacklistDomainsListGetHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	resp, err := um.GetBlacklistedDomainsList(ctx.Request.Context())
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umerrors.ErrUnableGetDomainBlacklist(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// swagger:operation GET /domain/{domain} DomainBlacklist BlacklistDomainGetHandler
// Check if domain is in blacklist.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
//  - name: domain
//    in: path
//    type: string
//    required: true
// responses:
//  '200':
//    description: blacklisted domain
//    schema:
//      $ref: '#/definitions/Domain'
//  default:
//    $ref: '#/responses/error'
func BlacklistDomainGetHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	resp, err := um.GetBlacklistedDomain(ctx.Request.Context(), ctx.Param("domain"))
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umerrors.ErrUnableGetDomainBlacklist(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// swagger:operation POST /domain DomainBlacklist BlacklistDomainAddHandler
// Add domain to blacklist.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/Domain'
// responses:
//  '202':
//    description: domain added to blacklist
//  default:
//    $ref: '#/responses/error'
func BlacklistDomainAddHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request models.Domain
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umerrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	errs := validation.ValidateDomain(request)
	if errs != nil {
		gonic.Gonic(umerrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	err := um.AddDomainToBlacklist(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umerrors.ErrUnableBlacklistDomain(), ctx)
		}
		return
	}

	ctx.Status(http.StatusAccepted)
}

// swagger:operation DELETE /domain/{domain} DomainBlacklist BlacklistDomainDeleteHandler
// Remove domain from blacklist.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
//  - name: domain
//    in: path
//    type: string
//    required: true
// responses:
//  '202':
//    description: domain removed from blacklist
//  default:
//    $ref: '#/responses/error'
func BlacklistDomainDeleteHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	err := um.RemoveDomainFromBlacklist(ctx.Request.Context(), ctx.Param("domain"))
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umerrors.ErrUnableUnblacklistDomain(), ctx)
		}
		return
	}

	ctx.Status(http.StatusAccepted)
}
