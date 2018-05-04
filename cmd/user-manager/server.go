package main

import (
	"fmt"
	"os"

	"time"

	"context"
	"net/http"
	"os/signal"

	"text/tabwriter"

	"git.containerum.net/ch/user-manager/pkg/clients"
	"git.containerum.net/ch/user-manager/pkg/db"
	"git.containerum.net/ch/user-manager/pkg/router"
	"git.containerum.net/ch/user-manager/pkg/server"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

//go:generate swagger generate spec -m -i ../../swagger-basic.yml -o ../../swagger.json

func initServer(c *cli.Context) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.TabIndent|tabwriter.Debug)
	for _, f := range c.GlobalFlagNames() {
		fmt.Fprintf(w, "Flag: %s\t Value: %s\n", f, c.String(f))
	}
	w.Flush()

	setupLogs(c)

	err := oauthClientsSetup(c)
	exitOnErr(err)

	userManager, err := getUserManager(c, server.Services{
		MailClient:            getService(getMailClient(c)).(clients.MailClient),
		DB:                    getService(getDB(c)).(db.DB),
		AuthClient:            getService(getAuthClient(c)).(clients.AuthClientCloser),
		ReCaptchaClient:       getService(getReCaptchaClient(c)).(clients.ReCaptchaClient),
		ResourceServiceClient: getService(getResourceServiceClient(c)).(clients.ResourceServiceClient),
	})
	exitOnErr(err)
	defer userManager.Close()

	app := router.CreateRouter(&userManager, c.Bool(corsFlag))

	// graceful shutdown support
	srv := http.Server{
		Addr:    ":" + c.String(portFlag),
		Handler: app,
	}

	go exitOnErr(srv.ListenAndServe())

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logrus.Infoln("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}

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
