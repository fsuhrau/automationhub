package hub

import (
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
	"time"

	"github.com/sirupsen/logrus"
)

var ErrNodeNotFound = fmt.Errorf("node not found")
var ErrNodeNotConnected = fmt.Errorf("node not connected")

const (
	NodeConnectionTimeout   = 5 * time.Minute
	NodeHealthCheckInterval = 5 * time.Second
)

type NodeManager struct {
	db    *gorm.DB
	nodes map[manager.NodeIdentifier]Node
	stop  bool
	log   *logrus.Entry
}

func NewNodeManager(logger *logrus.Logger, db *gorm.DB) *NodeManager {
	return &NodeManager{log: logger.WithFields(logrus.Fields{
		"prefix": "dm",
	}),
		db:    db,
		nodes: make(map[manager.NodeIdentifier]Node),
	}
}

func (nm *NodeManager) Run(ctx context.Context) {
	nm.log.Debugf("Start NodeManager")
	go func() {
		for {
			select {
			case <-ctx.Done():
				nm.log.Infof("Stop NodeManager")
				return
			default:
			}

			for k, m := range nm.nodes {
				if m.Handler != nil {
					if err := m.Handler.Ping(); err != nil {
						m.Status = NodeStatusDisconnected
						m.Handler = nil
						nm.nodes[k] = m
					}
				}
			}
			time.Sleep(NodeHealthCheckInterval)
		}
	}()
}

func (nm *NodeManager) GetNodes() []manager.NodeIdentifier {
	var nodes []manager.NodeIdentifier
	for k, _ := range nm.nodes {
		nodes = append(nodes, k)
	}
	return nodes
}

func (nm *NodeManager) GetManagers(nodeIdentifier manager.NodeIdentifier) (map[string][]device.Device, error) {
	if n, ok := nm.nodes[nodeIdentifier]; ok && n.Handler != nil {
		managers, err := n.Handler.GetDevices()
		if err != nil {
			nm.log.Errorf("fail to fetch managers from node '%s': %v", nodeIdentifier, err)
			return nil, err
		}
		return managers, nil
	}
	return nil, fmt.Errorf("node '%s' not found", nodeIdentifier)
}

func (nm *NodeManager) IsNodeKnown(nodeIdentifier manager.NodeIdentifier) bool {
	var dbNode models.Node
	if err := nm.db.First(&dbNode, "identifier = ?", nodeIdentifier).Error; err == gorm.ErrRecordNotFound {
		return false
	}
	return true
}

func (nm *NodeManager) RegisterNode(nodeIdentifier manager.NodeIdentifier, client manager.RPCClient, hostname, address, operationSystem, environment string, port int32, managers []string) {
	/*
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			fmt.Println(err)
		}
	*/
	nm.nodes[nodeIdentifier] = Node{
		Identifier:      string(nodeIdentifier),
		Hostname:        hostname,
		Address:         address,
		Port:            port,
		OperationSystem: operationSystem,
		Path:            environment,
		Managers:        managers,
		Status:          NodeStatusConnected,
		Handler:         client,
	}
}

func (nm *NodeManager) getNode(nodeIdentifier manager.NodeIdentifier) (error, *Node) {
	if v, ok := nm.nodes[nodeIdentifier]; ok {
		return nil, &v
	}

	return ErrNodeNotFound, nil
}

func (nm *NodeManager) getHandler(nodeIdentifier manager.NodeIdentifier) (manager.RPCClient, error) {
	err, n := nm.getNode(nodeIdentifier)
	if err != nil {
		return nil, err
	}

	if n.Handler == nil {
		return nil, ErrNodeNotConnected
	}
	return n.Handler, nil
}

func (nm *NodeManager) GetStatus(nodeIdentifier manager.NodeIdentifier) (int, error) {
	err, n := nm.getNode(nodeIdentifier)
	if err != nil {
		return 0, err
	}

	return int(n.Status), nil
}

// Manager Actions
func (nm *NodeManager) StartDevice(nodeIdentifier manager.NodeIdentifier, deviceId string) error {
	handler, err := nm.getHandler(nodeIdentifier)
	if err != nil {
		return err
	}

	return handler.StartDevice(deviceId)
}
func (nm *NodeManager) StopDevice(nodeIdentifier manager.NodeIdentifier, deviceId string) error {
	handler, err := nm.getHandler(nodeIdentifier)
	if err != nil {
		return err
	}

	return handler.StopDevice(deviceId)
}

// device Actions
func (nm *NodeManager) IsAppInstalled(nodeIdentifier manager.NodeIdentifier, deviceId string, parameter *app.Parameter) (bool, error) {
	handler, err := nm.getHandler(nodeIdentifier)
	if err != nil {
		return false, err
	}

	return handler.IsAppInstalled(deviceId, parameter)
}

