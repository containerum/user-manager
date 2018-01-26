package main

import (
	"fmt"
	"os"

	"time"

	"context"
	"net/http"
	"os/signal"

	"git.containerum.net/ch/user-manager/clients"
	"git.containerum.net/ch/user-manager/models"
	"git.containerum.net/ch/user-manager/routes"
	"git.containerum.net/ch/user-manager/server"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func exitOnErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getService(service interface{}, err error) interface{} {
	exitOnErr(err)
	return service
}

func main() {
	viper.SetEnvPrefix("ch_user")
	viper.AutomaticEnv()
	exitOnErr(setupLogger())

	logrus.Infoln("starting server...")

	app := gin.New()
	app.Use(gin.RecoveryWithWriter(logrus.StandardLogger().WithField("component", "gin_recovery").WriterLevel(logrus.ErrorLevel)))
	app.Use(ginrus.Ginrus(logrus.StandardLogger(), time.RFC3339, true))

	err := oauthClientsSetup()
	exitOnErr(err)

	userManager, err := getUserManager(server.Services{
		MailClient:            getService(getMailClient()).(clients.MailClient),
		DB:                    getService(getDB()).(models.DB),
		AuthClient:            getService(getAuthClient()).(clients.AuthClientCloser),
		ReCaptchaClient:       getService(getReCaptchaClient()).(clients.ReCaptchaClient),
		WebAPIClient:          getService(getWebAPIClient()).(clients.WebAPIClient),
		ResourceServiceClient: getService(getResourceServiceClient()).(clients.ResourceServiceClient),
	})
	exitOnErr(err)
	defer userManager.Close()

	routes.SetupRoutes(app, userManager)

	// graceful shutdown support

	srv := http.Server{
		Addr:    getListenAddr(),
		Handler: app,
	}

	go exitOnErr(srv.ListenAndServe())

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logrus.Infoln("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	exitOnErr(srv.Shutdown(ctx))
}
