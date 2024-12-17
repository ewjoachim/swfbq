package swf

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/swf"
	"github.com/aws/aws-sdk-go-v2/service/swf/types"
	"github.com/ewjoachim/swfbq/bigquery"
	"github.com/ewjoachim/swfbq/models"
	"go.uber.org/zap"
)

type Worker struct {
	swfClient  *swf.Client
	bqClient   *bigquery.Client
	domain     string
	taskList   string
	logger     *zap.Logger
	maxWorkers int
	workerPool chan struct{}
	workerWg   sync.WaitGroup
}

func NewWorker(swfClient *swf.Client, bqClient *bigquery.Client, domain, taskList string, logger *zap.Logger) *Worker {
	const maxWorkers = 10 // Maximum number of concurrent queries
	return &Worker{
		swfClient:  swfClient,
		bqClient:   bqClient,
		domain:     domain,
		taskList:   taskList,
		logger:     logger,
		maxWorkers: maxWorkers,
		workerPool: make(chan struct{}, maxWorkers),
	}
}

func (w *Worker) Start(ctx context.Context) error {
	for i := 0; i < w.maxWorkers; i++ {
		w.workerPool <- struct{}{} // Initialize worker pool
	}

	for {
		select {
		case <-ctx.Done():
			w.workerWg.Wait() // Wait for all workers to finish
			return ctx.Err()
		case worker := <-w.workerPool:
			// Only poll when we have a worker available
			if err := w.pollAndProcessTask(ctx, worker); err != nil {
				w.logger.Error("Error processing task", zap.Error(err))
				w.workerPool <- worker // Return worker to pool on error
			}
		}
	}
}

func (w *Worker) pollAndProcessTask(ctx context.Context, worker struct{}) error {
	input := &swf.PollForActivityTaskInput{
		Domain:   aws.String(w.domain),
		TaskList: &types.TaskList{Name: aws.String(w.taskList)},
		Identity: aws.String("bigquery-worker"),
	}

	resp, err := w.swfClient.PollForActivityTask(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to poll for task: %v", err)
	}

	if resp.TaskToken == nil {
		w.workerPool <- worker // Return worker to pool if no task
		return nil             // No tasks available
	}

	var job models.Job
	if err := json.Unmarshal([]byte(*resp.Input), &job); err != nil {
		w.workerPool <- worker // Return worker to pool on error
		return w.completeTask(ctx, resp.TaskToken, nil, fmt.Errorf("invalid input: %v", err))
	}

	w.workerWg.Add(1)
	go func() {
		defer w.workerWg.Done()
		defer func() { w.workerPool <- worker }() // Return worker to pool

		w.logger.Info("Processing task",
			zap.String("task_token", *resp.TaskToken),
			zap.String("gcp_project", job.GCPProject))

		err := w.bqClient.ExecuteQuery(ctx, &job)
		if err := w.completeTask(ctx, resp.TaskToken, &job, err); err != nil {
			w.logger.Error("Failed to complete task",
				zap.String("task_token", *resp.TaskToken),
				zap.Error(err))
		}
	}()

	return nil
}

func (w *Worker) marshalJob(job *models.Job) (string, error) {
	if job == nil {
		return "", nil
	}
	// Create a copy to truncate SQL if needed
	resultJob := *job
	if len(resultJob.SQLQuery) > 1000 {
		resultJob.SQLQuery = resultJob.SQLQuery[:997] + "..."
	}
	result, err := json.Marshal(&resultJob)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job: %v", err)
	}
	return string(result), nil
}

func (w *Worker) completeTask(ctx context.Context, taskToken *string, job *models.Job, err error) error {
	if err != nil {
		failInput := &swf.RespondActivityTaskFailedInput{
			TaskToken: taskToken,
			Reason:    aws.String(err.Error()),
		}
		if details, err := w.marshalJob(job); err != nil {
			w.logger.Error("Failed to marshal job details", zap.Error(err))
		} else if details != "" {
			failInput.Details = aws.String(details)
		}
		_, err := w.swfClient.RespondActivityTaskFailed(ctx, failInput)
		return err
	}

	result, err := w.marshalJob(job)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %v", err)
	}

	completeInput := &swf.RespondActivityTaskCompletedInput{
		TaskToken: taskToken,
		Result:    aws.String(result),
	}
	_, err = w.swfClient.RespondActivityTaskCompleted(ctx, completeInput)
	return err
}
