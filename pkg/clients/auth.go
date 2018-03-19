package clients

import (
	"io"

	auth "git.containerum.net/ch/auth/proto"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// AuthClientCloser is an extension of AuthClient interface with Closer interface to close connection
type AuthClientCloser interface {
	auth.AuthClient
	io.Closer
}

type grpcAuthClient struct {
	auth.AuthClient
	clientConn *grpc.ClientConn
}

// NewGRPCAuthClient returns client for auth service works using grpc protocol
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
