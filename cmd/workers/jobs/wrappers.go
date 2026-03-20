package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/robfig/cron/v3"
	webhooks "github.com/typical-developers/discord-webhooks-go/v2"
	"github.com/typical-developers/public-experience-api/cmd/workers/config"
	"go.uber.org/zap"
)

var (
	PanicWebook *webhooks.WebhookClient
)

func init() {
	if config.C.LogWebhookURL != nil {
		PanicWebook = webhooks.NewWebhookClientFromURL(*config.C.LogWebhookURL)
	} else {
		zap.L().Warn("Webhook logging is not set up.")
	}
}

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

type logPanicPayload struct {
	Job   string `json:"job"`
	Error string `json:"error"`
	Time  string `json:"Time"`
}

func logPanic(job Job, panicValue any) {
	if PanicWebook == nil {
		return
	}

	ctx := context.Background()
	now := time.Now().UTC()

	panicInfo, err := json.MarshalIndent(logPanicPayload{
		Job:   job.Name(),
		Error: fmt.Sprintf("%v", panicValue),
		Time:  now.Format(time.RFC3339),
	}, "", " ")
	if err != nil {
		zap.L().Error("webhook log failed",
			zap.String("job", job.Name()),
			zap.Error(err),
		)
	}

	info := webhooks.WebhookFile{
		FileName: "panic.json",
		Reader:   bytes.NewReader(panicInfo),
	}

	stack := webhooks.WebhookFile{
		FileName: "stack.txt",
		Reader:   bytes.NewReader(debug.Stack()),
	}

	if _, _, err := PanicWebook.Execute(ctx, webhooks.MessagePayload{
		Content: fmt.Sprintf("`[%s][%s]` Something went wrong when executing a job. <@&1484636254904651927>", now.Format(time.TimeOnly), job.Name()),
		Files:   []webhooks.WebhookFile{info, stack},
		AllowedMentions: &webhooks.AllowedMentions{
			Roles: []string{"1484636254904651927"},
		},
	}, nil); err != nil {
		zap.L().Error("webhook log failed",
			zap.String("job", job.Name()),
			zap.Error(err),
		)
	}
}

// WebhookLogPanic will log the job's panic to a webhook.
func WebhookLogPanic(job Job) cron.JobWrapper {
	jobFunc := func(j cron.Job) {
		defer func() {
			if r := recover(); r != nil {
				logPanic(job, r)
				panic(r)
			}
		}()

		j.Run()
	}

	return func(j cron.Job) cron.Job {
		return cron.FuncJob(func() {
			jobFunc(j)
		})
	}
}
