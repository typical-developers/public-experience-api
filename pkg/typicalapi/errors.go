package typicalapi

import "fmt"

type Error struct {
	Type     string    `json:"type"`
	Message  string    `json:"message"`
	Response *Response `json:"-"`
}

func (e *Error) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("typicalapi: %s", e.Response.Status)
	}

	return fmt.Sprintf("typicalapi: %s: %s (%s)", e.Response.Status, e.Message, e.Type)
}
