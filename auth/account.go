package auth

type AccountI interface {
	Username() string
	SetPassword(string) error
	ValidatePasswordHash(string) error
}

type Account struct {
	ID       uint32 `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
	// Uid, Gid      uint32
	// Username      string
	// Name          string
	// HomeDir       string
	// Authenticated bool
	// Groups []string `json:"groups"`
}

var Matthias Account = Account{1, "matt", "pwd"}
var Wolfgang Account = Account{3, "wolfgang", "gnagflow"}
var Vicky Account = Account{2, "vicky", "pwd2"}

var CurrentUser *Account = &Matthias

var Anonymous Account = Account{0, "anonymous", "none"}

// var CurrentUser *Account = &Matthias
