package client

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bhakiyakalimuthu/state-tracker/client/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const workerPoolSize = 5

func RunClient(ctx context.Context, cancel context.CancelFunc, logger *zap.Logger, serverAddrs string) {
	conn, err := grpc.DialContext(ctx, serverAddrs, grpc.WithInsecure())
	if err != nil {
		logger.Fatal("failed to dial to proxy server", zap.Error(err))
	}
	client := pb.NewServiceClient(conn)

	// producer channel
	pChan := make(chan []Job, 1)
	// consumer channel
	cChan := make(chan []Job, workerPoolSize)

	// create worker
	w, err := NewWorker(logger, pChan, cChan, client)
	if err != nil {
		logger.Fatal("failed to create worker", zap.Error(err))
	}

	//  setup wait group with cancellation support
	wg := new(sync.WaitGroup)

	// start  the worker and consumer
	go w.Start(ctx)
	go w.Consume(ctx)

	// start worker process and add worker pool
	wg.Add(workerPoolSize)
	for i := 1; i <= workerPoolSize; i++ {
		go w.Process(wg, i)
	}
	// handle shut down
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-shutdown:
		logger.Warn("Received shutdown signal & server shutting down")
		cancel()
	case <-ctx.Done():
		logger.Warn("Context cancelled")
	}
	signal.Stop(shutdown)

	cancel() // cancel context
	// even if cancellation received, current running job will be not be interrupted until it completes
	wg.Wait() // wait for the workers to be completed
	logger.Warn("All jobs are done, shutting down")
}
