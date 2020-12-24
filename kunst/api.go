package kunst

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/ihleven/errors"
	"github.com/ihleven/pkg/auth"
	"github.com/ihleven/pkg/hidrive"
)

func ApiHandler(drive *hidrive.Drive, repo *Repo, usermap map[string]string) http.HandlerFunc {

	BilderHandler := BildHandler{repo, drive}
	SerienHandler := SerieHandler{repo, drive}
	AusstellungsHandler := AusstellungHandler{repo, drive}

	return func(w http.ResponseWriter, r *http.Request) {

		head, id, tail := parseURLPath(r.URL.Path)

		claims, _, err := auth.GetClaims(r)
		if err != nil && head != "signin" {
			http.Error(w, "", 401)
			return
		}
		username := claims.Username

		// prefix, _ := shiftPath(r.URL.Path)

		var response interface{}

		switch head {
		case "":
			fmt.Fprintf(w, "username: %v", username)

		case "bilder":
			if tail == "/hidrive" {
				err = BilderHandler.Upload(w, r, id, username)
			} else {
				err = BilderHandler.Dispatch(w, r, id, username)
			}

		case "serien":
			err = SerienHandler.Dispatch(w, r, id, tail, username)

		case "ausstellungen":
			switch {
			case id == 0:
				err = AusstellungsHandler.ListCreateAusstellungen(w, r, username)
			case id > 0 && tail == "":
				response, err = AusstellungsHandler.GetUpdateDeleteAusstellung(r, id, username)

			case tail == "/documents":
				response, err = AusstellungsHandler.AusstellungDocuments(r, id, username)

			}
		case "fotos":
			switch r.Method {
			// case "PATCH":
			// 	a.repo.Update("foto", id, map[string]interface{}{"kommentar": "DELETE"})
			case "DELETE":
				err = repo.Update("foto", id, map[string]interface{}{"kommentar": "DELETE"})
			}

		case "thumbs":

			body, err := drive.Thumbnail(tail, r.URL.Query(), username)
			if err != nil {
				break
			}
			defer body.Close()
			if _, err := io.Copy(w, body); err != nil {
				err = errors.Wrap(err, "failed to Copy hidrive file to responseWriter")
			}

		case "file":

			var body io.ReadCloser
			body, err = drive.File(tail, username)
			if err == nil {
				defer body.Close()
				if _, err := io.Copy(w, body); err != nil {
					err = errors.Wrap(err, "failed to Copy hidrive file to responseWriter")
				}
			}
		case "auth":
			// hidrive authentifizierung
			// das sollte sich zur auth admin page ausweiten: * übersicht hd accounts, ‘nen account authentifizieren/löschen, lokale auth als matt notwendig um auf die Seite zu kommen
			drive.Auth.ServeHTTP(w, r)

		case "refresh":
			err = drive.Auth.ForceRefresh(username)

		case "signin":
			// lokale authentifizierung mit usernamen wolfgang / matt
			auth.SigninHandler(usermap)(w, r)
		case "welcome":
			// lokale authentifizierung mit usernamen wolfgang / matt
			auth.Welcome(w, r)

		case "signout":
			auth.SignoutHandler(w, r)

		}

		if err != nil {

			code := errors.Code(err)
			if code == 65535 {
				code = 500
			}

			http.Error(w, err.Error(), code)
			return
		}
		if response != nil {
			render(response, w)
		} else {
			// w.WriteHeader(http.StatusNoContent)
		}
	}

}

func shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

func parseURLPath(p string) (string, int, string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], 0, ""
	}
	handler := p[1:i]
	i2 := strings.Index(p[i+1:], "/")
	if i2 == -1 {
		i2 = len(p[i+1:])
	}
	idstr := p[i+1 : i+i2+1]
	id, err := strconv.Atoi(idstr)
	if err == nil {
		return handler, id, p[i+i2+1:]
	}

	return handler, 0, p[i:]
}

// HTTP middleware setting a value on the request context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path == "/signin" {
			next.ServeHTTP(w, r)
			return
		}
		claims, _, err := auth.GetClaims(r)
		if err != nil {
			http.Error(w, "not cookie auth", 401)
		} else {
			ctx := context.WithValue(r.Context(), "username", claims.Username)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		next.ServeHTTP(w, r)
	})
}

// func addCorsHeader(res http.ResponseWriter) {
// 	headers := res.Header()
// 	headers.Add("Access-Control-Allow-Origin", "*")
// 	headers.Add("Vary", "Origin")
// 	headers.Add("Vary", "Access-Control-Request-Method")
// 	headers.Add("Vary", "Access-Control-Request-Headers")
// 	headers.Add("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, token")
// 	headers.Add("Access-Control-Allow-Methods", "GET, POST,OPTIONS")
// }

func UploadFileBinary(w http.ResponseWriter, r *http.Request) {
	file, err := ioutil.TempFile("media", "binary-*")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	n, err := io.Copy(file, r.Body)
	if err != nil {
		panic(err)
	}
	w.Write([]byte(fmt.Sprintf("%d bytes reveived", n)))
}
