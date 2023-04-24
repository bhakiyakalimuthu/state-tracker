package client

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/bhakiyakalimuthu/state-tracker/client/pb"
	"go.uber.org/zap"
)

const JobSize = 5

type Worker struct {
	logger *zap.Logger
	pChan  chan []Job // channel to receive jobs from producer
	cChan  chan []Job // channel to consume jobs and send it to writer
	client pb.ServiceClient
	file   *os.File
}

func NewWorker(logger *zap.Logger, pChan, cChan chan []Job, client pb.ServiceClient) (*Worker, error) {
	file, err := os.OpenFile(fmt.Sprintf("./block_data_%d.json", time.Now().Unix()), os.O_APPEND|os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		logger.Error("failed to create and open file", zap.Error(err))
		return nil, err
	}
	// Write the "test_result" key at the beginning of the file
	_, err = file.Write([]byte(`{"test_result":`))
	if err != nil {
		logger.Error("failed to write data to file", zap.Error(err))
		return nil, err
	}
	return &Worker{
		logger: logger,
		pChan:  pChan,
		cChan:  cChan,
		client: client,
		file:   file,
	}, nil
}

type Job struct {
	Hash   string `json:"hash"`
	Height int64  `json:"height"`
}

func (w *Worker) Process(wg *sync.WaitGroup, workerID int) error {
	defer wg.Done()
	for job := range w.cChan {
		w.logger.Debug("received job", zap.Any("job", job))
		encoder := json.NewEncoder(w.file)
		// Append the JSON data to the file
		if err := encoder.Encode(job); err != nil {
			w.logger.Error("failed to marshal jobs", zap.Error(err))
			continue
		}
	}
	err := w.file.Close()
	if err != nil {
		w.logger.Error("failed to close file", zap.Error(err))
	}
	w.logger.Warn("gracefully finishing worker process", zap.Int("WorkerID", workerID))
	return nil
}

// Consume used for gradual job flow
func (w *Worker) Consume(ctx context.Context) {
	for {
		select {
		case job := <-w.pChan: // fetch job from producer
			w.logger.Debug("received msg from consumerChan")
			w.cChan <- job // pass job to consumer
		case <-ctx.Done():
			close(w.cChan)
			return
		}
	}
}

func (w *Worker) Start(ctx context.Context) {
	// Initialize a variable to store the last processed block height
	var lastBlockHeight int64
	var jobs []Job
	// Cosmos block time is 1 second
	// To avoid querying the same block again
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			close(w.pChan)
			return
		case <-ticker.C:
			out, err := w.client.GetLatestBlock(ctx, &pb.GetLatestBlockRequest{})
			if err != nil {
				w.logger.Warn("failed to get latest block", zap.Error(err))
				continue
			}
			currentBlockHeight := out.GetBlock().GetHeader().GetHeight()
			if currentBlockHeight <= lastBlockHeight {
				w.logger.Info("same as current height, ", zap.Int64("currentBlockHeight", currentBlockHeight), zap.Int64("lastBlockHeight", lastBlockHeight)) // TODO:remove
				continue
			}
			lastBlockHeight = currentBlockHeight
			job := Job{
				Height: currentBlockHeight,
				Hash:   hex.EncodeToString(out.GetBlockId().GetHash()),
			}
			w.logger.Info("adding jobs", zap.Any("job", job)) // TODO:remove
			jobs = append(jobs, job)
			if len(jobs) == JobSize {
				w.logger.Info("job size reached, sending in to process", zap.Any("job", job)) // TODO:remove
				w.pChan <- jobs                                                               // send in jobs to process
				jobs = nil                                                                    // reset the job slice
			}
		}
	}
}
