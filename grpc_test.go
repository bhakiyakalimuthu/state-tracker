package state_tracker

import (
	"context"
	"testing"
	"time"

	"github.com/bhakiyakalimuthu/state-tracker/client/pb"
	proxyserver "github.com/bhakiyakalimuthu/state-tracker/server"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func TestGRPCServer(t *testing.T) {
	// Start the server in a separate goroutine
	go func() {
		if err := proxyserver.RunServer(context.Background(), func() {}, zap.L(), "9090", "grpc.osmosis.zone:9090"); err != nil {
			t.Fatalf("failed to run proxy server: %v", zap.Error(err))
		}
	}()
	<-time.After(time.Second * 2)
	actualRes := dialer(t, "localhost:9090")           // connect to local proxy
	expectedRes := dialer(t, "grpc.osmosis.zone:9090") // connect to actual server

	assert.Equal(t, actualRes.GetBlockId(), expectedRes.GetBlockId())
	assert.Equal(t, actualRes.GetBlock(), expectedRes.GetBlock())
	t.Logf("GetLatestBlockResponse from local server %+v", actualRes.GetBlock())
	t.Logf("GetLatestBlockResponse from osmosi server %+v", expectedRes.GetBlock())
}

func dialer(t *testing.T, serverAddress string) *pb.GetLatestBlockResponse {
	// Create a new gRPC client that connects to the server
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := pb.NewServiceClient(conn)

	// Make a gRPC call to the server
	res, err := client.GetLatestBlock(context.Background(), &pb.GetLatestBlockRequest{})
	if err != nil {
		t.Fatalf("failed to call MyMethod: %v", err)
	}
	return res
}
