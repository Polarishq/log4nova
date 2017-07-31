package context

import (
	"github.com/go-openapi/runtime"
)

// Interface for client methods
type ClientInterface interface {
	GetContextsID(params *GetContextsIDParams) (*GetContextsIDOK, error)
	PostContexts(params *PostContextsParams) (*PostContextsCreated, error)

	SetTransport(transport runtime.ClientTransport)
}
