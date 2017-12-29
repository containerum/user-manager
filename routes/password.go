package routes

import (
	"net/http"

	"time"

	"git.containerum.net/ch/grpc-proto-files/auth"
	"git.containerum.net/ch/grpc-proto-files/common"
	"git.containerum.net/ch/json-types/errors"
	mttypes "git.containerum.net/ch/json-types/mail-templater"
	umtypes "git.containerum.net/ch/json-types/user-manager"
	"git.containerum.net/ch/user-manager/models"
	"git.containerum.net/ch/user-manager/utils"
	"github.com/gin-gonic/gin"
)

const (
	invalidPassword    = "invalid password provided"
	linkNotForPassword = "link %s is not for password changing"
	userBanned         = "user %s banned"
)

func passwordChangeHandler(ctx *gin.Context) {
	userID := ctx.GetHeader(umtypes.UserIDHeader)
	var request umtypes.PasswordChangeRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	user, err := svc.DB.GetUserByID(userID)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if user == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, errors.Format(userWithIDNotFound, userID))
		return
	}

	if !utils.CheckPassword(user.Login, request.CurrentPassword, user.Salt, user.PasswordHash) {
		ctx.AbortWithStatusJSON(http.StatusForbidden, errors.New(invalidPassword))
		return
	}

	_, err = svc.AuthClient.DeleteUserTokens(ctx, &auth.DeleteUserTokensRequest{
		UserId: &common.UUID{Value: user.ID},
	})
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	user.PasswordHash = utils.GetKey(user.Login, request.NewPassword, user.Salt)
	err = svc.DB.UpdateUser(user)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = svc.MailClient.SendPasswordChangedMail(&mttypes.Recipient{
		ID:        user.ID,
		Name:      user.Login,
		Email:     user.Login,
		Variables: map[string]string{},
	})
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// TODO: get access from resource manager

	tokens, err := svc.AuthClient.CreateToken(ctx, &auth.CreateTokenRequest{
		UserAgent:   ctx.Request.UserAgent(),
		Fingerprint: ctx.GetHeader(umtypes.FingerprintHeader),
		UserId:      &common.UUID{Value: user.ID},
		UserIp:      ctx.ClientIP(),
		UserRole:    auth.Role(user.Role),
		RwAccess:    true,
		Access:      &auth.ResourcesAccess{},
		PartTokenId: nil,
	})
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusAccepted, tokens)
}

func passwordResetHandler(ctx *gin.Context) {
	userID := ctx.GetHeader(umtypes.UserIDHeader)
	var request umtypes.PasswordChangeRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	user, err := svc.DB.GetUserByID(userID)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if user == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, errors.Format(userWithIDNotFound, userID))
		return
	}
	if user.IsInBlacklist {
		ctx.AbortWithStatusJSON(http.StatusForbidden, errors.Format(userBanned, user.Login))
		return
	}

	var link *models.Link
	err = svc.DB.Transactional(func(tx *models.DB) (err error) {
		link, err = svc.DB.CreateLink(umtypes.LinkTypePwdChange, 24*time.Hour, user)
		return
	})
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = svc.MailClient.SendPasswordResetMail(&mttypes.Recipient{
		ID:        user.ID,
		Name:      user.Login,
		Email:     user.Login,
		Variables: map[string]string{"TOKEN": link.Link},
	})
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}

func passwordRestoreHandler(ctx *gin.Context) {
	var request umtypes.PasswordRestoreRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	link, err := svc.DB.GetLinkFromString(request.Link)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if link == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, errors.Format(linkNotFound, request.Link))
		return
	}
	if link.Type != umtypes.LinkTypePwdChange {
		ctx.AbortWithStatusJSON(http.StatusForbidden, errors.Format(linkNotForPassword, request.Link))
		return
	}

	_, err = svc.AuthClient.DeleteUserTokens(ctx, &auth.DeleteUserTokensRequest{
		UserId: &common.UUID{Value: link.User.ID},
	})
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = svc.DB.Transactional(func(tx *models.DB) error {
		link.User.PasswordHash = utils.GetKey(link.User.Login, request.NewPassword, link.User.Salt)
		if err := svc.DB.UpdateUser(link.User); err != nil {
			return err
		}
		link.IsActive = false
		return svc.DB.UpdateLink(link)
	})

	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = svc.MailClient.SendPasswordChangedMail(&mttypes.Recipient{
		ID:        link.User.ID,
		Name:      link.User.Login,
		Email:     link.User.Login,
		Variables: map[string]string{},
	})
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// TODO: get access from resource manager

	tokens, err := svc.AuthClient.CreateToken(ctx, &auth.CreateTokenRequest{
		UserAgent:   ctx.Request.UserAgent(),
		Fingerprint: ctx.GetHeader(umtypes.FingerprintHeader),
		UserId:      &common.UUID{Value: link.User.ID},
		UserIp:      ctx.ClientIP(),
		UserRole:    auth.Role(link.User.Role),
		RwAccess:    true,
		Access:      &auth.ResourcesAccess{},
		PartTokenId: nil,
	})
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusAccepted, tokens)
}
