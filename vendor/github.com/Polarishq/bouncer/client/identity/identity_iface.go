package identity

import (
	"github.com/go-openapi/runtime"
)

// Interface for client methods
type ClientInterface interface {
	DeleteAccountAPIKeysClientID(params *DeleteAccountAPIKeysClientIDParams, authInfo runtime.ClientAuthInfoWriter) (*DeleteAccountAPIKeysClientIDOK, error)
	GetAccount(params *GetAccountParams, authInfo runtime.ClientAuthInfoWriter) (*GetAccountOK, error)
	GetAccountAPIKeys(params *GetAccountAPIKeysParams, authInfo runtime.ClientAuthInfoWriter) (*GetAccountAPIKeysOK, error)
	GetAccountAPIKeysClientID(params *GetAccountAPIKeysClientIDParams, authInfo runtime.ClientAuthInfoWriter) (*GetAccountAPIKeysClientIDOK, error)
	GetTenants(params *GetTenantsParams, authInfo runtime.ClientAuthInfoWriter) (*GetTenantsOK, error)
	PostAccountAPIKeys(params *PostAccountAPIKeysParams, authInfo runtime.ClientAuthInfoWriter) (*PostAccountAPIKeysCreated, error)
	PostTenants(params *PostTenantsParams, authInfo runtime.ClientAuthInfoWriter) (*PostTenantsOK, error)
	PutAccount(params *PutAccountParams, authInfo runtime.ClientAuthInfoWriter) (*PutAccountOK, error)
	PutAccountAPIKeysClientID(params *PutAccountAPIKeysClientIDParams, authInfo runtime.ClientAuthInfoWriter) (*PutAccountAPIKeysClientIDOK, error)

	SetTransport(transport runtime.ClientTransport)
}
