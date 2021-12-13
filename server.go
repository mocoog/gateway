package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/mocoog/gateway/pkg/logger"
	"github.com/mocoog/gateway/pkg/runner"
	pb "github.com/mocoog/grpc-go-packages/gateway/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// ServerUnary is server
type ServerUnary struct {
	pb.UnimplementedGatewayServiceServer
}

const port = ":9881"

func main() {
	runner.Run(server)
}

func server(ctx context.Context) int {
	l, err := logger.New()
	if err != nil {
		_, ferr := fmt.Fprintf(os.Stderr, "failed to create logger: %s", err)
		if ferr != nil {
			// Unhandleable, something went wrong...
			panic(fmt.Sprintf("failed to write log:`%s` original error is:`%s`", ferr, err))
		}
		return 1
	}
	gatewayLogger := l.WithName("gateway")

	lis, err := net.Listen("tcp", port)
	if err != nil {
		panic(errors.Wrap(err, "ポート失敗"))
	}

	s := grpc.NewServer()
	var server ServerUnary
	pb.RegisterGatewayServiceServer(s, &server)

	errCh := make(chan error, 1)
	go func() {
		if err := s.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	// if err := s.Serve(lis); err != nil {
	//     return errors.Wrap(err, "サーバ起動失敗")
	// }
	// return nil

	select {
	case err := <-errCh:
		gatewayLogger.Error(err, "failed to serve http server")
		return 1
	case <-ctx.Done():
		gatewayLogger.Info("shutting down...")
		return 0
	}
}

func (s *ServerUnary) HealthCheck(ctx context.Context, in *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	fmt.Println(in)
	return &pb.HealthCheckResponse{
		Status: in.Status,
	}, nil
}
