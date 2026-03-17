package jobs

import (
	"context"
	"fmt"
	"reflect"
	"runtime"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type JobConfig struct {
	Enabled bool
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

// name will get the name of the job based on the provided function.
func (j *Job) name() string {
	rf := runtime.FuncForPC(reflect.ValueOf(j.JobFunc).Pointer())
	if rf == nil {
		return fmt.Sprintf("Job %s", j.Spec)
	}

	return rf.Name()
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
	c   *cron.Cron
	ctx context.Context
}

type HandlerOpts struct {
	Cron *cron.Cron
}

func NewHandler(opts HandlerOpts, jobs ...Job) *Handler {
	h := &Handler{
		c:   opts.Cron,
		ctx: context.Background(),
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

		if _, err := h.c.AddFunc(job.Spec, job.Run(h.ctx)); err != nil {
			zap.L().Error("registry failed",
				zap.String("job", job.name()),
				zap.String("spec", job.Spec),
				zap.Error(err),
			)
		}
	}
}

func (h *Handler) Start() {
	h.c.Start()
	runtime.Goexit()
}
