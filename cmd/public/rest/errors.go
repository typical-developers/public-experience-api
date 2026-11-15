package rest

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/redis/go-redis/v9"
	"github.com/typical-developers/public-experience-api/internal/apperror"
)

var (
	HttpErrorNotFound         = apperror.NewAppError("NotFound", "The requested URL does not exist on the server.", http.StatusNotFound, nil)
	HttpErrorResourceNotFound = apperror.NewAppError("ResourceNotFound", "The requested resource does not exist.", http.StatusNotFound, nil)

	HttpInternalServerError    = apperror.NewAppError("InternalServerError", "Something went wrong, please try again later.", http.StatusInternalServerError, nil)
	HttpErrorResourceNotCached = apperror.NewAppError("ResourceNotCached", "The current resource is not cached, please try again later.", http.StatusServiceUnavailable, nil)
)

// ErrNotFound marks a specific resource as not existing, or as hidden from the client.
var ErrNotFound = errors.New("resource not found")

// retryAfterSeconds is sent with ResourceNotCached so clients know when to try again.
const retryAfterSeconds = "60"

// WriteRESTError will write an error based on the type of error passed in.
//
// A redis.Nil is treated as data that should exist but has not been cached yet (503).
// Use WriteLookupError when fetching a single resource by a client-supplied identifier.
func WriteRESTError(w http.ResponseWriter, r *http.Request, err error) {
	var restError *apperror.AppError

	switch {
	case errors.Is(err, ErrNotFound):
		restError = HttpErrorResourceNotFound
	case errors.Is(err, redis.Nil):
		restError = HttpErrorResourceNotCached
		w.Header().Set("Retry-After", retryAfterSeconds)
	default:
		restError = HttpInternalServerError
	}

	if err != nil {
		restError = restError.WithError(err)
	}

	render.Status(r, restError.Status)
	render.JSON(w, r, ErrorResponse{
		Type:    restError.Type,
		Message: restError.Message,
	})
}

// WriteLookupError will write an error for a single resource fetched by a client-supplied identifier.
// A redis.Nil means the resource does not exist (404), any other error is handled by WriteRESTError.
func WriteLookupError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, redis.Nil) {
		err = ErrNotFound
	}

	WriteRESTError(w, r, err)
}
