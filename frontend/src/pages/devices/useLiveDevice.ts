import IDeviceData from '../../types/device';
import {INodeData} from '../../types/node';
import {useHubState} from '../../hooks/HubStateProvider';
import {mergeLiveDevice, resolveNode} from './deviceDetailsModel';

// useLiveDevice is the loaded device with the live status and lease the
// device list keeps up to date, and the node it is connected to
export function useLiveDevice(stored: IDeviceData): { device: IDeviceData, node?: INodeData } {
    const {state} = useHubState();
    const device = mergeLiveDevice(stored, state.devices?.find(d => d.id === stored.id));
    return {device, node: resolveNode(device, state.nodes)};
}
