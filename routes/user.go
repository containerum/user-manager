package routes

import (
	"net/http"

	"time"

	"strings"

	"fmt"

	"git.containerum.net/ch/grpc-proto-files/auth"
	"git.containerum.net/ch/grpc-proto-files/common"
	"git.containerum.net/ch/mail-templater/upstreams"
	"git.containerum.net/ch/user-manager/models"
	"git.containerum.net/ch/user-manager/utils"
	chutils "git.containerum.net/ch/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type UserCreateRequest struct {
	UserName  string `json:"username" binding:"required;email"`
	Password  string `json:"password" binding:"required"`
	Referral  string `json:"referral" binding:"required"`
	ReCaptcha string `json:"recaptcha" binding:"required"`
}

type UserCreateResponse struct {
	ID       string `json:"id"`
	Login    string `json:"login"`
	IsActive bool   `json:"is_active"`
}

type ActivateRequest struct {
	Link string `json:"link" binding:"required"`
}

type ResendLinkRequest struct {
	UserName string `json:"username" binding:"required;email"`
}

type InfoByIDGetResponse struct {
	Login string            `json:"login"`
	Data  map[string]string `json:"data"`
}

type UserToBlacklistRequest struct {
	UserID string `json:"user_id" binding:"required;uuidv4"`
}

type BlacklistedUserEntry struct {
	Login string `json:"login"`
	ID    string `json:"id"`
}

type BlacklistGetResponse struct {
	BlacklistedUsers []BlacklistedUserEntry `json:"blacklist_users"`
}

func userCreateHandler(ctx *gin.Context) {
	var request UserCreateRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, chutils.Error{Text: err.Error()})
		return
	}

	blacklisted, err := svc.DB.IsInBlacklist(strings.Split(request.UserName, "@")[1])
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if blacklisted {
		ctx.AbortWithStatusJSON(http.StatusForbidden, chutils.Error{Text: "user in blacklist"})
		return
	}

	user, err := svc.DB.GetUserByLogin(request.UserName)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if user != nil {
		ctx.AbortWithStatusJSON(http.StatusConflict, chutils.Error{Text: "user already exists"})
		return
	}

	salt := utils.GenSalt(request.UserName, request.Referral)
	passwordHash := utils.GetKey(request.Password, salt)
	newUser := &models.User{
		Login:        request.UserName,
		PasswordHash: passwordHash,
		Salt:         salt,
		Role:         models.RoleUser,
		IsActive:     false,
		IsDeleted:    false,
	}

	var link *models.Link

	err = svc.DB.Transactional(func(tx *models.DB) error {
		if err := svc.DB.CreateUser(newUser); err != nil {
			ctx.Error(err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return err
		}

		if err := svc.DB.CreateProfile(&models.Profile{
			User:      *newUser,
			Referral:  request.Referral,
			Access:    "rw",
			CreatedAt: time.Now().UTC(),
		}); err != nil {
			return err
		}

		link, err = svc.DB.CreateLink(models.LinkTypeConfirm, 24*time.Hour, newUser)
		return err
	})

	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = svc.MailClient.SendConfirmationMail(&upstreams.Recipient{
		ID:        newUser.ID,
		Name:      request.UserName,
		Email:     request.UserName,
		Variables: map[string]string{"CONFIRM": link.Link},
	})
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	err = svc.DB.Transactional(func(tx *models.DB) error {
		link.SentAt = time.Now().UTC()
		return tx.UpdateLink(link)
	})

	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, UserCreateResponse{
		ID:       newUser.ID,
		Login:    newUser.Login,
		IsActive: newUser.IsActive,
	})
}

