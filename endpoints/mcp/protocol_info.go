package mcp

import (
	"time"

	"github.com/fsuhrau/automationhub/storage/models"
)

type protocolInfo struct {
	ID       uint     `json:"id"`
	ParentID *uint    `json:"parentId,omitempty"`
	Name     string   `json:"name"`
	DeviceID *uint    `json:"deviceId,omitempty"`
	Device   string   `json:"device,omitempty"`
	Result   string   `json:"result"`
	Duration string   `json:"duration"`
	AvgFPS   *float64 `json:"avgFps,omitempty"`
	AvgMEM   *float64 `json:"avgMem,omitempty"`
	AvgCPU   *float64 `json:"avgCpu,omitempty"`
	// keys the app reported for the test through the plugin, e.g. player_id
	Keys map[string]string `json:"keys,omitempty"`
}

func duration(p *models.TestProtocol) string {
	if p.EndedAt == nil {
		return "running"
	}
	return p.EndedAt.Sub(p.StartedAt).Round(time.Second).String()
}

func nonZero(v float64) *float64 {
	if v == 0 || v != v { // skip zero and NaN
		return nil
	}
	return &v
}

func toProtocolInfo(p *models.TestProtocol) protocolInfo {
	info := protocolInfo{
		ID: p.ID, ParentID: p.ParentTestProtocolID, Name: p.TestName, DeviceID: p.DeviceID,
		Result: resultNames()[p.TestResult], Duration: duration(p),
		AvgFPS: nonZero(p.AvgFPS), AvgMEM: nonZero(p.AvgMEM), AvgCPU: nonZero(p.AvgCPU),
	}
	if p.Device != nil {
		info.Device = p.Device.Name
	}
	if len(p.Keys) > 0 {
		info.Keys = make(map[string]string, len(p.Keys))
		for _, k := range p.Keys {
			info.Keys[k.Key] = k.Value
		}
	}
	return info
}

func toProtocolInfos(list []models.TestProtocol) []protocolInfo {
	infos := make([]protocolInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toProtocolInfo(&list[i]))
	}
	return infos
}
