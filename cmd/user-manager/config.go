package main

import (
	"errors"

	"git.containerum.net/ch/user-manager/pkg/server"
	"git.containerum.net/ch/user-manager/pkg/server/impl"

	"fmt"

	"git.containerum.net/ch/user-manager/pkg/clients"
	"git.containerum.net/ch/user-manager/pkg/db"
	"git.containerum.net/ch/user-manager/pkg/db/postgres"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	serviceClientHTTP  = "http"
	serviceClientDummy = "dummy"
)

const (
	portFlag              = "port"
	debugFlag             = "debug"
	textlogFlag           = "textlog"
	dbFlag                = "db"
	dbPGLoginFlag         = "db_pg_login"
	dbPGPasswordFlag      = "db_pg_password"
	dbPGAddrFlag          = "db_pg_addr"
	dbPGNameFlag          = "db_pg_dbname"
	dbPGNoSSLFlag         = "db_pg_nossl"
	dbMigrationsFlag      = "db_migrations"
	mailFlag              = "mail"
	mailURLFlag           = "mail_url"
	recaptchaFlag         = "recaptcha"
	recaptchaKeyFlag      = "recaptcha_key"
	oauthClientsFlag      = "oauth_clients"
	authFlag              = "auth"
	authHTTPAddrFlag      = "auth_http_addr"
	permissionsFlag       = "permissions"
	permissionsURLFlag    = "permissions_url"
	eventsFlag            = "events"
	eventsURLFlag         = "events_url"
	telegramFlag          = "telegram"
	telegramBotIDFlag     = "telegram_bot_id"
	telegramBotTokenFlag  = "telegram_bot_token"
	telegramBotChatIDFlag = "telegram_bot_chat_id"
	umFlag                = "user_manager"
	corsFlag              = "cors"
	adminPwdFlag          = "admin_password"
)

var flags = []cli.Flag{
	cli.StringFlag{
		EnvVar: "CH_USER_PORT",
		Name:   portFlag,
		Value:  "8111",
		Usage:  "port for solutions server",
	},
	cli.BoolFlag{
		EnvVar: "CH_USER_DEBUG",
		Name:   debugFlag,
		Usage:  "start the server in debug mode",
	},
	cli.BoolFlag{
		EnvVar: "CH_USER_TEXTLOG",
		Name:   textlogFlag,
		Usage:  "output log in text format",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_DB",
		Name:   dbFlag,
		Value:  "postgres",
		Usage:  "DB for project",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_PG_LOGIN",
		Name:   dbPGLoginFlag,
		Usage:  "DB Login (PostgreSQL)",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_PG_PASSWORD",
		Name:   dbPGPasswordFlag,
		Usage:  "DB Password (PostgreSQL)",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_PG_ADDR",
		Name:   dbPGAddrFlag,
		Usage:  "DB Address (PostgreSQL)",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_PG_DBNAME",
		Name:   dbPGNameFlag,
		Usage:  "DB name (PostgreSQL)",
	},
	cli.BoolFlag{
		EnvVar: "CH_USER_PG_NOSSL",
		Name:   dbPGNoSSLFlag,
		Usage:  "DB disable ssl (PostgreSQL)",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_MIGRATIONS_PATH",
		Name:   dbMigrationsFlag,
		Value:  "../../pkg/migrations/",
		Usage:  "Location of DB migrations",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_MAIL",
		Name:   mailFlag,
		Value:  serviceClientHTTP,
		Usage:  "Mail-Templater kind",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_MAIL_URL",
		Name:   mailURLFlag,
		Value:  "http://mail-templater:7070",
		Usage:  "Mail-Templater URL",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_RECAPTCHA",
		Name:   recaptchaFlag,
		Value:  serviceClientHTTP,
		Usage:  "Recaptcha kind",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_RECAPTCHA_KEY",
		Name:   recaptchaKeyFlag,
		Usage:  "Recaptcha key",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_OAUTH_CLIENTS",
		Name:   oauthClientsFlag,
		Value:  serviceClientHTTP,
		Usage:  "Recaptcha kind",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_AUTH",
		Name:   authFlag,
		Value:  serviceClientHTTP,
		Usage:  "Auth client kind",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_AUTH_HTTP_ADDR",
		Name:   authHTTPAddrFlag,
		Usage:  "Auth HTTP server address",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_PERMISSIONS",
		Name:   permissionsFlag,
		Value:  serviceClientHTTP,
		Usage:  "Permissions service kind",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_PERMISSIONS_URL",
		Name:   permissionsURLFlag,
		Value:  "http://permissions:4242",
		Usage:  "Permissions service URL",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_EVENTS",
		Name:   eventsFlag,
		Value:  serviceClientHTTP,
		Usage:  "Events-API service kind",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_EVENTS_URL",
		Name:   eventsURLFlag,
		Value:  "http://events-api:1667",
		Usage:  "Events-API service URL",
	},
	cli.BoolFlag{
		EnvVar: "CH_USER_TELEGRAM",
		Name:   telegramFlag,
		Usage:  "Telegram client enabled",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_TELEGRAM_BOT_ID",
		Name:   telegramBotIDFlag,
		Usage:  "Telegram bot ID",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_TELEGRAM_BOT_TOKEN",
		Name:   telegramBotTokenFlag,
		Usage:  "Telegram bot token",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_TELEGRAM_BOT_CHAT_ID",
		Name:   telegramBotChatIDFlag,
		Usage:  "Telegram bot chat ID",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_USER_MANAGER",
		Name:   umFlag,
		Value:  "impl",
		Usage:  "Resource service kind",
	},
	cli.BoolFlag{
		EnvVar: "CH_USER_CORS",
		Name:   corsFlag,
		Usage:  "enable CORS",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_ADMIN_PASSWORD",
		Name:   adminPwdFlag,
		Usage:  "Admin password",
	},
}