func linkResendHandler(ctx *gin.Context) {
	var request ResendLinkRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, chutils.Error{Text: err.Error()})
		return
	}

	user, err := svc.DB.GetUserByLogin(request.UserName)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if user == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, chutils.Error{Text: "user " + request.UserName + " not found"})
		return
	}

	link, err := svc.DB.GetLinkForUser(models.LinkTypeConfirm, user)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if link == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, chutils.Error{
			Text: "link type " + models.LinkTypeConfirm + " not found for user " + request.UserName,
		})
		return
	}

	if tdiff := time.Now().UTC().Sub(link.SentAt); tdiff < 5*time.Minute {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, chutils.Error{
			Text: fmt.Sprintf("can`t resend link, wait %f seconds", tdiff.Seconds()),
		})
		return
	}

	if link == nil {
		link, err = svc.DB.CreateLink(models.LinkTypeConfirm, 24*time.Hour, user)
		if err != nil {
			ctx.Error(err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	err = svc.MailClient.SendConfirmationMail(&upstreams.Recipient{
		ID:        user.ID,
		Name:      request.UserName,
		Email:     request.UserName,
		Variables: map[string]string{"CONFIRM": link.Link},
	})
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	err = svc.DB.Transactional(func(tx *models.DB) error {
		link.SentAt = time.Now().UTC()
		return tx.UpdateLink(link)
	})
}

func activateHandler(ctx *gin.Context) {
	var request ActivateRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, chutils.Error{Text: err.Error()})
		return
	}

	link, err := svc.DB.GetLinkFromString(request.Link)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if link == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, chutils.Error{
			Text: "link " + request.Link + " not found",
		})
		return
	}
	if !link.IsActive || time.Now().UTC().After(link.ExpiredAt) {
		ctx.AbortWithStatusJSON(http.StatusGone, chutils.Error{
			Text: "link " + request.Link + " expired",
		})
		return
	}

	// TODO: send request to billing manager

	err = svc.MailClient.SendActivationMail(&upstreams.Recipient{
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

	tokens, err := svc.AuthClient.CreateToken(ctx, &auth.CreateTokenRequest{
		UserAgent:   ctx.Request.UserAgent(),
		UserId:      &common.UUID{Value: link.User.ID},
		UserIp:      ctx.ClientIP(),
		UserRole:    auth.Role_USER,
		RwAccess:    true,
		Access:      &auth.ResourcesAccess{},
		PartTokenId: nil,
	})
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

func infoByIDGetHandler(ctx *gin.Context) {
	userID := ctx.Param("user_id")
	user, err := svc.DB.GetUserByID(userID)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if user == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, chutils.Error{Text: "User with id " + userID + " was not found"})
		return
	}

	profile, err := svc.DB.GetProfileByUser(user)
	if err != nil || profile == nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, &InfoByIDGetResponse{
		Login: user.Login,
		Data:  profile.Data,
	})
}

func userToBlacklistHandler(ctx *gin.Context) {
	var request UserToBlacklistRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, chutils.Error{Text: err.Error()})
		return
	}

	user, err := svc.DB.GetUserByID(request.UserID)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if user == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, chutils.Error{Text: "User with id " + request.UserID + " was not found"})
		return
	}

	profile, err := svc.DB.GetProfileByUser(user)
	if err != nil || profile == nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// TODO: send request to resource manager

	err = svc.MailClient.SendBlockedMail(&upstreams.Recipient{
		ID:    user.ID,
		Name:  user.Login,
		Email: user.Login,
	})
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	user.IsInBlacklist = true
	profile.BlacklistAt = time.Now().UTC()
	err = svc.DB.Transactional(func(tx *models.DB) error {
		err := tx.UpdateUser(user)
		if err != nil {
			return err
		}
		return tx.UpdateProfile(profile)
	})
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusAccepted)
}

func blacklistGetHandler(ctx *gin.Context) {
	blacklisted, err := svc.DB.GetBlacklistedUsers()
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var resp BlacklistGetResponse
	for _, v := range blacklisted {
		resp.BlacklistedUsers = append(resp.BlacklistedUsers, BlacklistedUserEntry{
			Login: v.Login,
			ID:    v.ID,
		})
	}
	ctx.JSON(http.StatusAccepted, resp)
}
