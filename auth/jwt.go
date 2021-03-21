package auth

import (
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ihleven/errors"
)

var jwtKey = []byte("my_secret_key")

// Claims
type Claims struct {
	Username           string `json:"username"`
	jwt.StandardClaims `json:"-"`
}

func TokenString(username string, expirationTime time.Time) (string, error) {

	// Declare the expiration time of the token
	// here, we have kept it as 5 minutes
	if expirationTime.IsZero() {
		expirationTime = time.Now().Add(5 * time.Minute)
	}
	// Create the JWT claims, which includes the username and expiry time
	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", errors.Wrap(err, "Couldn‘t sign token %v", token)
	}

	return tokenString, nil
}

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
