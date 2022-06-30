package auth

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

var Anonymous Account = Account{0, "anonymous", ""}
