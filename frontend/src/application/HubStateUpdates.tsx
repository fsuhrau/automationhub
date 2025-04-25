import React, {useEffect} from 'react';
import {useSSE} from 'react-hooks-sse';
import {HubStateActions} from "./HubState";
import {useProjectContext} from "../hooks/ProjectProvider";
import {useHubState} from "../hooks/HubStateProvider";
import {getAllDevices} from "../services/device.service";

interface DeviceChangePayload {
    deviceId: number,
    deviceState: number
}

const HubStateUpdates: React.FC<{ children: any }> = ({children}) => {

    const {dispatch, state} = useHubState()
    const {projectIdentifier} = useProjectContext()

    useEffect(() => {
        if (state.devices == null) {
            getAllDevices(projectIdentifier).then(devices => {
                dispatch({type: HubStateActions.UpdateDevices, payload: devices})
            }).catch(e => {
            });
        }
    }, [state.projects])

    const deviceStateChange = useSSE<DeviceChangePayload | null>('devices', null);

    useEffect(() => {
        if (deviceStateChange === null)
            return;
        dispatch({type: HubStateActions.UpdateDeviceState, payload: deviceStateChange})
    }, [deviceStateChange]);


    return <>
        {children}
    </>
};

export default HubStateUpdates;