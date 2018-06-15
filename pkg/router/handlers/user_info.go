package handlers

import (
	"net/http"
	"strconv"
	"strings"

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

// swagger:operation GET /user/info UserInfo UserInfoGetHandler
// Get user info.
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
func UserInfoGetHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	resp, err := um.GetUserInfo(ctx.Request.Context())
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

// swagger:operation POST /user/info UserInfo UserInfoUpdateHandler
// Update user info.
//
// ---
// x-method-visibility: private
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/UserData'
// responses:
//  '202':
//    description: updated user info
//    schema:
//      $ref: '#/definitions/User'
//  default:
//    $ref: '#/responses/error'
func UserInfoUpdateHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var newData models.UserData
	if err := ctx.ShouldBindWith(&newData, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	errs := validation.ValidateUserData(newData)
	if errs != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	resp, err := um.UpdateUser(ctx.Request.Context(), newData)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableUpdateUserInfo(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusAccepted, resp)
}

// swagger:operation GET /user/info/id/{user_id} UserInfo UserGetByIDHandler
// Get user info by ID.
//
// ---
// x-method-visibility: private
// parameters:
//  - name: user_id
//    in: path
//    type: string
//    required: true
// responses:
//  '200':
//    description: user info
//    schema:
//      $ref: '#/definitions/User'
//  default:
//    $ref: '#/responses/error'
func UserGetByIDHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	resp, err := um.GetUserInfoByID(ctx.Request.Context(), ctx.Param("user_id"))
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

// swagger:operation GET /user/info/login/{login} UserInfo UserGetByLoginHandler
// Get user info by ID.
//
// ---
// x-method-visibility: private
// parameters:
//  - name: login
//    in: path
//    type: string
//    required: true
// responses:
//  '200':
//    description: user info
//    schema:
//      $ref: '#/definitions/User'
//  default:
//    $ref: '#/responses/error'
func UserGetByLoginHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	resp, err := um.GetUserInfoByLogin(ctx.Request.Context(), ctx.Param("login"))
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

// swagger:operation GET /user/list UserInfo UserListGetHandler
// Get user info.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserIDHeader'
//  - name: page
//    in: query
//    type: string
//    required: false
//  - name: per_page
//    in: query
//    type: string
//    required: false
// responses:
//  '200':
//    description: users list
//    schema:
//      $ref: '#/definitions/UserList'
//  default:
//    $ref: '#/responses/error'
func UserListGetHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	page := int64(1)
	pagestr, ok := ctx.GetQuery("page")
	if ok {
		var err error
		page, err = strconv.ParseInt(pagestr, 10, 64)
		if err != nil {
			ctx.Error(err)
		}
	}

	perPage := int64(10)
	perPagestr, ok := ctx.GetQuery("per_page")
	if ok {
		var err error
		perPage, err = strconv.ParseInt(perPagestr, 10, 64)
		if err != nil {
			ctx.Error(err)
		}
	}

	filters := strings.Split(ctx.Query("filters"), ",")
	resp, err := um.GetUsers(ctx.Request.Context(), int(page), int(perPage), filters...)
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

// swagger:operation POST /user/loginid UserInfo UserListLoginID
// Get users list.
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
//    description: users list
//    schema:
//      $ref: '#/definitions/LoginID'
//  default:
//    $ref: '#/responses/error'
func UserListLoginID(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var ids models.IDList
	if err := ctx.ShouldBindWith(&ids, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	if len(ids) < 1 {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetails("no users ids in request"), ctx)
		return
	}

	resp, err := um.GetUsersLoginID(ctx.Request.Context(), ids)
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
