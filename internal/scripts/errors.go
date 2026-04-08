package scripts

import "errors"

var (
	ErrNoScriptOutput                   = errors.New("no script output")
	ErrNoBinaryOutput                   = errors.New("no binary output")
	ErrScriptExecutionFailed            = errors.New("script execution failed")
	ErrScriptExecutionCancelled         = errors.New("script execution was cancelled")
	ErrScriptExecutionErrorUnknownState = errors.New("script execution returned an unknown state")

	// This error happens when there is an error with one of the modules loading in the script executed.
	ErrModuleErrored = errors.New("Requested module experienced an error while loading")
	// This error happens when a 401 or 403 is returned by the API.
	ErrNotAllowed = errors.New("The script failed to execute either due to a missing scope or the API key being invalid.")
)
