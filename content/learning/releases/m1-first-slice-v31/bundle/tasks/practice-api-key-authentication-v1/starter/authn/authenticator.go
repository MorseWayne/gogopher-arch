package authn

import "errors"

type Credential struct {
	Subject string
	Token   string
}

type Authenticator struct{}

func New(credentials []Credential) (*Authenticator, error) {
	return nil, errors.New("TODO: validate and hash credentials")
}

func (authenticator *Authenticator) Authenticate(authorization string) (string, bool) {
	return "", false
}
