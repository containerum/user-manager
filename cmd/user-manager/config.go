package main

import (
	"errors"

	"git.containerum.net/ch/user-manager/pkg/clients"
	"git.containerum.net/ch/user-manager/pkg/models"
	"git.containerum.net/ch/user-manager/pkg/models/postgres"
	"git.containerum.net/ch/user-manager/pkg/server"
	"git.containerum.net/ch/user-manager/pkg/server/impl"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func setupLogger() error {
	switch gin.Mode() {
	case gin.TestMode, gin.DebugMode:
	case gin.ReleaseMode:
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	viper.SetDefault("log_level", logrus.InfoLevel)
	level := logrus.Level(viper.GetInt("log_level"))
	if level > logrus.DebugLevel || level < logrus.PanicLevel {
		return errors.New("invalid log level")
	}
	logrus.SetLevel(level)
	return nil
}

const (
	serviceClientHTTP  = "http"
	serviceClientDummy = "dummy"
)

func getListenAddr() string {
	viper.SetDefault("listen_addr", ":8111")
	return viper.GetString("listen_addr")
}

func getDB() (models.DB, error) {
	viper.SetDefault("db", "postgres")
	viper.SetDefault("migrations_path", "../../pkg/migrations")

	switch viper.GetString("db") {
	case "postgres":
		return postgres.DBConnect(viper.GetString("pg_url"), viper.GetString("migrations_path"))
	default:
		return nil, errors.New("invalid db")
	}
}

func getMailClient() (clients.MailClient, error) {
	viper.SetDefault("mail", serviceClientHTTP)
	switch viper.GetString("mail") {
	case serviceClientHTTP:
		return clients.NewHTTPMailClient(viper.GetString("mail_url")), nil
	default:
		return nil, errors.New("invalid mail client")
	}
}

func getReCaptchaClient() (clients.ReCaptchaClient, error) {
	viper.SetDefault("recaptcha", serviceClientHTTP)
	switch viper.GetString("recaptcha") {
	case serviceClientHTTP:
		return clients.NewHTTPReCaptchaClient(viper.GetString("recaptcha_key")), nil
	case serviceClientDummy:
		return clients.NewDummyReCaptchaClient(), nil
	default:
		return nil, errors.New("invalid reCaptcha client")
	}
}

func oauthClientsSetup() error {
	viper.SetDefault("oauth_clients", serviceClientHTTP)
	switch viper.GetString("oauth_clients") {
	case serviceClientHTTP:
		clients.RegisterOAuthClient(clients.NewGithubOAuthClient())
		clients.RegisterOAuthClient(clients.NewGoogleOAuthClient())
		clients.RegisterOAuthClient(clients.NewFacebookOAuthClient())
	default:
		return errors.New("invalid oauth clients kind")
	}
	return nil
}

func getWebAPIClient() (clients.WebAPIClient, error) {
	viper.SetDefault("web_api", serviceClientHTTP)
	switch viper.GetString("web_api") {
	case serviceClientHTTP:
		return clients.NewHTTPWebAPIClient(viper.GetString("web_api_url")), nil
	default:
		return nil, errors.New("invalid web_api client")
	}
}

func getAuthClient() (clients.AuthClientCloser, error) {
	viper.SetDefault("auth", "grpc")
	switch viper.GetString("auth") {
	case "grpc":
		return clients.NewGRPCAuthClient(viper.GetString("auth_grpc_addr"))
	default:
		return nil, errors.New("invalid auth client")
	}
}

func getUserManager(services server.Services) (server.UserManager, error) {
	viper.SetDefault("user_manager", "impl")
	switch viper.Get("user_manager") {
	case "impl":
		return impl.NewUserManagerImpl(services), nil
	default:
		return nil, errors.New("invalid user manager impl")
	}
}

func getResourceServiceClient() (clients.ResourceServiceClient, error) {
	viper.SetDefault("resource_service", serviceClientHTTP)
	switch viper.GetString("resource_service") {
	case serviceClientHTTP:
		return clients.NewHTTPResourceServiceClient(viper.GetString("resource_service_url")), nil
	default:
		return nil, errors.New("invalid resource service client")
	}
}
