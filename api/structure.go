package api

type APIResponse[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

type ErrorAPIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type OaklandsTranslations APIResponse[map[string]string]

type OaklandsTranslationKeys APIResponse[[]string]

func Ptr[T any](v T) *T {
	return &v
}
