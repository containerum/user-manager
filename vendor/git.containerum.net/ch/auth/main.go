package main

import (
	"fmt"

	"os"

	"os/signal"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func logExit(err error) {
	if err != nil {
		logrus.WithError(err).Fatalf("Setup error")
		os.Exit(1)
	}
}

func main() {
	viper.SetEnvPrefix("ch_auth")
	viper.AutomaticEnv()

	if err := logLevelSetup(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := logModeSetup(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	viper.SetDefault("http_listenaddr", ":8080")
	httpTracer, err := getTracer(viper.GetString("http_listenaddr"), "ch-auth-rest")
	logExit(err)

	viper.SetDefault("grpc_listenaddr", ":8888")
	grpcTracer, err := getTracer(viper.GetString("grpc_listenaddr"), "ch-auth-grpc")
	logExit(err)

	storage, err := getStorage()
	logExit(err)

	servers := []Server{
		NewHTTPServer(viper.GetString("http_listenaddr"), httpTracer, storage),
		NewGRPCServer(viper.GetString("grpc_listenaddr"), grpcTracer, storage),
	}

	RunServers(servers...)

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	StopServers(servers...)
}
