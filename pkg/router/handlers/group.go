package handlers

import (
	"net/http"

	kube_types "github.com/containerum/kube-client/pkg/model"

	"git.containerum.net/ch/user-manager/pkg/models"
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

	ctx.JSON(http.StatusOK, resp)
}

// swagger:operation POST /groups UserGroups CreateGroupHandler
// Create user group.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/UserGroup'
// responses:
//  '201':
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

	ctx.JSON(http.StatusCreated, resp)
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
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/UserGroupMember'
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

	if errs := validation.ValidateUpdateMember(request); errs != nil {
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

	if err := um.UpdateGroupMemberAccess(ctx.Request.Context(), *group, ctx.Param("login"), string(request.Access)); err != nil {
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

// swagger:operation POST /groups/{group}/members UserGroups AddGroupMembersHandler
// Add members to the group.
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
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/UserGroupMembers'
// responses:
//  '202':
//    description: user added
//  default:
//    $ref: '#/responses/error'
func AddGroupMembersHandler(ctx *gin.Context) {
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

// swagger:operation DELETE /groups/{group}/members/{member} UserGroups DeleteGroupMemberHandler
// Remove members from the group.
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
//    description: user removed from the group
//  default:
//    $ref: '#/responses/error'
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

	if err := um.DeleteGroupMember(ctx.Request.Context(), *group, ctx.Param("login")); err != nil {
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

// swagger:operation DELETE /groups/{group} UserGroups DeleteGroupHandler
// Delete user group.
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
//  '202':
//    description: group deleted
//  default:
//    $ref: '#/responses/error'
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

// swagger:operation POST /groups/loginid UserGroups GroupListLabelID
// Get groups list.
//
// ---
// x-method-visibility: public
// parameters:
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/IDList'
// responses:
//  '200':
//    description: groups list
//    schema:
//      $ref: '#/definitions/LoginID'
//  default:
//    $ref: '#/responses/error'
func GroupListLabelID(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	if ctx.Param("group") != "labelid" {
		ctx.Status(http.StatusNotFound)
		return
	}

	var ids models.IDList
	if err := ctx.ShouldBindWith(&ids, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	if len(ids) < 1 {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetails("no group ids in request"), ctx)
		return
	}

	resp, err := um.GetGroupListLabelID(ctx.Request.Context(), ids)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableGetUsersList(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
