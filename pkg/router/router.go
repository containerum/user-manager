package router

import (
	"net/http"
	"time"

	h "git.containerum.net/ch/user-manager/pkg/router/handlers"
	m "git.containerum.net/ch/user-manager/pkg/router/middleware"
	"git.containerum.net/ch/user-manager/pkg/server"
	"git.containerum.net/ch/user-manager/pkg/umErrors"
	"git.containerum.net/ch/user-manager/static"
	"github.com/containerum/cherry/adaptors/cherrylog"
	"github.com/containerum/cherry/adaptors/gonic"

	utils "github.com/containerum/utils/httputil"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	headers "github.com/containerum/utils/httputil"
	"gopkg.in/gin-contrib/cors.v1"
)

//CreateRouter initialises router and middlewares
func CreateRouter(um *server.UserManager, enableCORS bool) http.Handler {
	e := gin.New()
	initMiddlewares(e, um, enableCORS)
	initRoutes(e)
	return e
}

func initMiddlewares(e *gin.Engine, um *server.UserManager, enableCORS bool) {
	/* CORS */
	if enableCORS {
		cfg := cors.DefaultConfig()
		cfg.AllowAllOrigins = true
		cfg.AddAllowMethods(http.MethodDelete)
		cfg.AddAllowHeaders(headers.UserRoleXHeader, headers.UserIDXHeader, headers.UserAgentXHeader, headers.UserClientXHeader, headers.UserIPXHeader, headers.TokenIDXHeader, "X-Session-ID")
		e.Use(cors.New(cfg))
	}
	e.Group("/static").
		StaticFS("/", static.HTTP)
	/* System */
	e.Use(ginrus.Ginrus(logrus.WithField("component", "gin"), time.RFC3339, true))
	e.Use(gonic.Recovery(umErrors.ErrInternalError, cherrylog.NewLogrusAdapter(logrus.WithField("component", "gin"))))
	/* Custom */
	e.Use(m.RegisterServices(um))
	e.Use(utils.PrepareContext)
	e.Use(utils.SaveHeaders)
}

// SetupRoutes sets up http router needed to handle requests from clients.
func initRoutes(app *gin.Engine) {
	requireIdentityHeaders := utils.RequireHeaders(umErrors.ErrRequiredHeadersNotProvided, headers.UserIDXHeader, headers.UserRoleXHeader)
	requireLoginHeaders := utils.RequireHeaders(umErrors.ErrRequiredHeadersNotProvided, headers.UserAgentXHeader, headers.UserClientXHeader, headers.UserIPXHeader)
	//TODO
	requireLogoutHeaders := utils.RequireHeaders(umErrors.ErrRequiredHeadersNotProvided, headers.TokenIDXHeader, "X-Session-ID")

	root := app.Group("")
	{
		root.POST("/logout", requireLogoutHeaders, m.RequireUserExist, h.LogoutHandler)
	}

	user := app.Group("/user")
	{
		user.POST("/sign_up", requireLoginHeaders, h.UserCreateHandler)
		user.POST("/sign_up/resend", h.LinkResendHandler)
		user.POST("/activation", requireLoginHeaders, h.ActivateHandler)

		user.GET("/list", requireIdentityHeaders, m.RequireAdminRole, h.UserListGetHandler)
		user.GET("/links/:user_id", requireIdentityHeaders, m.RequireAdminRole, h.LinksGetHandler)

		user.POST("/loginid", h.UserListLoginID)

		deleteuser := user.Group("/delete", requireIdentityHeaders)
		{
			deleteuser.POST("/partial", m.RequireUserExist, h.PartialDeleteHandler)
			deleteuser.POST("/complete", m.RequireAdminRole, h.CompleteDeleteHandler)
		}

		info := user.Group("/info")
		{
			info.GET("/id/:user_id", h.UserGetByIDHandler)
			info.GET("/login/:login", h.UserGetByLoginHandler)
			info.GET("", requireIdentityHeaders, m.RequireUserExist, h.UserInfoGetHandler)
			info.PUT("", requireIdentityHeaders, m.RequireUserExist, h.UserInfoUpdateHandler)
		}

		accounts := user.Group("/bound_accounts", requireIdentityHeaders, m.RequireUserExist)
		{
			accounts.GET("", h.GetBoundAccountsHandler)
			accounts.POST("", h.AddBoundAccountHandler)
			accounts.DELETE("", h.DeleteBoundAccountHandler)
		}

		blacklist := user.Group("/blacklist", requireIdentityHeaders, m.RequireAdminRole)
		{
			blacklist.GET("", h.BlacklistGetHandler)
			blacklist.POST("", h.UserToBlacklistHandler)
			blacklist.DELETE("", h.UserDeleteFromBlacklistHandler)
		}
	}

	login := app.Group("/login", requireLoginHeaders)
	{
		login.POST("/basic", h.BasicLoginHandler)
		login.POST("/token", h.OneTimeTokenLoginHandler)
		login.POST("/oauth", h.OAuthLoginHandler)
	}

	password := app.Group("/password")
	{
		password.POST("/reset", h.PasswordResetHandler)
		password.POST("/restore", h.PasswordRestoreHandler)

		password.PUT("/change", requireIdentityHeaders, m.RequireUserExist, h.PasswordChangeHandler)
	}

	domainBlacklist := app.Group("/domain", requireIdentityHeaders, m.RequireAdminRole)
	{
		domainBlacklist.GET("", h.BlacklistDomainsListGetHandler)
		domainBlacklist.GET("/:domain", h.BlacklistDomainGetHandler)

		domainBlacklist.POST("", h.BlacklistDomainAddHandler)

		domainBlacklist.DELETE("/:domain", h.BlacklistDomainDeleteHandler)
	}

	admin := app.Group("/admin/user", requireIdentityHeaders, m.RequireAdminRole)
	{
		admin.POST("/sign_up", h.AdminUserCreateHandler)
		admin.POST("/activation", h.AdminUserActivateHandler)
		admin.POST("/deactivation", h.AdminUserDeactivateHandler)
		admin.POST("/password/reset", h.AdminResetPasswordHandler)
		admin.POST("", h.AdminSetAdminHandler)

		admin.DELETE("", h.AdminUnsetAdminHandler)
	}

	userGroups := app.Group("/groups", requireIdentityHeaders, m.RequireUserExist)
	{
		userGroups.GET("", h.GetGroupsListHandler)
		userGroups.GET("/:group", m.RequireAdminRole, h.GetGroupHandler)

		userGroups.POST("", m.RequireAdminRole, h.CreateGroupHandler)
		userGroups.POST("/:group/members", m.RequireAdminRole, h.AddGroupMembersHandler)

		userGroups.PUT("/:group/members/:login", m.RequireAdminRole, h.UpdateGroupMemberHandler)

		userGroups.DELETE("/:group/members/:login", m.RequireAdminRole, h.DeleteGroupMemberHandler)
		userGroups.DELETE("/:group", m.RequireAdminRole, h.DeleteGroupHandler)
	}
}
