package models

type JobStatus string

const (
	JobStatusPending   JobStatus = "PENDING"
	JobStatusRunning   JobStatus = "RUNNING"
	JobStatusCompleted JobStatus = "COMPLETED"
	JobStatusFailed    JobStatus = "FAILED"
)

// Job represents the structure of our SWF job
type Job struct {
	GCPProject     string    `json:"gcp_project"`
	SQLQuery       string    `json:"sql_query"`
	JobID          string    `json:"job_id,omitempty"`
	Status         JobStatus `json:"status,omitempty"`
	Error          string    `json:"error,omitempty"`
	StartTime      string    `json:"start_time,omitempty"`
	EndTime        string    `json:"end_time,omitempty"`
	RowsCount      int64     `json:"rows_count,omitempty"`
	BytesProcessed int64     `json:"bytes_processed,omitempty"`
}
