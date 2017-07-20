package search

import (
	"github.com/go-openapi/runtime"
)

// Interface for client methods
type ClientInterface interface {
	PostSearch(params *PostSearchParams) (*PostSearchOK, error)

	SetTransport(transport runtime.ClientTransport)
}
