package httpauth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/ihleven/pkg/errors"
)

type Carrier interface {
	Synthesize(string, time.Time) (string, error)
	Parse(string) (string, error)
}

// Claims
type Claims struct {
	jwt.StandardClaims
	Username string
}

func NewCookieCarrier(secret []byte) Carrier {
	return &jwtcarrier{secret: secret}
}

type jwtcarrier struct {
	secret []byte
}

// func TokenString(username string, expirationTime time.Time) (string, error)
func (c *jwtcarrier) Synthesize(username string, expiry time.Time) (string, error) {

	if expiry.IsZero() {
		expiry = time.Now().Add(5 * time.Minute)
	}

	claims := Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expiry.Unix(),
		},
	}
	// Create a new token object, specifying signing method and the claims you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(c.secret)
	if err != nil {
		return "", errors.Wrap(err, "Couldn‘t sign token %v", token)
	}

	return tokenString, nil
}

func (c *jwtcarrier) Parse(input string) (string, error) {

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return c.secret, nil
	}

	token, err := jwt.ParseWithClaims(input, &Claims{}, keyFunc)
	if err != nil {
		return "", errors.Wrap(err, "Failed to parse jwt")
	}

	if !token.Valid {
		return "", errors.New("Parsed token is invalid")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return "", errors.New("Claims not of expected type: %v", claims)
	}

	authkey := claims.Username

	return authkey, nil
}
