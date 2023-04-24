package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/bhakiyakalimuthu/state-tracker/server/pb"
	"github.com/siderolabs/grpc-proxy/proxy"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type server struct {
	pb.UnimplementedProxyServiceServer
}

// RunServer run server
func RunServer(ctx context.Context, cancel context.CancelFunc, logger *zap.Logger, port, proxyTo string) error {
	director := func(ctx context.Context, fullMethodName string) (proxy.Mode, []proxy.Backend, error) {
		// Backend where all the inbound request will be forwarded
		backend := &proxy.SingleBackend{
			GetConn: func(ctx context.Context) (context.Context, *grpc.ClientConn, error) {
				md, _ := metadata.FromIncomingContext(ctx)
				logger.Info("backend meta data", zap.Any("metaData", md), zap.String("methodName", fullMethodName))
				// Copy the inbound metadata explicitly.
				outCtx := metadata.NewOutgoingContext(ctx, md.Copy())
				// Make sure we use DialContext so that dialing can be cancelled/time out together with the context.
				conn, err := grpc.DialContext(ctx, proxyTo, grpc.WithCodec(proxy.Codec()), grpc.WithTransportCredentials(insecure.NewCredentials())) // nolint: staticcheck
				return outCtx, conn, err
			},
		}
		// Decide on which backend to dial
		// one2one proxying should have exactly one connection
		return proxy.One2One, []proxy.Backend{backend}, nil
	}

	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	// init grpc proxy server
	s := grpc.NewServer(grpc.CustomCodec(proxy.Codec()), grpc.UnknownServiceHandler(proxy.TransparentHandler(director))) // nolint:staticcheck
	pb.RegisterProxyServiceServer(s, &server{})

	go func() {
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-shutdown:
			logger.Warn("Received shutdown signal & server shutting down")
			cancel()
		case <-ctx.Done():
			logger.Warn("Context cancelled")
		}
		s.Stop() // stop the grpc proxy server
		signal.Stop(shutdown)
	}()
	logger.Info("Starting grpc proxy", zap.String("port", port))
	return s.Serve(listen)
}
