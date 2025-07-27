package luau

import (
	"context"
	"encoding/json"
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

type ScriptPath string

const (
	TestBinaryInputScriptPath ScriptPath = "oaklands/TestBinaryInput.luau"

	OaklandsUpdatesScriptPath ScriptPath = "oaklands/OaklandsUpdates.luau"
)

type RunOptions struct {
	BinaryInput        *string
	EnableBinaryOutput *bool
}

// Runs a luau script on a specific place.
// If you are using a binary input, you must create and provide it yourself.
func Run[R any, B any](ctx context.Context, universeId, placeId string, scriptPath ScriptPath, opts *RunOptions) (result *R, bResult *B, tError error) {
	p := filepath.Join(scriptsPath, string(scriptPath))

	file, err := os.ReadFile(p)
	if err != nil {
		return nil, nil, err
	}

	var runOpts RunOptions
	if opts != nil {
		runOpts = *opts
	}

	task, _, err := Opencloud.LuauExecution.CreateLuauExecutionSessionTask(ctx, universeId, placeId, nil, opencloud.LuauExecutionTaskCreate{
		Script:             opencloud.Pointer(string(file)),
		BinaryInput:        runOpts.BinaryInput,
		EnableBinaryOutput: runOpts.EnableBinaryOutput,
	})
	if err != nil {
		return nil, nil, err
	}

	universeId, placeId, versionId, sessionId, taskId := task.TaskInfo()
	methodutil.PollMethod(func(done func()) {
		task, resp, err := Opencloud.LuauExecution.GetLuauExecutionSessionTask(ctx, universeId, placeId, versionId, sessionId, taskId)
		if err != nil {
			tError = err
			done()
			return
		}

		if resp.StatusCode == 429 {
			return
		}

		if task.State != opencloud.LuauExecutionStateProcessing && task.State != opencloud.LuauExecutionStateQueued {
			if task.Output != nil && len(task.Output.Results) > 0 {
				result = new(R)

				if r, ok := task.Output.Results[0].(R); ok {
					result = &r
				}
			}

			if task.EnableBinaryOutput && task.BinaryOutputURI != "" {
				output, err := task.BinaryOutput(ctx)

				if err == nil {
					err = json.Unmarshal(output, &bResult)
					if err != nil {
						tError = err
					}
				} else {
					tError = err
				}
			}

			if task.Error != nil {
				tError = fmt.Errorf("LuauExecutionTask[%s]: %s", task.Error.Code, task.Error.Message)
			}

			done()
		}
	}, 0)

	return result, bResult, tError
}
