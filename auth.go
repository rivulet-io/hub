package hub

import "github.com/nats-io/nats-server/v2/server"

var _ server.Authentication = &CustomAuthenticator{}

type CustomAuthenticator struct {
	authMethod func(string) bool
}

func (ca *CustomAuthenticator) Check(c server.ClientAuthentication) bool {
	if ca.authMethod == nil {
		return true
	}

	opt := c.GetOpts()
	if opt == nil {
		return false
	}

	return ca.authMethod(opt.Token)
}

func NewCustomAuthenticator(authMethod func(string) bool) *CustomAuthenticator {
	return &CustomAuthenticator{authMethod: authMethod}
}
