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

// swagger:operation GET /groups UserGroups GetGroupsListHandler
// Get user groups list.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
// responses:
//  '200':
//    description: groups list
//    schema:
//      $ref: '#/definitions/UserGroups'
//  default:
//    $ref: '#/responses/error'
func GetGroupsListHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	resp, err := um.GetGroupsList(ctx.Request.Context(), httputil.MustGetUserID(ctx.Request.Context()))
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableGetGroup(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// swagger:operation GET /groups/{group} UserGroups GetGroupHandler
// Get user groups list.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
//  - name: group
//    in: path
//    type: string
//    required: true
// responses:
//  '200':
//    description: group
//    schema:
//      $ref: '#/definitions/UserGroup'
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

// swagger:operation POST /groups UserGroups CreateGroupHandler
// Create user group.
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

// swagger:operation POST /groups/{group}/members/{member} UserGroups UpdateGroupMemberHandler
// Change group member access.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
//  - name: group
//    in: path
//    type: string
//    required: true
//  - name: member
//    in: path
//    type: string
//    required: true
// responses:
//  '202':
//    description: user access changed
//  default:
//    $ref: '#/responses/error'
func UpdateGroupMemberHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request kube_types.UserGroupMember
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
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

	if group.OwnerID == ctx.Param("id") {
		gonic.Gonic(umErrors.ErrUnableChangeOwnerPermissions(), ctx)
		return
	}

	if err := um.UpdateGroupMemberAccess(ctx.Request.Context(), ctx.Param("group"), ctx.Param("id"), string(request.Access)); err != nil {
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

	if group.OwnerID == ctx.Param("id") {
		gonic.Gonic(umErrors.ErrUnableRemoveOwner(), ctx)
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

func DeleteGroupHandler(ctx *gin.Context) {
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

	if err := um.DeleteGroup(ctx.Request.Context(), ctx.Param("group")); err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableDeleteGroup(), ctx)
		}
		return
	}

	ctx.Status(http.StatusAccepted)
}
