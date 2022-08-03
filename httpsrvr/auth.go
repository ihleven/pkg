package httpsrvr

import "net/http"

type AuthParser interface {
	ParseRequestAuth(*http.Request) string
}
