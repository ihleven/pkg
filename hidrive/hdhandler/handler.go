package hdhandler

import (
	"embed"
	"fmt"
	"net/http"
	"text/template"

	"github.com/ihleven/pkg/hidrive"
)

//go:embed templates/*
var templates embed.FS

// ServeHTTP has 3 modes:
// 1) GET without params shows a button calling ServeHTTP with the authorize param
//    and a form POSTing ServeHTTP with a code
// 2) GET with authorize param redirects to the hidrive /client/authorize endpoint where a user can authorize the app.
//    This endpoint will call registered token-callback which is not reachable locally ( => copy code and use form from 1)
// 3) POSTing username and code triggering oauth2/token endpoint generating an access token
func AuthHandler(m *hidrive.AuthManager) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		switch r.URL.Path {

		case "/authorize":
			clientAuthURL := m.GetClientAuthorizeURL(r.URL.Query().Get("state"), r.URL.Query().Get("next"))
			http.Redirect(w, r, clientAuthURL, 302)

		case "/authcode":
			// HandleAuthorizeCallback verarbeitet die Weiterleitung nach erfolgtem authorize
			// sollte unter  https://ihle.cloud/hidrive-token-callback erreichbar sein

			key := r.URL.Query().Get("state")
			if code, ok := r.URL.Query()["code"]; ok {
				fmt.Println("code", code, ok)
				_, err := m.AddAuth(key, code[0])
				if err != nil {
					fmt.Printf("error => %+v\n", err)
					fmt.Fprintf(w, "error => %+v\n", err)
				}
				// p.writeTokenFile()

				next := r.URL.Query().Get("next")
				if next == "" {
					next = "/hidrive/auth"
				}
				fmt.Println("code", code, ok, next)
				fmt.Fprintf(w, `<head><meta http-equiv="Refresh" content="0; URL=%s"></head>`, next)
			}

		default:

			if key, found := r.URL.Query()["refresh"]; found {
				m.Refresh(key[0], false)
			}

			t, err := template.ParseFS(templates, "templates/*.html")
			if err != nil {
				fmt.Fprintf(w, "Cannot parse templates: %v", err)
				return
			}

			w.WriteHeader(200)
			t.ExecuteTemplate(w, "tokenmgmt.html", map[string]interface{}{}) // "ClientID": m.clientID, "ClientSecret": m.clientSecret, "tokens": m.authmap})
			return
		}
	}
}

// func (m *hidrive.AuthManager) AuthTokenCallback(w http.ResponseWriter, r *http.Request) {

// 	key := r.URL.Query().Get("state")
// 	code := r.URL.Query().Get("code")

// 	_, err := m.AddAuth(key, code)
// 	if err != nil {
// 		fmt.Printf("error => %+v\n", err)
// 		fmt.Fprintf(w, "error => %+v\n", err)
// 	}
// 	// p.writeTokenFile()

// 	next := r.URL.Query().Get("next")
// 	if next == "" {
// 		next = "/hidrive/auth"
// 	}

// 	fmt.Fprintf(w, `<head><meta http-equiv="Refresh" content="0; URL=%s"></head>`, next)

// }
