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
		user.POST("/sign_up", reCaptchaMiddleware, userCreateHandler)
		user.POST("/sign_up/resend", linkResendHandler)
		user.POST("/activation", activateHandler)
		user.POST("/:user_id/blacklist", userToBlacklistHandler)

		user.GET("/:user_id", infoByIDGetHandler)
		user.GET("/:user_id/links", linksGetHandler)
		user.GET("/blacklist", blacklistGetHandler)
	}

	login := app.Group("/login")
	{
		login.POST("/basic", reCaptchaMiddleware, basicLoginHandler)
		login.POST("/token", oneTimeTokenLoginHandler)
		login.POST("/oauth", oauthLoginHandler)
	}

	password := app.Group("/password")
	{
		password.PUT("/change", passwordChangeHandler)

		password.POST("/reset", passwordResetHandler)
		password.POST("/restore", passwordRestoreHandler)
	}
}
