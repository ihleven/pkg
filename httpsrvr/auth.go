package httpsrvr

import "net/http"

type authParser interface {
	ParseRequestAuth(*http.Request) string
}
