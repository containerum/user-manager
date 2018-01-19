package main

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"runtime/debug"

	"context"
	"time"

	"git.containerum.net/ch/auth/routes"
	"git.containerum.net/ch/grpc-proto-files/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/husobee/vestigo"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type httpServer struct {
	listenAddr string
	server     *http.Server
}

// Server interface for grpc, http, etc. servers
type Server interface {
	// Run starts server. Call must be blocking
	Run() error

	// Stop stops server. Server should support graceful shutdown
	Stop() error
}

// NewHTTPServer returns server which servers REST requests
func NewHTTPServer(listenAddr string, tracer opentracing.Tracer, storage auth.AuthServer) Server {
	router := vestigo.NewRouter()
	routes.SetupRoutes(router, tracer, storage)
	server := &http.Server{
		Addr:    listenAddr,
		Handler: router,
	}
	return &httpServer{
		listenAddr: listenAddr,
		server:     server,
	}
}

func (s *httpServer) Run() error {
	logrus.WithField("listenAddr", s.listenAddr).Info("Starting HTTP server")
	return s.server.ListenAndServe()
}

func (s *httpServer) Stop() error {
	logrus.Info("Stopping HTTP server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

type grpcServer struct {
	listenAddr string
	server     *grpc.Server
}

func panicHandler(p interface{}) (err error) {
	logrus.Errorf("panic: %v", p)
	debug.PrintStack()
	return fmt.Errorf("panic: %v", p)
}

// NewGRPCServer reteurns server which servers request using grpc protocol
func NewGRPCServer(listenAddr string, tracer opentracing.Tracer, storage auth.AuthServer) Server {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			otgrpc.OpenTracingServerInterceptor(tracer, otgrpc.LogPayloads()),
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(panicHandler)),
			grpc_logrus.UnaryServerInterceptor(logrus.WithField("component", "grpc_server")),
		)),
	)
	auth.RegisterAuthServer(server, storage)
	return &grpcServer{
		listenAddr: listenAddr,
		server:     server,
	}
}

func (s *grpcServer) Run() error {
	logrus.WithField("listenAddr", s.listenAddr).Infof("Starting GRPC server")
	listener, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	return s.server.Serve(listener)
}

func (s *grpcServer) Stop() error {
	logrus.Infof("Stopping GRPC server")
	s.server.GracefulStop()
	return nil
}

// RunServers runs multiple servers in dedicated goroutines.
// Error on starting causes exit with code 1
func RunServers(servers ...Server) {
	for _, server := range servers {
		go func(s Server) {
			if err := s.Run(); err != nil {
				logrus.WithError(err).Error("Run server failed")
				os.Exit(1)
			}
		}(server)
	}
}

// StopServers stops servers.
// It should be triggered after receiving interrupt signal from OS.
func StopServers(servers ...Server) {
	for _, server := range servers {
		if err := server.Stop(); err != nil {
			logrus.WithError(err).Error("Error at stopping server")
		}
	}
}
