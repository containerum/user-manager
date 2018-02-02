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
		user.POST("/sign_up", userCreateHandler)
		user.POST("/sign_up/resend", linkResendHandler)
		user.POST("/activation", activateHandler)
		user.POST("/delete/partial", requireIdentityHeaders, requireUserExist, partialDeleteHandler)
		user.POST("/delete/complete", requireIdentityHeaders, requireAdminRole, completeDeleteHandler)

		user.GET("/info/id/:user_id", userGetByIDHandler)
		user.GET("/info/login/:login", userGetByLoginHandler)
		user.GET("/info", requireIdentityHeaders, requireUserExist, userInfoGetHandler)
		user.PUT("/info", userInfoUpdateHandler)

		user.GET("/users", requireIdentityHeaders, requireAdminRole, userListGetHandler)

		user.GET("/links/:user_id", requireIdentityHeaders, requireAdminRole, linksGetHandler)

		user.GET("/bound_accounts", requireIdentityHeaders, requireUserExist, getBoundAccountsHandler)
		user.POST("/bound_accounts", requireIdentityHeaders, requireUserExist, addBoundAccountHandler)
		user.DELETE("/bound_accounts", requireIdentityHeaders, requireUserExist, deleteBoundAccountHandler)

		user.GET("/blacklist", requireIdentityHeaders, requireAdminRole, blacklistGetHandler)
		user.POST("/blacklist", requireIdentityHeaders, requireAdminRole, userToBlacklistHandler)
		user.DELETE("/blacklist", requireIdentityHeaders, requireAdminRole, userDeleteFromBlacklistHandler)

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

	domainBlacklist := app.Group("/domain")
	{
		domainBlacklist.POST("/", requireIdentityHeaders, requireAdminRole, blacklistDomainAddHandler)

		domainBlacklist.GET("/", requireIdentityHeaders, requireAdminRole, blacklistDomainsListGetHandler)
		domainBlacklist.GET("/:domain", requireIdentityHeaders, requireAdminRole, blacklistDomainGetHandler)

		domainBlacklist.DELETE("/:domain", requireIdentityHeaders, requireAdminRole, blacklistDomainDeleteHandler)
	}
}
