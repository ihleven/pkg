package auth

type AccountI interface {
	Username() string
	SetPassword(string) error
	ValidatePasswordHash(string) error
}

type Account struct {
	ID       uint32
	Username string
	Password string
	// Uid, Gid      uint32
	// Username      string
	// Name          string
	// HomeDir       string
	// Authenticated bool
}

var Matthias Account = Account{1, "matt", "pwd"}
var Wolfgang Account = Account{3, "wolfgang", "gnagflow"}
var Vicky Account = Account{2, "vicky", "pwd2"}

var CurrentUser *Account = &Matthias
