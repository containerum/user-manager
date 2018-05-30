package handlers

import (
	"net/http"

	kube_types "github.com/containerum/kube-client/pkg/model"

	m "git.containerum.net/ch/user-manager/pkg/router/middleware"
	"git.containerum.net/ch/user-manager/pkg/server"
	"git.containerum.net/ch/user-manager/pkg/umErrors"
	"git.containerum.net/ch/user-manager/pkg/validation"
	"github.com/containerum/cherry"
	"github.com/containerum/cherry/adaptors/gonic"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// swagger:operation POST /groups UserGroups CreateGroupHandler
// Add domain to blacklist.
// https://ch.pages.containerum.net/api-docs/modules/user-manager/index.html#add-domain-to-blacklist
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
//    description: group created
//  default:
//    $ref: '#/responses/error'
func CreateGroupHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request kube_types.UserGroup
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	if errs := validation.ValidateCreateGroup(request); errs != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	err := um.CreateGroup(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			//TODO
			gonic.Gonic(umErrors.ErrUnableBlacklistDomain(), ctx)
		}
		return
	}

	ctx.Status(http.StatusAccepted)
}

// swagger:operation GET /user/info UserInfo UserInfoGetHandler
// Get user info.
// https://ch.pages.containerum.net/api-docs/modules/user-manager/index.html#get-profile-info
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
// responses:
//  '202':
//    description: user info
//    schema:
//      $ref: '#/definitions/User'
//  default:
//    $ref: '#/responses/error'
func GetGroupHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	resp, err := um.GetGroup(ctx.Request.Context(), ctx.Param("group"))
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
