package events

import (
	"github.com/go-openapi/runtime"
)

// Interface for client methods
type ClientInterface interface {
	GetEvents(params *GetEventsParams, authInfo runtime.ClientAuthInfoWriter) (*GetEventsOK, error)
	Events(params *EventsParams, authInfo runtime.ClientAuthInfoWriter) (*EventsOK, error)

	SetTransport(transport runtime.ClientTransport)
}
