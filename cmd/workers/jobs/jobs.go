package jobs

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

type RetryConfig struct {
	RetryAttempts int
	RetryDelay    time.Duration
	Logger        cron.Logger
}

// WithRetry will retry the job on failure automatically.
func WithRetry(c RetryConfig) cron.JobWrapper {
	if c.RetryAttempts <= 0 {
		c.RetryAttempts = 1
	}

	job := func(j cron.Job) {
		for attempt := 1; attempt <= c.RetryAttempts; attempt++ {
			failed := false

			var panicResult any
			func() {
				defer func() {
					if r := recover(); r != nil {
						failed = true
						panicResult = r
					}
				}()

				j.Run()
			}()

			if !failed {
				return
			}

			if attempt == c.RetryAttempts {
				panic(panicResult)
			}

			c.Logger.Error(
				fmt.Errorf("panic: %v", panicResult),
				fmt.Sprintf("job failed, retrying in %s", c.RetryDelay),
			)

			time.Sleep(c.RetryDelay)
		}
	}

	return func(j cron.Job) cron.Job {
		return cron.FuncJob(func() {
			job(j)
		})
	}
}
