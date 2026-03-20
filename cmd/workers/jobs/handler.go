package jobs

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type RetryConfig struct {
	RetryAttempts int
	RetryDelay    time.Duration
}

type JobConfig struct {
	Enabled bool
	Retry   *RetryConfig
}

type Job struct {
	Spec   string
	Config JobConfig

	JobFunc func(context.Context)
}

func (j *Job) Run(ctx context.Context) func() {
	return func() {
		j.JobFunc(ctx)
	}
}

// Name will get the name of the job based on the provided function.
func (j *Job) Name() string {
	rf := runtime.FuncForPC(reflect.ValueOf(j.JobFunc).Pointer())
	if rf == nil {
		return fmt.Sprintf("Job %s", j.Spec)
	}

	name := rf.Name()
	if idx := strings.LastIndex(name, "."); idx >= 0 && idx < len(name)-1 {
		name = name[idx+1:]
	}

	return strings.TrimSuffix(name, "-fm")
}

func NewJob(spec string, cmd func(context.Context), config JobConfig) *Job {
	j := &Job{
		Spec:   spec,
		Config: config,

		JobFunc: cmd,
	}

	return j
}

type Handler struct {
	ctx context.Context
	c   *cron.Cron
	l   cron.Logger
}

type HandlerOpts struct {
	Cron   *cron.Cron
	Logger cron.Logger
}

func NewHandler(opts HandlerOpts, jobs ...Job) *Handler {
	h := &Handler{
		ctx: context.Background(),
		c:   opts.Cron,
		l:   opts.Logger,
	}

	if len(jobs) > 0 {
		h.AddJobs(jobs...)
	}

	return h
}

func (h *Handler) AddJobs(jobs ...Job) {
	for _, job := range jobs {
		if !job.Config.Enabled {
			continue
		}

		wrappers := []cron.JobWrapper{
			WebhookLogPanic(job),
		}
		if job.Config.Retry != nil {
			wrappers = append(wrappers, WithRetry(job, h.l))
		}

		chain := cron.NewChain(wrappers...).
			Then(cron.FuncJob(job.Run(h.ctx)))

		if _, err := h.c.AddFunc(job.Spec, chain.Run); err != nil {
			zap.L().Error("registry failed",
				zap.String("job", job.Name()),
				zap.String("spec", job.Spec),
				zap.Error(err),
			)
		}
	}
}
