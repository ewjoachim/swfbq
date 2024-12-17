package bigquery

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/ewjoachim/swfbq/models"
	"go.uber.org/zap"
)

type Client struct {
	logger *zap.Logger
}

func NewClient(logger *zap.Logger) *Client {
	return &Client{
		logger: logger,
	}
}

func (c *Client) ExecuteQuery(ctx context.Context, job *models.Job) error {
	client, err := bigquery.NewClient(ctx, job.GCPProject)
	if err != nil {
		return fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	q := client.Query(job.SQLQuery)

	// Start the query
	bqJob, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("query.Run: %v", err)
	}

	job.JobID = bqJob.ID()
	job.Status = models.JobStatusRunning
	job.StartTime = time.Now().UTC().Format(time.RFC3339)

	status, err := bqJob.Wait(ctx)
	if err != nil {
		job.Status = models.JobStatusFailed
		job.Error = err.Error()
		return fmt.Errorf("job.Wait: %v", err)
	}

	job.EndTime = time.Now().UTC().Format(time.RFC3339)

	if status.Err() != nil {
		job.Status = models.JobStatusFailed
		job.Error = status.Err().Error()
		return fmt.Errorf("job status error: %v", status.Err())
	}

	stats := status.Statistics
	job.Status = models.JobStatusCompleted
	job.BytesProcessed = stats.TotalBytesProcessed

	c.logger.Info("Query completed successfully",
		zap.String("job_id", job.JobID),
		zap.Int64("bytes_processed", job.BytesProcessed),
	)

	return nil
}
