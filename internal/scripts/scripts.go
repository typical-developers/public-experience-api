package scripts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/typical-developers/goblox/opencloud"
)

var (
	ErrNoScriptOutput = errors.New("no script output")
	ErrNoBinaryOutput = errors.New("no binary output")

	ErrScriptExecutionFailed            = errors.New("script execution failed")
	ErrScriptExecutionCancelled         = errors.New("script execution was cancelled")
	ErrScriptExecutionErrorUnknownState = errors.New("script execution returned an unknown state")
)

type Script struct {
	content string
	client  *opencloud.Client
}

// NewScript will create a new Script based on direct script content.
func NewScript(client *opencloud.Client, content string) *Script {
	return &Script{
		content: content,
		client:  client,
	}
}

// NewScriptFromFile will create a new Script based on a file path.
func NewScriptFromFile(client *opencloud.Client, path string) (*Script, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	script := NewScript(client, string(content))

	return script, nil
}

type pollResults struct {
	task *opencloud.LuauExecutionTask
	err  error
}

// poll will poll the provided task until it isnt processing or queued and return its final state.
func (s *Script) poll(ctx context.Context, task *opencloud.LuauExecutionTask) (*opencloud.LuauExecutionTask, error) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	universeID, placeID, versionId, sessionId, taskId := task.TaskInfo()
	for {
		select {
		case <-ctx.Done():
			return task, ctx.Err()
		case <-ticker.C:
			task, resp, err := s.client.LuauExecution.GetLuauExecutionSessionTask(ctx, universeID, placeID, versionId, sessionId, taskId)
			fmt.Printf("%+v\n", task)

			if err != nil {
				return task, err
			}

			if resp.StatusCode == http.StatusTooManyRequests {
				// TODO: log warning here for ratelimit
				continue
			}

			switch task.State {
			case opencloud.LuauExecutionStateProcessing, opencloud.LuauExecutionStateQueued:
				continue
			case opencloud.LuauExecutionStateUnspecified:
				return task, ErrScriptExecutionFailed
			case opencloud.LuauExecutionStateCancelled:
				return task, ErrScriptExecutionCancelled
			case opencloud.LuauExecutionStateFailed:
				if task.Error != nil {
					return task, errors.New(task.Error.Message)
				}
				return task, ErrScriptExecutionFailed
			case opencloud.LuauExecutionStateComplete:
				return task, nil
			default:
				return task, ErrScriptExecutionErrorUnknownState
			}
		}
	}
}

type ExecuteOptions struct {
	UniverseID string
	PlaceID    string
	Version    *string

	EnableBinaryOutput *bool
}

type ExecuteResult struct {
	results      []any
	binaryOutput *[]byte
}

// DecodeResult will decode the execution output result into the provided value parameter.
func (e *ExecuteResult) DecodeResult(v any) error {
	if len(e.results) <= 0 || e.results[0] == nil {
		return ErrNoScriptOutput
	}

	result := e.results[0]
	if wrapped, ok := result.(map[string]any); ok {
		if returnValues, exists := wrapped["ReturnValues"]; exists {
			result = returnValues
		}
	}

	jsonb, err := json.Marshal(result)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(jsonb)
	if err := json.NewDecoder(reader).Decode(v); err != nil {
		return err
	}

	return nil
}

// DecodeBinaryOutput will decode the binary output result into the provided value parameter.
func (e *ExecuteResult) DecodeBinaryOutput(v any) error {
	if e.binaryOutput == nil {
		return ErrNoBinaryOutput
	}

	reader := bytes.NewReader(*e.binaryOutput)
	if err := json.NewDecoder(reader).Decode(v); err != nil {
		return err
	}

	return nil
}

// Execute will run the script and return its result.
func (s *Script) Execute(ctx context.Context, opts ExecuteOptions) (*ExecuteResult, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute*5) // luau execution has a 5 minute return window.
	defer cancel()

	task, _, err := s.client.LuauExecution.CreateLuauExecutionSessionTask(
		ctx,
		opts.UniverseID, opts.PlaceID, opts.Version,
		opencloud.LuauExecutionTaskCreate{
			Script:             &s.content,
			EnableBinaryOutput: opts.EnableBinaryOutput,
		},
	)
	if err != nil {
		return nil, err
	}

	pollFinished := make(chan pollResults, 1)
	go func() {
		task, err := s.poll(ctx, task)
		pollFinished <- pollResults{
			task: task,
			err:  err,
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-pollFinished:
		if res.err != nil {
			return nil, res.err
		}

		execResult := &ExecuteResult{
			results:      nil,
			binaryOutput: nil,
		}

		if res.task.Output != nil && len(res.task.Output.Results) != 0 {
			execResult.results = res.task.Output.Results
		}

		if res.task.BinaryOutputURI != "" {
			binaryOutput, err := res.task.BinaryOutput(ctx)
			if err != nil {
				return nil, err
			}
			execResult.binaryOutput = &binaryOutput
		}

		return execResult, nil
	}
}
