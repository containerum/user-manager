package routes

import (
	umtypes "git.containerum.net/ch/json-types/user-manager"
	"git.containerum.net/ch/user-manager/server"
	"github.com/gin-gonic/gin"
)

var srv server.UserManager

func SetupRoutes(app *gin.Engine, server server.UserManager) {
	srv = server

	app.Use(prepareContext)

	requireIdentityHeaders := requireHeaders(umtypes.UserIDHeader, umtypes.UserRoleHeader, umtypes.SessionIDHeader)

	root := app.Group("/")
	{
		root.POST("/logout/:token_id", requireIdentityHeaders, logoutHandler)
	}

	user := app.Group("/user")
	{
		user.PUT("/info", userInfoUpdateHandler)

		user.POST("/sign_up", userCreateHandler)
		user.POST("/sign_up/resend", linkResendHandler)
		user.POST("/activation", activateHandler)
		user.POST("/blacklist", requireIdentityHeaders, requireAdminRole, userToBlacklistHandler)
		user.POST("/delete/partial", requireIdentityHeaders, partialDeleteHandler)
		user.POST("/delete/complete", requireIdentityHeaders, requireAdminRole, completeDeleteHandler)

		user.GET("/links/:user_id", requireIdentityHeaders, requireAdminRole, linksGetHandler)
		user.GET("/blacklist", requireIdentityHeaders, requireAdminRole, blacklistGetHandler)
		user.GET("/info", requireIdentityHeaders, userInfoGetHandler)
		user.GET("/users", requireIdentityHeaders, requireAdminRole, userListGetHandler)
	}

	requireLoginHeaders := requireHeaders(umtypes.UserAgentHeader, umtypes.FingerprintHeader, umtypes.ClientIPHeader)
	login := app.Group("/login")
	{
		login.POST("/basic", requireLoginHeaders, basicLoginHandler)
		login.POST("/token", requireLoginHeaders, oneTimeTokenLoginHandler)
		login.POST("/oauth", requireLoginHeaders, oauthLoginHandler)
		login.POST("", requireLoginHeaders, webAPILoginHandler) // login through old web-api
	}

	password := app.Group("/password")
	{
		password.PUT("/change", requireIdentityHeaders, passwordChangeHandler)

		password.POST("/reset", passwordResetHandler)
		password.POST("/restore", passwordRestoreHandler)
	}
}
