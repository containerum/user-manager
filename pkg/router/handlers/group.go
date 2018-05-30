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
	"github.com/containerum/utils/httputil"
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

	if request.UserGroupMembers != nil {
		if errs := validation.ValidateAddMembers(*request.UserGroupMembers); errs != nil {
			gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
			return
		}
	}

	id, err := um.CreateGroup(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableCreateGroup(), ctx)
		}
		return
	}

	resp, err := um.GetGroup(ctx.Request.Context(), *id)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableGetGroup(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusAccepted, resp)
}

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
func AddGroupMemberHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request kube_types.UserGroupMembers
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	if errs := validation.ValidateAddMembers(request); errs != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	group, err := um.GetGroup(ctx.Request.Context(), ctx.Param("group"))
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableGetGroup(), ctx)
		}
		return
	}

	if group.OwnerID != httputil.MustGetUserID(ctx.Request.Context()) {
		gonic.Gonic(umErrors.ErrNotGroupOwner(), ctx)
		return
	}

	err = um.AddGroupMembers(ctx.Request.Context(), ctx.Param("group"), request)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableAddGroupMember(), ctx)
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
			gonic.Gonic(umErrors.ErrUnableGetGroup(), ctx)
		}
		return
	}

	if resp.OwnerID != httputil.MustGetUserID(ctx.Request.Context()) {
		gonic.Gonic(umErrors.ErrNotGroupOwner(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func DeleteGroupMemberHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	group, err := um.GetGroup(ctx.Request.Context(), ctx.Param("group"))
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableGetGroup(), ctx)
		}
		return
	}

	if group.OwnerID != httputil.MustGetUserID(ctx.Request.Context()) {
		gonic.Gonic(umErrors.ErrNotGroupOwner(), ctx)
		return
	}

	if err := um.DeleteGroupMember(ctx.Request.Context(), ctx.Param("group"), ctx.Param("id")); err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableGetGroup(), ctx)
		}
		return
	}

	ctx.Status(http.StatusAccepted)
}

func UpdateGroupMemberHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	group, err := um.GetGroup(ctx.Request.Context(), ctx.Param("group"))
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableGetGroup(), ctx)
		}
		return
	}

	if group.OwnerID != httputil.MustGetUserID(ctx.Request.Context()) {
		gonic.Gonic(umErrors.ErrNotGroupOwner(), ctx)
		return
	}

	if err := um.DeleteGroupMember(ctx.Request.Context(), ctx.Param("group"), ctx.Param("id")); err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableGetGroup(), ctx)
		}
		return
	}

	ctx.Status(http.StatusAccepted)
}
