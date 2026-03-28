package rest

import (
	"errors"
	"net/http"

	"github.com/redis/go-redis/v9"
	"github.com/typical-developers/public-experience-api/internal/apperror"
	"github.com/typical-developers/public-experience-api/pkg/httpx"
)

var (
	HttpErrorNotFound = apperror.NewAppError("NotFound", "The requested URL does not exist on the server.", http.StatusNotFound, nil)

	HttpInternalServerError    = apperror.NewAppError("InternalServerError", "Something went wrong, please try again later.", http.StatusInternalServerError, nil)
	HttpErrorResourceNotCached = apperror.NewAppError("ResourceNotCached", "The current resource is not cached, please try again later.", http.StatusServiceUnavailable, nil)
)

// WriteRESTError will write an error based on the type of error passed in.
func WriteRESTError(w http.ResponseWriter, err error) {
	var restError *apperror.AppError

	if errors.Is(err, redis.Nil) {
		restError = HttpErrorNotFound
	} else {
		restError = HttpInternalServerError
	}

	if err != nil {
		restError = restError.WithError(err)
	}

	_ = httpx.WriteJSON(w, ErrorResponse{
		Type:    restError.Type,
		Message: restError.Message,
	}, restError.Status)
}
