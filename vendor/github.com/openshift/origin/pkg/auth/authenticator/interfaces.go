package authenticator

import (
	"net/http"

	"github.com/openshift/origin/pkg/auth/api"
	"k8s.io/kubernetes/pkg/auth/user"
)

type Token interface {
	AuthenticateToken(token string) (user.Info, bool, error)
}

type Request interface {
	AuthenticateRequest(req *http.Request) (user.Info, bool, error)
}

type Password interface {
	AuthenticatePassword(user, password string) (user.Info, bool, error)
}

type Assertion interface {
	AuthenticateAssertion(assertionType, data string) (user.Info, bool, error)
}

type Client interface {
	AuthenticateClient(client api.Client) (user.Info, bool, error)
}

type RequestFunc func(req *http.Request) (user.Info, bool, error)

func (f RequestFunc) AuthenticateRequest(req *http.Request) (user.Info, bool, error) {
	return f(req)
}
