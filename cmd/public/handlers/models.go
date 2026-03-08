package models

type Response[T any] struct {
	Data T `json:"data"`
}

type ErrorResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
