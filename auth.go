package hub

import "github.com/nats-io/nats-server/v2/server"

type AuthMethod func(username string, password string, token string) bool

var _ server.Authentication = &CustomAuthenticator{}

type CustomAuthenticator struct {
	authMethod AuthMethod
}

func (ca *CustomAuthenticator) Check(c server.ClientAuthentication) bool {
	if ca.authMethod == nil {
		return true
	}

	opt := c.GetOpts()
	if opt == nil {
		return false
	}

	return ca.authMethod(opt.Username, opt.Password, opt.Token)
}

func NewCustomAuthenticator(authMethod AuthMethod) *CustomAuthenticator {
	return &CustomAuthenticator{authMethod: authMethod}
}
