package GoBackgroundJobs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mia-burton/go-background-jobs/connectors"
	"golang.org/x/sync/semaphore"
	"log"
	"runtime"
	"time"
)

type JobsHandler struct {
	connector            connectors.Connector
	selectJobFunctionMap Queues
	config               Config
}

func (j *JobsHandler) Init(queues Queues, conf *Config) {
	j.setConfig(conf)
	j.selectJobFunctionMap = queues
	for job, _ := range j.selectJobFunctionMap {
		go j.startExecution(job)
	}
}

func (j *JobsHandler) AddJob(jobName string, data interface{}, maxRetry int) error {
	found := false
	for job, _ := range j.selectJobFunctionMap {
		if jobName == job {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("%s job function not found\n", jobName)
	}

	err := j.addToQueue(jobName, data, maxRetry)
	if err != nil {
		return err
	}
	return nil
}

func (j *JobsHandler) addToQueue(jobName string, data interface{}, maxRetry int) error {
	queueToWork := jobName + ":towork"
	id, err := j.connector.NewProgressiveId(queueToWork)
	if err != nil {
		return err
	}
	job := newJob(id, data, maxRetry)
	b, err := json.Marshal(job)
	if err != nil {
		return err
	}
	err = j.connector.Add(queueToWork, id, b)
	if err != nil {
		return err
	}
	return nil
}

func (j *JobsHandler) startExecution(jobName string) {

	queueToWork := jobName + ":towork"
	queueInProgress := jobName + ":inprogress"
	queueCompleted := jobName + ":completed"
	queueFailed := jobName + ":failed"

	sem := semaphore.NewWeighted(int64(j.config.MaxConcurrentJobs))

	for {
		time.Sleep(j.config.PollingInterval)

		// Check if there are elements in the to-work queue
		size, err := j.connector.GetQueueLength(queueToWork)
		if err != nil {
			log.Fatal(err)
		}

		if size == 0 {
			continue // Nothing to do, go back to sleep
		}

		// Get the first element from the to-work queue
		value, err := j.connector.RemoveFirstElement(queueToWork)
		if err != nil {
			log.Println(err.Error())
			continue // Something went wrong, go back to sleep
		}

		var job Job
		err = json.Unmarshal(value, &job)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		// Put job in in-progress queue
		_ = j.connector.Add(queueInProgress, job.Id, value)

		err = sem.Acquire(context.Background(), 1)
		if err != nil {
			log.Fatal(err.Error())
		}
		// Execute a new thread to process the job
		go func() {
			defer func() { sem.Release(1) }()
			jobInThread := job
			jobNameInThread := jobName
			retry := 0
			for {
				jobFunc, _ := j.selectJobFunction(jobNameInThread) // Select the function to process the job based on queue name
				err = jobFunc(jobInThread)                         // Call the function with the job data
				if err == nil {
					// Function succeeded, move element to completed queue
					_ = j.connector.Add(queueCompleted, jobInThread.Id, value)
					size, _ = j.connector.GetQueueLength(queueCompleted)
					if size > int64(j.config.MaxCompletedJob) {
						// Remove oldest element if queue exceeds 100 items
						_ = j.connector.Trim(queueCompleted, 0, int64(j.config.MaxCompletedJob))
					}
					// Remove element from in-progress queue
					_ = j.connector.RemoveById(queueInProgress, jobInThread.Id)
					retry = 0
					j.config.OnJobsSuccess(jobNameInThread, jobInThread)
					break
				} else if retry >= jobInThread.MaxRetry {
					// Function failed after maxRetry tries, move element to failed queue
					jobInThread.Error = err.Error()
					jobWithErr, _ := json.Marshal(jobInThread)
					_ = j.connector.Add(queueFailed, jobInThread.Id, jobWithErr)
					// Remove element from in-progress queue
					_, _ = j.connector.RemoveFirstElement(queueInProgress)
					retry = 0
					j.config.OnJobsFailure(jobNameInThread, jobInThread)
					break
				} else {
					// Function failed, try again
					time.Sleep(j.config.RetryInterval)
					retry++
				}
			}
		}()
	}
}

func (j *JobsHandler) selectJobFunction(queueName string) (func(jobData Job) error, error) {
	job := j.selectJobFunctionMap[queueName]
	if job == nil {
		return nil, fmt.Errorf("%s job function not found", queueName)
	}
	return job, nil
}

func (j *JobsHandler) setConfig(conf *Config) {
	j.config = Config{
		RetryInterval:     10 * time.Second,
		PollingInterval:   2 * time.Second,
		MaxCompletedJob:   100,
		MaxConcurrentJobs: runtime.NumCPU(),
		OnJobsSuccess:     func(queueName string, job Job) {},
		OnJobsFailure:     func(queueName string, job Job) {},
	}

	if conf != nil {
		j.config = *conf
	}

	if conf.OnJobsSuccess == nil {
		j.config.OnJobsSuccess = func(queueName string, job Job) {}
	}

	if conf.OnJobsFailure == nil {
		j.config.OnJobsFailure = func(queueName string, job Job) {}
	}
}

func newJob(id string, data interface{}, maxRetry int) Job {
	return Job{Id: id, Data: data, Timestamp: time.Now(), MaxRetry: maxRetry}
}

func NewRedisJobsHandler(host string, port string, pwd string) JobsHandler {
	client := connectors.RedisConnection(host, port, pwd)
	redisConnector := connectors.NewRedisConnector(context.Background(), client)
	return JobsHandler{connector: redisConnector}
}
