package GoBackgroundJobs

import "time"

type Queues map[string]func(jobData Job) error

type Config struct {
	PollingInterval   time.Duration
	RetryInterval     time.Duration
	MaxCompletedJob   int
	MaxConcurrentJobs int
	OnJobsSuccess     func(queueName string, job Job)
	OnJobsFailure     func(queueName string, job Job)
}

type Job struct {
	Id        string      `json:"id"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	MaxRetry  int         `json:"max_retry"`
	Error     interface{} `json:"error,omitempty"`
}
