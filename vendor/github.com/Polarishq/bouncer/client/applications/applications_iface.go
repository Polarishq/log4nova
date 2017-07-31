package applications

import (
	"github.com/go-openapi/runtime"
)

// Interface for client methods
type ClientInterface interface {
	GetApplications(params *GetApplicationsParams) (*GetApplicationsOK, error)
	GetApplicationsID(params *GetApplicationsIDParams) (*GetApplicationsIDOK, error)

	SetTransport(transport runtime.ClientTransport)
}
