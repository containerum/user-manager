package routes

import (
	"git.containerum.net/ch/grpc-proto-files/auth"
	"git.containerum.net/ch/user-manager/clients"
	"git.containerum.net/ch/user-manager/models"
	"github.com/gin-gonic/gin"
)

type Services struct {
	MailClient      *clients.MailClient
	DB              *models.DB
	AuthClient      auth.AuthClient
	ReCaptchaClient *clients.ReCaptchaClient
	WebAPIClient    *clients.WebAPIClient
}

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

		user.GET("/:user_id/links/", adminAccessMiddleware, linksGetHandler)
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
