package session

import (
	"drive/domain"
	"encoding/gob"
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
)

var SESSION_KEY = "aslkdjfheqhlieufhkasjbd"
var SESSION_NAME = "gosessionid"

// store will hold all session data
var store = sessions.NewCookieStore([]byte("something-very-secret"))

func init() {

	//authKeyOne := securecookie.GenerateRandomKey(64)
	//encryptionKeyOne := securecookie.GenerateRandomKey(32)
	//store = sessions.NewCookieStore(
	//	[]byte("asdaskdhasdhgsajdgasdsadksakdhasidoajsdousahdopj"),
	//	[]byte("hhfjhtdzjtfkhgkjfkufkztfjztfkuztfkztdhtesrgesdjg"),
	//)
	store.Options = &sessions.Options{
		//Domain:   "localhost",
		Path:     "/",
		MaxAge:   3600 * 8, // 8 hours
		HttpOnly: true,
		//Secure:   false,
	}
	gob.Register(domain.Account{})
}

type Session struct {
	Session *sessions.Session
	r       *http.Request
	w       http.ResponseWriter
}

// Clear the current session
func (s *Session) Clear() {
	for k := range s.Session.Values {
		s.Delete(k)
	}
}

// Delete a value from the current session.
func (s *Session) Delete(name interface{}) {
	delete(s.Session.Values, name)
}

func (s *Session) Get(name interface{}) interface{} {
	return s.Session.Values[name]
}
func (s *Session) GetString(name string) string {
	str := s.Session.Values[name].(string)
	return str
}

// GetOnce gets a value from the current session and then deletes it.
func (s *Session) GetOnce(name interface{}) interface{} {
	if x, ok := s.Session.Values[name]; ok {
		s.Delete(name)
		return x
	}
	return nil
}

func (s *Session) Save(r *http.Request, w http.ResponseWriter) error {
	return s.Session.Save(r, w)
}

// Set a value onto the current session. If a value with that name
// already exists it will be overridden with the new value.
func (s *Session) Set(name, value interface{}) {
	s.Session.Values[name] = value
}

// Get a session using a request and response.
func GetSession(r *http.Request, w http.ResponseWriter) (*Session, error) {

	session, err := store.Get(r, SESSION_NAME)
	if err != nil {
		//fmt.Println("ERROR session.GetSession", err.Error())
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	return &Session{
		Session: session,
		r:       r,
		w:       w,
	}, nil
}

func Get(r *http.Request, name interface{}) (interface{}, error) {

	session, err := GetSession(r, nil)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("empty session")
	}

	return session.Get(name), nil
}

func GetString(r *http.Request, name interface{}) string {
	session, _ := GetSession(r, nil)
	if session == nil {
		return ""
	}
	var value interface{}
	value = session.Get(name)
	if value != nil {
		return value.(string)
	}
	return ""
}

func Set(r *http.Request, w http.ResponseWriter, key, value interface{}) error {
	session, err := GetSession(r, nil)
	if err != nil {
		return err
	}
	session.Set(key, value)
	err = session.Save(r, w)
	if err != nil {
		return err //errors.Augment(err, errors.Session, "Session could not be saved")
	}
	return nil
}

func GetSessionUser(r *http.Request, w http.ResponseWriter) (*domain.Account, error) {
	sess, err := GetSession(r, w)
	if err != nil {
		return nil, err
	}

	if user, ok := sess.Get("user").(domain.Account); ok {
		return &user, nil
	} else {
		return &domain.Account{Authenticated: false}, nil
	}
}

func SetSessionUser(r *http.Request, w http.ResponseWriter, user *domain.Account) (err error) {
	sess, err := GetSession(r, w)
	if err != nil {
		return
	}
	sess.Set("user", user)
	sess.Save(r, w)
	return
}
