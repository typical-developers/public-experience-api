package models

type Response[T any] struct {
	Data T `json:"data"`
}

type ErrorResponse struct {
	// The type of error that was returned.
	Type string `json:"type"`
	// The message for the returned error.
	Message string `json:"message"`
} //	@name	API.ErrorResponse
