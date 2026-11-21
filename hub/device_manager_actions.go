package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
)

// responseHandler consumes a response of one action type before the generic
// dispatch sees it.
type responseHandler func(d device.Device, resp *action.Response)

func (dm *DeviceManager) responseHandlers() map[action.ActionType]responseHandler {
	return map[action.ActionType]responseHandler{
		action.ActionType_Log:    handleLogResponse,
		action.ActionType_RunKey: handleRunKeyResponse,
	}
}

func handleLogResponse(d device.Device, resp *action.Response) {
	if resp.Payload.LogData != nil {
		logError(d, resp.Payload.LogData)
	}
}

// handleRunKeyResponse hands keys to the running test, whichever executor runs it.
func handleRunKeyResponse(d device.Device, resp *action.Response) {
	if k := resp.Payload.RunKey; k != nil && k.Key != "" {
		d.SetKey(k.Key, k.Value)
	}
}

func (dm *DeviceManager) SendAction(dev device.Device, act action.Interface) error {
	start := time.Now()
	defer func() {
		dm.log.Debugf("send message took %s", time.Since(start).String())
	}()

	buf, err := act.Serialize()
	if err != nil {
		return fmt.Errorf("marshal action: %w", err)
	}

	if !dev.IsAppConnected() {
		return fmt.Errorf("device not connected")
	}

	dm.log.Debugf("Send Action: %s %v", reflect.TypeOf(act).Elem().Name(), act)
	dev.Log("action", "Send Action: %s", reflect.TypeOf(act).Elem().Name())
	return dev.Send(buf)
}

func (dm *DeviceManager) handleActions(d device.Device, ctx context.Context) {
	dm.log.Info("handleActions")

	con := d.Connection()
	if con == nil {
		return
	}

	defer func() {
		dm.log.Info("handleActions finished")
		con.Close()
		// do not reset a newer connection that replaced this one
		if d.Connection() == con {
			d.SetConnection(nil)
		}
	}()

	for {
		select {
		case data, ok := <-con.ResponseChannel:
			if !ok || data.Err != nil {
				dm.log.Info("channel closed")
				return
			}
			dm.dispatchData(d, con, data.Data)
		case <-ctx.Done():
			dm.log.Info("handleActions cancel")
			// executors registered on the device stay registered across app
			// reconnects; only a cancelled device context aborts them
			for _, handler := range d.ActionHandlers() {
				handler.OnActionResponse(d, nil)
			}
			return
		}
	}
}

func (dm *DeviceManager) dispatchData(d device.Device, con *device.Connection, data []byte) {
	resp := action.Response{}
	if err := json.Unmarshal(data, &resp); err != nil {
		dm.log.Errorf("invalid action response from %s: %v", d.DeviceID(), err)
		return
	}
	if h, ok := dm.responseHandler[resp.ActionType]; ok {
		h(d, &resp)
		return
	}
	if handlers := d.ActionHandlers(); len(handlers) > 0 {
		for i := range handlers {
			handlers[i].OnActionResponse(d, &resp)
		}
		return
	}
	dm.deliverResponse(d, con, resp)
}

// deliverResponse passes a response nobody else consumed to the waiting caller.
func (dm *DeviceManager) deliverResponse(d device.Device, con *device.Connection, resp action.Response) {
	if resp.ActionType == action.ActionType_Performance {
		return
	}
	if resp.ActionID == "" {
		d.Log("unhandled", "Received: %v", resp.Payload)
		return
	}
	// nobody is required to consume this channel; never block the read loop
	select {
	case con.ActionChannel <- resp:
	default:
		dm.log.Debugf("dropping unhandled action response %s (%v)", resp.ActionID, resp.ActionType)
	}
}
