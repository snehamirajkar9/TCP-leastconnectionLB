package server

type AuthScheme struct {
	allowedClients map[string][]string // client ID -> allowed upstreams
}

func NewAuthScheme() *AuthScheme {
	m := make(map[string][]string)
	return &AuthScheme{
		allowedClients: m,
	}
}

func (a *AuthScheme) AllowClient(clientID string, upstreams []string) {
	a.allowedClients[clientID] = upstreams
}

func (a *AuthScheme) GetAllowedUpstreams(clientID string) []string {
	allowed, ok := a.allowedClients[clientID]
	if !ok {
		return nil
	}
	return allowed
}