func setupLogs(c *cli.Context) {
	if c.Bool("debug") {
		gin.SetMode(gin.DebugMode)
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		gin.SetMode(gin.ReleaseMode)
		logrus.SetLevel(logrus.InfoLevel)
	}

	if c.Bool("textlog") {
		logrus.SetFormatter(&logrus.TextFormatter{})
	} else {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
}

func getDB(c *cli.Context) (db.DB, error) {
	switch c.String(dbFlag) {
	case "postgres":
		url := fmt.Sprintf("postgres://%v:%v@%v/%v", c.String(dbPGLoginFlag), c.String(dbPGPasswordFlag), c.String(dbPGAddrFlag), c.String(dbPGNameFlag))
		if c.Bool(dbPGNoSSLFlag) {
			url = url + "?sslmode=disable"
		}
		return postgres.DBConnect(url, c.String(dbMigrationsFlag))
	default:
		return nil, errors.New("invalid db")
	}
}

func getMailClient(c *cli.Context) (clients.MailClient, error) {
	switch c.String(mailFlag) {
	case serviceClientHTTP:
		return clients.NewHTTPMailClient(c.String(mailURLFlag)), nil
	default:
		return nil, errors.New("invalid mail client")
	}
}

func getReCaptchaClient(c *cli.Context) (clients.ReCaptchaClient, error) {
	switch c.String(recaptchaFlag) {
	case serviceClientHTTP:
		return clients.NewHTTPReCaptchaClient(c.String(recaptchaKeyFlag)), nil
	case serviceClientDummy:
		return clients.NewDummyReCaptchaClient(), nil
	default:
		return nil, errors.New("invalid reCaptcha client")
	}
}

func oauthClientsSetup(c *cli.Context) error {
	switch c.String(oauthClientsFlag) {
	case serviceClientHTTP:
		clients.RegisterOAuthClient(clients.NewGithubOAuthClient())
		clients.RegisterOAuthClient(clients.NewGoogleOAuthClient())
		clients.RegisterOAuthClient(clients.NewFacebookOAuthClient())
	default:
		return errors.New("invalid oauth clients kind")
	}
	return nil
}

func getAuthClient(c *cli.Context) (clients.AuthClient, error) {
	switch c.String(authFlag) {
	case serviceClientHTTP:
		return clients.NewHTTPAuthClient(c.String(authHTTPAddrFlag))
	default:
		return nil, errors.New("invalid auth client")
	}
}

func getPermissionsClient(c *cli.Context) (clients.PermissionsClient, error) {
	switch c.String(permissionsFlag) {
	case serviceClientHTTP:
		return clients.NewHTTPPermissionsClient(c.String(permissionsURLFlag)), nil
	default:
		return nil, errors.New("invalid permissions client")
	}
}

func getTelegramClient(c *cli.Context) (clients.TelegramClient, error) {
	if c.Bool(telegramFlag) {
		return clients.NewTelegramClient(c.String(telegramBotIDFlag), c.String(telegramBotTokenFlag), c.String(telegramBotChatIDFlag))
	}
	return nil, nil
}

func getEventsClient(c *cli.Context) (clients.EventsClient, error) {
	switch c.String(eventsFlag) {
	case serviceClientHTTP:
		return clients.NewHTTPEventsClient(c.String(eventsURLFlag)), nil
	default:
		return nil, errors.New("invalid events-api client")
	}
}

func getUserManager(c *cli.Context, services server.Services) (server.UserManager, error) {
	switch c.String(umFlag) {
	case "impl":
		return impl.NewUserManagerImpl(services), nil
	default:
		return nil, errors.New("invalid user manager impl")
	}
}
