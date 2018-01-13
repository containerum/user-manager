package main

import (
	"fmt"
	"os"

	"time"

	"context"
	"net/http"
	"os/signal"

	"git.containerum.net/ch/grpc-proto-files/auth"
	"git.containerum.net/ch/user-manager/clients"
	"git.containerum.net/ch/user-manager/routes"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func exitOnErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	viper.SetEnvPrefix("ch_user")
	viper.AutomaticEnv()
	exitOnErr(setupLogger())

	logrus.Infoln("starting server...")

	app := gin.New()
	app.Use(gin.RecoveryWithWriter(logrus.StandardLogger().WithField("component", "gin_recovery").WriterLevel(logrus.ErrorLevel)))
	app.Use(ginrus.Ginrus(logrus.StandardLogger(), time.RFC3339, true))

	db, err := getDB()
	exitOnErr(err)
	defer db.Close()

	mailClient := clients.NewMailClient(viper.GetString("mail_url"))

	reCaptchaClient := clients.NewReCaptchaClient(viper.GetString("recaptcha_key"))

	clients.RegisterOAuthClient(clients.NewGithubOAuthClient(viper.GetString("github_app_id"), viper.GetString("github_secret")))
	clients.RegisterOAuthClient(clients.NewGoogleOAuthClient(viper.GetString("google_app_id"), viper.GetString("google_secret")))
	clients.RegisterOAuthClient(clients.NewFacebookOAuthClient(viper.GetString("facebook_app_id"), viper.GetString("facebook_secret")))

	authConn, err := grpc.Dial(viper.GetString("auth_grpc_addr"), grpc.WithInsecure(), grpc.WithUnaryInterceptor(
		grpc_middleware.ChainUnaryClient(
			grpc_logrus.UnaryClientInterceptor(logrus.WithField("component", "auth_client")),
		)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             20 * time.Second,
			PermitWithoutStream: true,
		}))
	exitOnErr(err)
	defer authConn.Close()
	authClient := auth.NewAuthClient(authConn)

	webAPIClient := clients.NewWebAPIClient(viper.GetString("web_api_url"))

	routes.SetupRoutes(app, routes.Services{
		MailClient:      mailClient,
		DB:              db,
		AuthClient:      authClient,
		ReCaptchaClient: reCaptchaClient,
		WebAPIClient:    webAPIClient,
	})

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
