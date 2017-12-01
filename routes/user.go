package routes

import (
	"net/http"

	"time"

	"git.containerum.net/ch/mail-templater/upstreams"
	"git.containerum.net/ch/user-manager/models"
	"git.containerum.net/ch/user-manager/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type UserCreateRequest struct {
	UserName string `json:"username" binding:"required;email"`
	Password string `json:"password" binding:"required"`
	Referral string `json:"referral" binding:"required"`
}

type UserCreateResponse struct {
	ID       string `json:"id"`
	Login    string `json:"login"`
	IsActive bool   `json:"is_active"`
}

func userCreateHandler(ctx *gin.Context) {
	var request UserCreateRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, Error{Error: err.Error()})
		return
	}

	user, err := svc.DB.GetUserByLogin(request.UserName)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if user != nil {
		ctx.AbortWithStatusJSON(http.StatusConflict, Error{Error: "user already exists"})
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
			ctx.Error(err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return err
		}

		link, err = svc.DB.CreateLink(models.LinkTypeConfirm, 24*time.Hour, newUser)
		if err != nil {
			ctx.Error(err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return err
		}
		return nil
	})

	if err == models.ErrTransactionCommit || err == models.ErrTransactionRollback {
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

	ctx.JSON(http.StatusCreated, UserCreateResponse{
		ID:       newUser.ID,
		Login:    newUser.Login,
		IsActive: newUser.IsActive,
	})
}
