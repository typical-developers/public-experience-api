package main

import (
	"runtime"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/typical-developers/public-experience-api/services/tasks/jobs"
)

type JobRegistryEntry struct {
	Name             string
	RunOnceAtStartup bool
	Disabled         bool
	Interval         string
	Task             func()
}

var JobRegistry = []JobRegistryEntry{
	{
		Name:             "Oaklands Updates",
		RunOnceAtStartup: true,
		Disabled:         false,
		Interval:         "0 0 * * *",
		Task:             jobs.CheckOaklandsUpdates,
	},
}

var Cron = cron.New(cron.WithLocation(time.UTC))

// func functionName(i interface{}) string {
// 	funcName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
// 	pieces := strings.Split(funcName, "/")

// 	return pieces[len(pieces)-1]
// }

func main() {
	Cron.Start()

	for _, job := range JobRegistry {
		// jobName := functionName(job.Task)

		if job.Disabled {
			continue
		}

		_, err := Cron.AddFunc(job.Interval, job.Task)
		if err != nil {
			println(err.Error())
		}

		if job.RunOnceAtStartup {
			go job.Task()
		}
	}

	runtime.Goexit()
}
