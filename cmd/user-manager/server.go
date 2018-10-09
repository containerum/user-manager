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
	"github.com/containerum/kube-client/pkg/model"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func initServer(c *cli.Context) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.TabIndent|tabwriter.Debug)
	for _, f := range c.GlobalFlagNames() {
		fmt.Fprintf(w, "Flag: %s\t Value: %s\n", f, c.String(f))
	}
	w.Flush()

	setupLogs(c)

	err := oauthClientsSetup(c)
	exitOnErr(err)

	tgClient, err := getTelegramClient(c)
	exitOnErr(err)

	userManager, err := getUserManager(c, server.Services{
		MailClient:        getService(getMailClient(c)).(clients.MailClient),
		DB:                getService(getDB(c)).(db.DB),
		AuthClient:        getService(getAuthClient(c)).(clients.AuthClient),
		ReCaptchaClient:   getService(getReCaptchaClient(c)).(clients.ReCaptchaClient),
		PermissionsClient: getService(getPermissionsClient(c)).(clients.PermissionsClient),
		EventsClient:      getService(getEventsClient(c)).(clients.EventsClient),
		TelegramClient:    tgClient,
	})
	exitOnErr(err)
	defer userManager.Close()

	status := model.ServiceStatus{
		Name:     c.App.Name,
		Version:  c.App.Version,
		StatusOK: true,
	}

	app := router.CreateRouter(&userManager, &status, c.Bool(corsFlag))

	if c.String(adminPwdFlag) != "" {
		err := userManager.CreateFirstAdmin(c.String(adminPwdFlag))
		exitOnErr(err)
	}

	// graceful shutdown support
	srv := http.Server{
		Addr:    ":" + c.String(portFlag),
		Handler: app,
	}

	go exitOnErr(srv.ListenAndServe())

	quit := make(chan os.Signal, 1)
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
