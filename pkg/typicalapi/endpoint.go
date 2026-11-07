package typicalapi

import (
	"fmt"
	"net/url"
	"strings"
)

type Endpoint struct {
	Domain string
	Route  string
}

func NewEndpoint(route string) *Endpoint {
	return &Endpoint{
		Route: route,
	}
}

func (e *Endpoint) WithDomain(domain string) *Endpoint {
	e.Domain = domain
	return e
}

func (e *Endpoint) Compile(query url.Values, params ...any) string {
	path := e.Route

	for _, p := range params {
		start := strings.Index(path, "{")
		end := strings.Index(path, "}")

		if start == -1 || end < start {
			break
		}

		path = path[:start] + url.PathEscape(fmt.Sprint(p)) + path[end+1:]
	}

	if q := query.Encode(); q != "" {
		path += "?" + q
	}

	return path
}
