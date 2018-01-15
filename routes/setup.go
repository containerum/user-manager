package routes

import (
	"git.containerum.net/ch/user-manager/server"
	"github.com/gin-gonic/gin"
)

var srv server.UserManager

func SetupRoutes(app *gin.Engine, server server.UserManager) {
	srv = server

	app.Use(prepareContext)

	root := app.Group("/")
	{
		root.POST("/logout/:token_id", logoutHandler)
	}

	user := app.Group("/user")
	{
		user.PUT("/info", userInfoUpdateHandler)

		user.POST("/sign_up", userCreateHandler)
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
		login.POST("/basic", basicLoginHandler)
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
