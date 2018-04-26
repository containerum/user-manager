package handlers

import (
	"net/http"
	"strconv"

	"strings"

	ch "git.containerum.net/ch/kube-client/pkg/cherry"
	"git.containerum.net/ch/kube-client/pkg/cherry/adaptors/gonic"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/user-manager"
	umtypes "git.containerum.net/ch/user-manager/pkg/models"
	m "git.containerum.net/ch/user-manager/pkg/router/middleware"
	"git.containerum.net/ch/user-manager/pkg/server"
	"git.containerum.net/ch/user-manager/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func UserCreateHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request umtypes.RegisterRequest

	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	errs := validation.ValidateUserCreateRequest(request)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	resp, err := um.CreateUser(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableCreateUser(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

func UserInfoGetHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	resp, err := um.GetUserInfo(ctx.Request.Context())
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableGetUserInfo(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func UserGetByIDHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	resp, err := um.GetUserInfoByID(ctx.Request.Context(), ctx.Param("user_id"))
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableGetUserInfo(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func UserGetByLoginHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	resp, err := um.GetUserInfoByLogin(ctx.Request.Context(), ctx.Param("login"))
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableGetUserInfo(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func UserInfoUpdateHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var newData map[string]interface{}
	if err := ctx.ShouldBindWith(&newData, binding.JSON); err != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	errs := validation.ValidateUserData(newData)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	resp, err := um.UpdateUser(ctx.Request.Context(), newData)
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableUpdateUserInfo(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

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
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableGetUsersList(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func UserListLoginID(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	resp, err := um.GetUsersLoginID(ctx.Request.Context())
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableGetUsersList(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func PartialDeleteHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	err := um.PartiallyDeleteUser(ctx.Request.Context())
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableDeleteUser(), ctx)
		}
		return
	}

	ctx.Status(http.StatusAccepted)
}

func CompleteDeleteHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request umtypes.UserLogin
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	errs := validation.ValidateUserID(request)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	err := um.CompletelyDeleteUser(ctx.Request.Context(), request.ID)
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableDeleteUser(), ctx)
		}
		return
	}

	ctx.Status(http.StatusAccepted)
}
