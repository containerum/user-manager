package main

import (
	"fmt"
	"os"

	"time"

	"context"
	"net/http"
	"os/signal"

	"git.containerum.net/ch/user-manager/pkg/clients"
	"git.containerum.net/ch/user-manager/pkg/db"
	"git.containerum.net/ch/user-manager/pkg/router"
	"git.containerum.net/ch/user-manager/pkg/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//go:generate swagger generate spec -m -i ../../swagger-basic.yml -o ../../swagger.json

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

	err := oauthClientsSetup()
	exitOnErr(err)

	userManager, err := getUserManager(server.Services{
		MailClient:            getService(getMailClient()).(clients.MailClient),
		DB:                    getService(getDB()).(db.DB),
		AuthClient:            getService(getAuthClient()).(clients.AuthClientCloser),
		ReCaptchaClient:       getService(getReCaptchaClient()).(clients.ReCaptchaClient),
		ResourceServiceClient: getService(getResourceServiceClient()).(clients.ResourceServiceClient),
	})
	exitOnErr(err)
	defer userManager.Close()

	app := router.CreateRouter(&userManager)

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
