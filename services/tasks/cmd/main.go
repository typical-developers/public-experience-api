package main

import (
	"runtime"
	"time"

	. "github.com/luckfire-go/cron-scheduler"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	_ "github.com/typical-developers/public-experience-api/internal/logger"
	"github.com/typical-developers/public-experience-api/services/tasks/jobs"
)

func main() {
	registry := NewRegistry(cron.WithLocation(time.UTC))

	registry.OnJobAddSuccess = func(job *RegistryItem) {
		log.WithFields(log.Fields{
			"Name": job.Name(),
			"Spec": job.Spec,
		}).Infof("Job has successfully been registered.")
	}
	registry.OnJobAddFailure = func(job *RegistryItem, err error) {
		log.WithError(err).Error("Failed to add job")
	}

	registry.AddJobs([]RegistryItem{
		{
			Enabled:  true,
			Spec:     "0 0 * * *",
			TaskFunc: jobs.CheckOaklandsUpdates,
		},
	})

	registry.Start()
	runtime.Goexit()
}
