package hidrive

import "sync"

type Token struct {
	sync.RWMutex
	Client *OAuth2Client
}

func (t *Token) AccessToken() string {

	return ""
}
