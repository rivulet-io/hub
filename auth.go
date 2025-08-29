package hub

import "github.com/nats-io/nats-server/v2/server"

type AuthMethod func(string) bool

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

	return ca.authMethod(opt.Token)
}

func NewCustomAuthenticator(authMethod AuthMethod) *CustomAuthenticator {
	return &CustomAuthenticator{authMethod: authMethod}
}
