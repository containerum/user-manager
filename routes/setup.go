package routes

import (
	umtypes "git.containerum.net/ch/json-types/user-manager"
	"git.containerum.net/ch/user-manager/server"
	"git.containerum.net/ch/utils"
	"github.com/gin-gonic/gin"
)

var srv server.UserManager

// SetupRoutes sets up http router needed to handle requests from clients.
func SetupRoutes(app *gin.Engine, server server.UserManager) {
	srv = server

	app.Use(prepareContext)
	app.Use(utils.SaveHeaders)

	requireIdentityHeaders := requireHeaders(umtypes.UserIDHeader, umtypes.UserRoleHeader, umtypes.SessionIDHeader)

	root := app.Group("/")
	{
		root.POST("/logout/:token_id", requireIdentityHeaders, requireUserExist, logoutHandler)
	}

	user := app.Group("/user")
	{
		user.PUT("/info", userInfoUpdateHandler)

		user.POST("/sign_up", userCreateHandler)
		user.POST("/sign_up/resend", linkResendHandler)
		user.POST("/activation", activateHandler)
		user.POST("/blacklist", requireIdentityHeaders, requireAdminRole, userToBlacklistHandler)
		user.POST("/delete/partial", requireIdentityHeaders, requireUserExist, partialDeleteHandler)
		user.POST("/delete/complete", requireIdentityHeaders, requireAdminRole, completeDeleteHandler)
		user.POST("/bound_accounts", requireIdentityHeaders, requireUserExist, addBoundAccountHandler)

		user.GET("/links/:user_id", requireIdentityHeaders, requireAdminRole, linksGetHandler)
		user.GET("/blacklist", requireIdentityHeaders, requireAdminRole, blacklistGetHandler)
		user.GET("/info", requireIdentityHeaders, requireUserExist, userInfoGetHandler)
		user.GET("/users", requireIdentityHeaders, requireAdminRole, userListGetHandler)
		user.GET("/bound_accounts", requireIdentityHeaders, requireUserExist, getBoundAccountsHandler)

		user.DELETE("/bound_accounts", requireIdentityHeaders, requireUserExist, deleteBoundAccountHandler)
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
		password.PUT("/change", requireIdentityHeaders, requireUserExist, passwordChangeHandler)

		password.POST("/reset", passwordResetHandler)
		password.POST("/restore", passwordRestoreHandler)
	}
}
