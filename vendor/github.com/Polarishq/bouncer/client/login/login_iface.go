package login

import (
	"github.com/go-openapi/runtime"
)

// Interface for client methods
type ClientInterface interface {
	GetLoginAuthorize(params *GetLoginAuthorizeParams) error
	GetLoginAuthorizeCallback(params *GetLoginAuthorizeCallbackParams) error
	GetLoginProviders(params *GetLoginProvidersParams) (*GetLoginProvidersOK, error)
	PostLoginOauthRevoke(params *PostLoginOauthRevokeParams) (*PostLoginOauthRevokeOK, error)
	PostLoginOauthToken(params *PostLoginOauthTokenParams, authInfo runtime.ClientAuthInfoWriter) (*PostLoginOauthTokenOK, error)
	PostLoginToken(params *PostLoginTokenParams) (*PostLoginTokenOK, error)

	SetTransport(transport runtime.ClientTransport)
}
