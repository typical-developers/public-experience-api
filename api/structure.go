package api

type APIResponse[T any] struct {
	Success bool    `json:"success"`
	Message *string `json:"message,omitempty"`
	Data    *T      `json:"data,omitempty"`
}

func Ptr[T any](v T) *T {
	return &v
}
