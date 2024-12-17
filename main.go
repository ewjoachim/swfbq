package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/swf"
	"go.uber.org/zap"

	"github.com/ewjoachim/swfbq/bigquery"
	"github.com/ewjoachim/swfbq/cli"
	swfworker "github.com/ewjoachim/swfbq/swf"
)

func main() {
	// Parse CLI flags
	cfg, err := cli.ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	var logger *zap.Logger
	if cfg.Debug {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}
	defer logger.Sync()

	// Load AWS configuration
	awsCfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		logger.Fatal("Unable to load AWS SDK config", zap.Error(err))
	}

	// Initialize clients
	swfClient := swf.NewFromConfig(awsCfg)
	bqClient := bigquery.NewClient(logger)

	// Create worker
	worker := swfworker.NewWorker(
		swfClient,
		bqClient,
		cfg.Domain,
		cfg.TaskList,
		logger,
	)

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutting down...")
		cancel()
	}()

	// Start the worker
	if err := worker.Start(ctx); err != nil {
		logger.Fatal("Worker failed", zap.Error(err))
	}
}
