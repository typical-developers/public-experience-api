package luau

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/typical-developers/goblox/opencloud"
	"github.com/typical-developers/goblox/pkg/methodutil"
)

var (
	scriptsPath = "./internal/luau/scripts"

	Opencloud = opencloud.NewClient().WithAPIKey(os.Getenv("OPENCLOUD_API_KEY"))
)

type Script string

const (
	OaklandsTranslationsScript Script = "oaklands/Translations.luau"
)

func Run[T any](ctx context.Context, universeId, placeId string, script Script) (*T, error) {
	p := filepath.Join(scriptsPath, string(script))

	file, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}

	task, _, err := Opencloud.LuauExecution.CreateLuauExecutionSessionTask(ctx, universeId, placeId, nil, opencloud.LuauExecutionTaskCreate{
		Script: opencloud.Pointer(string(file)),
	})
	if err != nil {
		return nil, err
	}

	var taskResult T
	var taskError error

	universeId, placeId, versionId, sessionId, taskId := task.TaskInfo()
	methodutil.PollMethod(func(done func()) {
		task, resp, err := Opencloud.LuauExecution.GetLuauExecutionSessionTask(ctx, universeId, placeId, versionId, sessionId, taskId)

		if err != nil {
			taskError = err
			done()
			return
		}

		// Keeps polling if the task is ratelimited.
		if resp.StatusCode == 429 {
			return
		}

		if task.State != opencloud.LuauExecutionStateProcessing && task.State != opencloud.LuauExecutionStateQueued {
			if task.Output.Results != nil {
				taskResult = task.Output.Results[0].(T)
			}

			if task.Error != nil {
				taskError = fmt.Errorf("LuauExecutionTask[%s]: %s", task.Error.Code, task.Error.Message)
			}

			done()
		}
	}, 0)

	return &taskResult, taskError
}
