# Go Background Jobs
Go Background Jobs is a Go package that allows you to execute asynchronous jobs utilizing Redis queues.
This package currently supports Redis as the backend for job queuing, but there are plans to add support for similar tools in the future.

## Installation
To install the package, run the following command:

    go get github.com/mia-burton/go-background-jobs

## Initialization
To use the package, you need to initialize it by calling the `NewRedisJobsHandler` function to create a `JobsHandler` object. This function requires the Redis connection parameters (host, port, and password) to be passed as arguments.

    func NewRedisJobsHandler(host string, port string, pwd string) JobsHandler

The `NewRedisJobsHandler` function returns a `JobsHandler` object. To complete the initialization, call the `Init` function on the `JobsHandler` object. This function requires two arguments: `queues` and `conf`.

    func (j *JobsHandler) Init(queues Queues, conf *Config)

The `queues` argument is a map that associates a job name (string) with a job execution function. The `conf` argument is a pointer to a Config struct, which contains the configuration settings for the package.

## Adding Jobs
To add a job to be executed at runtime, use the `AddJob` function on the `JobsHandler` object.
    
    func (j *JobsHandler) AddJob(jobName string, data interface{}, maxRetry int) error

The `AddJob` function requires the job name, job data (the data to process inside the job), and the maximum number of retries for the job as arguments.

## Queues
The `Queues` type is a map that associates a job name (string) with a job execution function. The job execution function should have the following signature:
    
    func(jobData Job) error

## Configuration
The `Config` struct contains the following configuration settings for the package:

    type Config struct {
        PollingInterval   time.Duration
        RetryInterval     time.Duration
        MaxCompletedJob   int
        MaxConcurrentJobs int
        OnJobsSuccess     func()
        OnJobsFailure     func()
    }

`PollingInterval:` The interval at which the package polls the job queues for new jobs.

`RetryInterval:` The interval between retries for failed jobs.

`MaxCompletedJob:` The maximum number of completed jobs to keep track of.

`MaxConcurrentJobs:` The maximum number of concurrent jobs to be executed.

`OnJobsSuccess:` A callback function to be executed when all jobs are successfully completed.

`OnJobsFailure:` A callback function to be executed when any job fails.


## Examples
Here is an example of how to use the package:

    package main

    import (
    "time"
    
        "github.com/mia-burton/go-background-jobs"
    )
    
    func main() {
        queues := jobs.Queues{
        "job1": job1,
        "job2": job2,
        }
    
        conf := &jobs.Config{
            PollingInterval:   1 * time.Second,
            RetryInterval:     5 * time.Minute,
            MaxCompletedJob:   1000,
            MaxConcurrentJobs: 10,
            OnJobsSuccess:     onJobsSuccess,
            OnJobsFailure:     onJobsFailure,
        }
    
        handler := jobs.NewRedisJobsHandler("localhost", "6379", "password")
        handler.Init(queues, conf)
    
        // Add jobs at runtime
        jobData := map[string]string{"name": "John"} // You can use any kind of object
        err := handler.AddJob("job1", jobData, 3)
        if err != nil {
            // Handle error
        }

    }
    
    func job1(job jobs.Job) error {
        // Execute job logic for job1
        // The data to be process is inside job.Data
        return nil
    }
    
    func job2(job jobs.Job) error {
        // Execute job logic for job2
        return nil
    }
    
    func onJobsSuccess() {
        // Handle successful completion of all jobs
    }
    
    func onJobsFailure() {
        // Handle failure of any job
    }

This example initializes the package with Redis as the job queue backend and adds two job execution functions (`job1` and `job2`) to the `Queues` map. The package is then configured with the desired settings, and the jobs handler is initialized with the specified queues and configuration. Finally, a job is added at runtime using the `AddJob` function.

## Future Development
Future updates to this package may include support for other tools similar to Redis for job queuing and the ability to schedule recurring jobs with specified intervals.

## Contributing
If you would like to contribute to this package, feel free to submit a pull request on this repository.