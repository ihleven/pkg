package httpauth

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// type Accounter interface {
// 	Authkey() string
// SetPassword(string) error
// 	ValidatePasswordHash(string) error
// }

// type Authenticator interface {
// 	Login(r *http.Request, user string) string
// 	Authenticate(string, string) string
// }

// type Account struct {
// 	ID       uint32
// 	Username string
// 	Password string
// 	// Uid, Gid      uint32
// 	// Username      string
// 	// Name          string
// 	// HomeDir       string
// 	// Authenticated bool
// }

func NewAuth(usermap map[string]string, secret []byte) *Auth {
	return &Auth{
		cookieName: "token",
		authmap:    usermap,
		// authenticator: authenticator,
		carrier: &jwtcarrier{secret: secret},
	}
}

type authmap map[string]string

// func (a authmap) GetUser(u string) string {
// 	return a[u]
// }

type Auth struct {
	cookieName string
	authmap    authmap
	// authenticator Authenticator
	carrier Carrier
}

func (a *Auth) Authenticate(username string, password string) string {
	pwdhash, ok := a.authmap[username]
	if ok {
		bcrypterr := bcrypt.CompareHashAndPassword([]byte(pwdhash), []byte(password))
		if bcrypterr == nil {
			return username
		}

		// tmp workaround für plaintext passwords
		if pwdhash == password {
			return username
		}
	}
	return ""
}

func (a *Auth) Login(w http.ResponseWriter, authkey string) {

	expirationTime := time.Now().Add(500 * time.Hour)

	token, err := a.carrier.Synthesize(authkey, expirationTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// scheme := r.Header.Get("X-Forwarded-Proto")
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  expirationTime,
		Secure:   false, //scheme == "https",
		HttpOnly: true,
		// SameSite: http.SameSiteStrictMode,
		Path: "/",
	})
}

func (a *Auth) ParseRequestAuth(r *http.Request) string {
	cookie, err := r.Cookie(a.cookieName)
	if err != nil {
		return ""
	}
	authkey, err := a.carrier.Parse(cookie.Value)
	if err != nil {
		return ""
	}
	return authkey
}

// TODO: get rid of jwtKey / GetClaims
var jwtKey = []byte("my_secret_key")

func GetClaims(r *http.Request) (Claims, int, error) {

	claims := Claims{}

	cookie, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			return claims, http.StatusUnauthorized, err
		}
		return claims, http.StatusBadRequest, err
	}

	// tkn, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
	// 	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
	// 		return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	// 	}
	// 	return jwtKey, nil
	// })

	tkn, err := jwt.ParseWithClaims(cookie.Value, &claims, func(token *jwt.Token) (interface{}, error) { return jwtKey, nil })
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return claims, http.StatusUnauthorized, err
		}
		return claims, http.StatusBadRequest, err
	}
	if !tkn.Valid {
		return claims, http.StatusUnauthorized, err
	}

	// if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
	// 	ctx := context.WithValue(r.Context(), "props", claims)
	// 	// Access context values in handlers like this
	// 	// props, _ := r.Context().Value("props").(jwt.MapClaims)
	// 	next.ServeHTTP(w, r.WithContext(ctx))
	// } else {
	// 	fmt.Println(err)
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	w.Write([]byte("Unauthorized"))
	// }

	return claims, 0, nil
}
