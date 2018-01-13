package clients

import (
	"io"

	"git.containerum.net/ch/grpc-proto-files/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type AuthClientCloser interface {
	auth.AuthClient
	io.Closer
}

type grpcAuthClient struct {
	auth.AuthClient
	clientConn *grpc.ClientConn
}

func NewGRPCAuthClient(serverAddr string) (AuthClientCloser, error) {
	authConn, err := grpc.Dial(serverAddr, grpc.WithInsecure(), grpc.WithUnaryInterceptor(
		grpc_middleware.ChainUnaryClient(
			grpc_logrus.UnaryClientInterceptor(logrus.WithField("component", "auth_client")),
		)),
	)
	if err != nil {
		return nil, err
	}
	return &grpcAuthClient{
		AuthClient: auth.NewAuthClient(authConn),
		clientConn: authConn,
	}, nil
}

func (g *grpcAuthClient) Close() error {
	return g.clientConn.Close()
}
