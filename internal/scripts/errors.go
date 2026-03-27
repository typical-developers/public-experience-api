package scripts

import "errors"

var (
	// This error happens when there is an error with one of the modules loading in the script executed.
	ErrModuleErrored = errors.New("Requested module experienced an error while loading")
)
