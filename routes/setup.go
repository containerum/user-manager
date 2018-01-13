package routes

import (
	"git.containerum.net/ch/grpc-proto-files/auth"
	"git.containerum.net/ch/json-types/errors"
	"git.containerum.net/ch/user-manager/clients"
	"git.containerum.net/ch/user-manager/models"
	"github.com/gin-gonic/gin"
)

type Services struct {
	MailClient      clients.MailClient
	DB              models.DB
	AuthClient      auth.AuthClient
	ReCaptchaClient clients.ReCaptchaClient
	WebAPIClient    *clients.WebAPIClient
}

// for response with http.InternalServerError
var (
	userGetFailed              = errors.New("get user from db failed")
	linkGetFailed              = errors.New("get user link from db failed")
	linkCreateFailed           = errors.New("link creation failed")
	emailSendFailed            = errors.New("email send failed")
	tokenCreateFailed          = errors.New("token create failed")
	getTokenFailed             = errors.New("get token from db failed")
	oauthUserInfoGetFailed     = errors.New("get user info over oauth failed")
	boundAccountsGetFailed     = errors.New("get user bound accounts from db failed")
	bindAccountFailed          = errors.New("bind account failed")
	reCaptchaRequestFailed     = errors.New("reCaptcha check request failed")
	deleteTokenFailed          = errors.New("delete token failed")
	userUpdateFailed           = errors.New("update user in db failed")
	oneTimeTokenDeleteFailed   = errors.New("one-time token delete failed")
	blacklistDomainCheckFailed = errors.New("check if domain blacklisted failed")
	userCreateFailed           = errors.New("create user in db failed")
	profileGetFailed           = errors.New("get profile failed")
	blacklistUserFailed        = errors.New("user blacklisting failed")
	blacklistUsersGetFailed    = errors.New("get blacklisted users from db failed")
)

var svc Services

func SetupRoutes(app *gin.Engine, services Services) {
	svc = services

	root := app.Group("/")
	{
		root.POST("/logout/:token_id", logoutHandler)
	}

	user := app.Group("/user")
	{
		user.PUT("/info", userInfoUpdateHandler)

		user.POST("/sign_up", reCaptchaMiddleware, userCreateHandler)
		user.POST("/sign_up/resend", linkResendHandler)
		user.POST("/activation", activateHandler)
		user.POST("/blacklist", adminAccessMiddleware, userToBlacklistHandler)
		user.POST("/delete/partial", partialDeleteHandler)
		user.POST("/delete/complete", adminAccessMiddleware, completeDeleteHandler)

		user.GET("/links/:user_id", adminAccessMiddleware, linksGetHandler)
		user.GET("/blacklist", adminAccessMiddleware, blacklistGetHandler)
		user.GET("/info", userInfoGetHandler)
		user.GET("/users", adminAccessMiddleware, userListGetHandler)
	}

	login := app.Group("/login")
	{
		login.POST("/basic", reCaptchaMiddleware, basicLoginHandler)
		login.POST("/token", oneTimeTokenLoginHandler)
		login.POST("/oauth", oauthLoginHandler)
		login.POST("", webAPILoginHandler) // login through old web-api
	}

	password := app.Group("/password")
	{
		password.PUT("/change", passwordChangeHandler)

		password.POST("/reset", passwordResetHandler)
		password.POST("/restore", passwordRestoreHandler)
	}
}
