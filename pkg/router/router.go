package router

import (
	"net/http"
	"time"

	umtypes "git.containerum.net/ch/json-types/user-manager"
	"git.containerum.net/ch/kube-client/pkg/cherry/adaptors/cherrylog"
	"git.containerum.net/ch/kube-client/pkg/cherry/adaptors/gonic"
	h "git.containerum.net/ch/user-manager/pkg/router/handlers"
	m "git.containerum.net/ch/user-manager/pkg/router/middleware"
	"git.containerum.net/ch/user-manager/pkg/server"
	"git.containerum.net/ch/utils"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	cherry "git.containerum.net/ch/kube-client/pkg/cherry/user-manager"
)

//CreateRouter initialises router and middlewares
func CreateRouter(um *server.UserManager) http.Handler {
	e := gin.New()
	initMiddlewares(e, um)
	initRoutes(e)
	return e
}

func initMiddlewares(e *gin.Engine, um *server.UserManager) {
	/* System */
	e.Use(ginrus.Ginrus(logrus.WithField("component", "gin"), time.RFC3339, true))
	e.Use(gonic.Recovery(cherry.ErrInternalError, cherrylog.NewLogrusAdapter(logrus.WithField("component", "gin"))))
	/* Custom */
	e.Use(m.RegisterServices(um))
	e.Use(m.PrepareContext)
	e.Use(utils.SaveHeaders)
}

// SetupRoutes sets up http router needed to handle requests from clients.
func initRoutes(app *gin.Engine) {
	requireIdentityHeaders := m.RequireHeaders(umtypes.UserIDHeader, umtypes.UserRoleHeader)
	requireLoginHeaders := m.RequireHeaders(umtypes.UserAgentHeader, umtypes.FingerprintHeader, umtypes.ClientIPHeader)
	requireLogoutHeaders := m.RequireHeaders(umtypes.TokenIDHeader, umtypes.SessionIDHeader)

	root := app.Group("/")
	{
		root.POST("/logout", requireLogoutHeaders, m.RequireUserExist, h.LogoutHandler)
	}

	user := app.Group("/user")
	{
		user.POST("/sign_up", requireLoginHeaders, h.UserCreateHandler)
		user.POST("/sign_up/resend", h.LinkResendHandler)
		user.POST("/activation", h.ActivateHandler)
		user.POST("/delete/partial", requireIdentityHeaders, m.RequireUserExist, h.PartialDeleteHandler)
		user.POST("/delete/complete", requireIdentityHeaders, m.RequireAdminRole, h.CompleteDeleteHandler)

		user.GET("/info/id/:user_id", h.UserGetByIDHandler)
		user.GET("/info/login/:login", h.UserGetByLoginHandler)
		user.GET("/info", requireIdentityHeaders, m.RequireUserExist, h.UserInfoGetHandler)
		user.PUT("/info", requireIdentityHeaders, m.RequireUserExist, h.UserInfoUpdateHandler)

		user.GET("/list", requireIdentityHeaders, m.RequireAdminRole, h.UserListGetHandler)
		user.GET("/loginid", requireIdentityHeaders, h.UserListLoginID)

		user.GET("/links/:user_id", requireIdentityHeaders, m.RequireAdminRole, h.LinksGetHandler)

		user.GET("/bound_accounts", requireIdentityHeaders, m.RequireUserExist, h.GetBoundAccountsHandler)
		user.POST("/bound_accounts", requireIdentityHeaders, m.RequireUserExist, h.AddBoundAccountHandler)
		user.DELETE("/bound_accounts", requireIdentityHeaders, m.RequireUserExist, h.DeleteBoundAccountHandler)

		user.GET("/blacklist", requireIdentityHeaders, m.RequireAdminRole, h.BlacklistGetHandler)
		user.POST("/blacklist", requireIdentityHeaders, m.RequireAdminRole, h.UserToBlacklistHandler)
		user.DELETE("/blacklist", requireIdentityHeaders, m.RequireAdminRole, h.UserDeleteFromBlacklistHandler)

	}

	login := app.Group("/login")
	{
		login.POST("/basic", requireLoginHeaders, h.BasicLoginHandler)
		login.POST("/token", requireLoginHeaders, h.OneTimeTokenLoginHandler)
		login.POST("/oauth", requireLoginHeaders, h.OAuthLoginHandler)
	}

	password := app.Group("/password")
	{
		password.PUT("/change", requireIdentityHeaders, m.RequireUserExist, h.PasswordChangeHandler)

		password.POST("/reset", h.PasswordResetHandler)
		password.POST("/restore", h.PasswordRestoreHandler)
	}

	domainBlacklist := app.Group("/domain")
	{
		domainBlacklist.POST("/", requireIdentityHeaders, m.RequireAdminRole, h.BlacklistDomainAddHandler)

		domainBlacklist.GET("/", requireIdentityHeaders, m.RequireAdminRole, h.BlacklistDomainsListGetHandler)
		domainBlacklist.GET("/:domain", requireIdentityHeaders, m.RequireAdminRole, h.BlacklistDomainGetHandler)

		domainBlacklist.DELETE("/:domain", requireIdentityHeaders, m.RequireAdminRole, h.BlacklistDomainDeleteHandler)
	}
}
