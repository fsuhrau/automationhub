package api

import (
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	hubmcp "github.com/fsuhrau/automationhub/endpoints/mcp"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/fsuhrau/automationhub/service/access"
	appsvc "github.com/fsuhrau/automationhub/service/app"
	devicesvc "github.com/fsuhrau/automationhub/service/device"
	overviewsvc "github.com/fsuhrau/automationhub/service/overview"
	"github.com/fsuhrau/automationhub/service/project"
	protocolsvc "github.com/fsuhrau/automationhub/service/protocol"
	runsvc "github.com/fsuhrau/automationhub/service/run"
	testsvc "github.com/fsuhrau/automationhub/service/test"
)

// Deps are the services the API serves; the composition root builds each
// of them once.
type Deps struct {
	Logger httpx.Logger
	// NodeURL is the address of the node the control video connects to.
	NodeURL string
	// Cors lists the origins that may open the video websocket.
	Cors []string
	// DataDir is the folder of the test data served below /api/data.
	DataDir string

	Projects  *project.Service
	Access    *access.Service
	Apps      *appsvc.Service
	Tests     *testsvc.Service
	Runs      *runsvc.Service
	Protocols *protocolsvc.Service
	Overview  *overviewsvc.Service
	Devices   *devicesvc.Service

	Control *hubcontrol.Service
	Broker  *sse.Broker
	MCP     *hubmcp.Handler
}
