package types

import (
	"time"
)

type Job struct {
	ID          string    `json:"job_id"`
	CreateTime  time.Time `json:"create_time"`
	ResourceIds string    `json:"resource_ids"`
	Owner       string    `json:"owner"`
	ErrorCodes  string    `json:"error_codes"`
	Status      string    `json:"status"`
	StatusTime  time.Time `json:"status_time"`
}

type DescribeJobsResponse struct {
	ResponseStatus
	Total int    `json:"total_count"`
	Jobs  []*Job `json:"job_set"`
}