func (nm *NodeManager) IsAppUploaded(nodeIdentifier manager.NodeIdentifier, parameter *app.Parameter) (bool, error) {
	handler, err := nm.getHandler(nodeIdentifier)
	if err != nil {
		return false, err
	}

	return handler.IsAppUploaded(parameter)
}

func (nm *NodeManager) UploadApp(ctx context.Context, nodeIdentifier manager.NodeIdentifier, parameter *app.Parameter) error {
	handler, err := nm.getHandler(nodeIdentifier)
	if err != nil {
		return err
	}

	// check if we need to upload or of it exists already
	exists, err := handler.IsAppUploaded(parameter)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	return handler.UploadApp(ctx, parameter)
}

func (nm *NodeManager) InstallApp(nodeIdentifier manager.NodeIdentifier, deviceId string, parameter *app.Parameter) error {
	handler, err := nm.getHandler(nodeIdentifier)
	if err != nil {
		return err
	}

	return handler.InstallApp(deviceId, parameter)
}

func (nm *NodeManager) UninstallApp(nodeIdentifier manager.NodeIdentifier, deviceId string, parameter *app.Parameter) error {
	handler, err := nm.getHandler(nodeIdentifier)
	if err != nil {
		return err
	}

	return handler.UninstallApp(deviceId, parameter)
}

func (nm *NodeManager) StartApp(nodeIdentifier manager.NodeIdentifier, deviceId string, config *device.DeviceConfig, parameter *app.Parameter, sessionId string, nodeUrl string) error {
	handler, err := nm.getHandler(nodeIdentifier)
	if err != nil {
		return err
	}

	return handler.StartApp(deviceId, config, parameter, sessionId, nodeUrl)
}

func (nm *NodeManager) StopApp(nodeIdentifier manager.NodeIdentifier, deviceId string, parameter *app.Parameter) error {
	handler, err := nm.getHandler(nodeIdentifier)
	if err != nil {
		return err
	}

	return handler.StopApp(deviceId, parameter)
}

func (nm *NodeManager) IsConnected(nodeIdentifier manager.NodeIdentifier, deviceId string) bool {
	handler, err := nm.getHandler(nodeIdentifier)
	if err != nil {
		return false
	}

	return handler.IsConnected(deviceId)
}

func (nm *NodeManager) StartRecording(nodeIdentifier manager.NodeIdentifier, deviceId string, path string) error {
	handler, err := nm.getHandler(nodeIdentifier)
	if err != nil {
		return err
	}

	return handler.StartRecording(deviceId, "")
}

func (nm *NodeManager) StopRecording(nodeIdentifier manager.NodeIdentifier, deviceId string) error {
	handler, err := nm.getHandler(nodeIdentifier)
	if err != nil {
		return err
	}

	return handler.StopRecording(deviceId)
}

func (nm *NodeManager) GetScreenshot(nodeIdentifier manager.NodeIdentifier, deviceId string) ([]byte, int, int, error) {
	handler, err := nm.getHandler(nodeIdentifier)
	if err != nil {
		return nil, 0, 0, err
	}

	return handler.GetScreenshot(deviceId)
}

func (nm *NodeManager) HasFeature(nodeIdentifier manager.NodeIdentifier, deviceId string, feature string) bool {
	handler, err := nm.getHandler(nodeIdentifier)
	if err != nil {
		return false
	}

	return handler.HasFeature(deviceId, feature)
}

func (nm *NodeManager) Execute(nodeIdentifier manager.NodeIdentifier, deviceId string, data string) {
	handler, err := nm.getHandler(nodeIdentifier)
	if err != nil {
		return
	}

	handler.Execute(deviceId, data)
}

func (nm *NodeManager) ConnectionTimeout(nodeIdentifier manager.NodeIdentifier, deviceId string) time.Duration {
	handler, err := nm.getHandler(nodeIdentifier)
	if err != nil {
		return 5 * time.Minute
	}

	return handler.ConnectionTimeout(deviceId)
}

func (nm *NodeManager) RunNativeScript(nodeIdentifier manager.NodeIdentifier, deviceId string, script []byte) {
	handler, err := nm.getHandler(nodeIdentifier)
	if err != nil {
		return
	}

	handler.RunNativeScript(deviceId, script)
}

func (nm *NodeManager) SendAction(nodeIdentifier manager.NodeIdentifier, deviceId string, action []byte) {
	handler, err := nm.getHandler(nodeIdentifier)
	if err != nil {
		return
	}

	handler.SendAction(deviceId, action)
}
