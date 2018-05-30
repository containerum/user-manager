package main

import (
	"errors"

	"fmt"

	"git.containerum.net/ch/user-manager/pkg/clients"
	"git.containerum.net/ch/user-manager/pkg/db"
	"git.containerum.net/ch/user-manager/pkg/db/postgres"
	"git.containerum.net/ch/user-manager/pkg/server"
	"git.containerum.net/ch/user-manager/pkg/server/impl"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	serviceClientHTTP  = "http"
	serviceClientDummy = "dummy"
)

const (
	portFlag         = "port"
	debugFlag        = "debug"
	textlogFlag      = "textlog"
	dbFlag           = "db"
	dbPGLoginFlag    = "db_pg_login"
	dbPGPasswordFlag = "db_pg_password"
	dbPGAddrFlag     = "db_pg_addr"
	dbPGNameFlag     = "db_pg_dbname"
	dbPGNoSSLFlag    = "db_pg_nossl"
	dbMigrationsFlag = "db_migrations"
	mailFlag         = "mail"
	mailURLFlag      = "mail_url"
	recaptchaFlag    = "racaptcha"
	recaptchaKeyFlag = "recaptcha_key"
	oauthClientsFlag = "oauth_clients"
	authFlag         = "auth"
	authGRPCAddrFlag = "auth_grpc_addr"
	resourceFlag     = "resource_service"
	resourceURLFlag  = "resource_service_url"
	umFlag           = "user_manager"
	corsFlag         = "cors"
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
		Value:  "grpc",
		Usage:  "Recaptcha kind",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_AUTH_GRPC_ADDR",
		Name:   authGRPCAddrFlag,
		Usage:  "Recaptcha key",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_RESOURCE_SERVICE",
		Name:   resourceFlag,
		Value:  serviceClientHTTP,
		Usage:  "Resource service kind",
	},
	cli.StringFlag{
		EnvVar: "CH_USER_RESOURCE_SERVICE_URL",
		Name:   resourceURLFlag,
		Value:  "http://resource-service:1213",
		Usage:  "MResource service URL",
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

func getAuthClient(c *cli.Context) (clients.AuthClientCloser, error) {
	switch c.String(authFlag) {
	case "grpc":
		return clients.NewGRPCAuthClient(c.String(authGRPCAddrFlag))
	default:
		return nil, errors.New("invalid auth client")
	}
}

func getResourceServiceClient(c *cli.Context) (clients.ResourceServiceClient, error) {
	switch c.String(resourceFlag) {
	case serviceClientHTTP:
		return clients.NewHTTPResourceServiceClient(c.String(resourceURLFlag)), nil
	default:
		return nil, errors.New("invalid resource service client")
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
