# SWFBQ (SWF BigQuery Worker)

A Go worker that executes BigQuery queries through AWS Simple Workflow (SWF) tasks.

## Overview

This worker:
1. Polls an AWS SWF task list for jobs
2. Executes BigQuery queries based on the job specifications
3. Reports results back to SWF

## Prerequisites

- AWS credentials configured
- GCP credentials configured
- Access to AWS SWF
- Access to GCP BigQuery

## Installation

### Using Docker

The worker is available as a Docker image from GitHub Container Registry:

```bash
# Pull the latest version
docker pull ghcr.io/ewjoachim/swfbq:latest

# Or a specific version
docker pull ghcr.io/ewjoachim/swfbq:v1.0.0

# Run with required configuration
docker run \
    -e AWS_ACCESS_KEY_ID=your_key \
    -e AWS_SECRET_ACCESS_KEY=your_secret \
    -e GOOGLE_APPLICATION_CREDENTIALS=/app/credentials.json \
    -v /path/to/your/credentials.json:/app/credentials.json:ro \
    ghcr.io/ewjoachim/swfbq:latest \
    -domain your-domain \
    -task-list your-tasklist
```

### From Source

```bash
git clone github.com/ewjoachim/swfbq
cd swfbq
go mod download
```

## Configuration

The application uses environment variables and command-line flags for configuration:

### Environment Variables

```bash
# AWS Authentication (if not using IAM roles)
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"

# GCP Authentication
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/credentials.json"
```

### Command-line Flags

```bash
swfbq -domain your-domain -task-list your-tasklist [-debug]
```

### AWS Authentication

The worker uses the AWS SDK's default credential chain. You can authenticate by:
- AWS CLI configuration (`aws configure`)
- Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
- IAM roles (when running on AWS infrastructure)

### GCP Authentication

The worker uses Google Cloud's Application Default Credentials. You can authenticate by:
- Setting `GOOGLE_APPLICATION_CREDENTIALS` environment variable
- Using GCP workload identity
- Running on GCP infrastructure with appropriate service account

## Job Format

Jobs should be submitted to SWF in the following JSON format:

```json
{
    "gcp_project": "your-gcp-project-id",
    "sql_query": "SELECT * FROM `project.dataset.table` LIMIT 10"
}
```

### Response Format

The worker will complete the task with a result in the following format:

```json
{
    "gcp_project": "your-gcp-project-id",
    "sql_query": "SELECT * FROM `project.dataset.table` LIMIT 10",
    "job_id": "bigquery-job-id",
    "status": "COMPLETED",
    "start_time": "2024-01-01T00:00:00Z",
    "end_time": "2024-01-01T00:00:01Z",
    "rows_count": 10,
    "bytes_processed": 1024
}
```

## Project Structure

```
.
├── README.md
├── main.go                 # Application entry point
├── go.mod                 # Go module definition
├── config/
├── bigquery/
│   └── client.go         # BigQuery operations
├── swf/
│   └── worker.go         # SWF worker implementation
└── models/
    └── job.go            # Shared data structures
```

## Monitoring

The application uses structured logging via `zap` logger, logging:
- Task processing events
- Query execution details
- Error conditions
- Performance metrics

## Error Handling

The worker handles several types of errors:
- SWF polling errors
- Invalid job specifications
- BigQuery execution errors
- Authentication/authorization errors

All errors are:
1. Logged with appropriate context
2. Reported back to SWF
3. Handled gracefully to ensure worker continuity

## License

This project is licensed under the MIT License - see the LICENSE file for details.
