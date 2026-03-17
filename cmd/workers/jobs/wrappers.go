package jobs

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// WithRetry will retry the job on failure automatically.
func WithRetry(job Job, l cron.Logger) cron.JobWrapper {
	if job.Config.Retry.RetryAttempts <= 0 {
		job.Config.Retry.RetryAttempts = 1
	}
	if l == nil {
		l = cron.DefaultLogger
	}

	jobFunc := func(j cron.Job) {
		for attempt := 1; attempt <= job.Config.Retry.RetryAttempts; attempt++ {
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

			if attempt == job.Config.Retry.RetryAttempts {
				panic(panicResult)
			}

			l.Error(
				fmt.Errorf("%s panic: %v", job.Name(), panicResult),
				"failed",
				zap.Duration("retry_in", job.Config.Retry.RetryDelay),
				zap.Int("current_attempt", attempt),
				zap.Int("remaining_attempts", job.Config.Retry.RetryAttempts-attempt),
			)

			time.Sleep(job.Config.Retry.RetryDelay)
		}
	}

	return func(j cron.Job) cron.Job {
		return cron.FuncJob(func() {
			jobFunc(j)
		})
	}
}
