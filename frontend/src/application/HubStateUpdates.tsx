import React, {useEffect, useRef} from 'react';
import {useSSE} from 'react-hooks-sse';
import {HubStateActions} from "./HubState";
import {useProjectContext} from "../hooks/ProjectProvider";
import {useHubState} from "../hooks/HubStateProvider";
import {getAllDevices} from "../services/device.service";
import {getAllApps} from "../services/app.service";
import {useError} from "../ErrorProvider";

interface DeviceChangePayload {
    deviceId: number,
    deviceState: number
}

// the SSE reducer keeps a short numbered backlog, so device changes that
// arrive within one render are all applied
interface DeviceChanges {
    seq: number,
    items: { seq: number, change: DeviceChangePayload }[],
}

const BACKLOG = 50;

const HubStateUpdates: React.FC<{ children: any }> = ({children}) => {

    const {dispatch} = useHubState()
    const {projectIdentifier} = useProjectContext()
    const {setError} = useError()

    // devices and apps belong to the project; reload both when it changes
    useEffect(() => {
        let cancelled = false;
        getAllApps(projectIdentifier).then(apps => {
            if (!cancelled) {
                dispatch({type: HubStateActions.AppsUpdate, payload: apps})
            }
        }).catch(e => {
            if (!cancelled) {
                setError(e);
            }
        });
        getAllDevices(projectIdentifier).then(devices => {
            if (!cancelled) {
                dispatch({type: HubStateActions.UpdateDevices, payload: devices})
            }
        }).catch(e => {
            if (!cancelled) {
                setError(e);
            }
        });
        return () => {
            cancelled = true;
        };
    }, [projectIdentifier, dispatch, setError])

    const changes = useSSE<DeviceChanges, DeviceChangePayload | null>('devices', {seq: 0, items: []}, {
        stateReducer: (state, action) => {
            if (!action.data) return state;
            const seq = state.seq + 1;
            return {seq, items: [...state.items, {seq, change: action.data}].slice(-BACKLOG)};
        }
    });
    const applied = useRef(0);

    useEffect(() => {
        changes.items.filter(i => i.seq > applied.current).forEach(i => {
            dispatch({type: HubStateActions.UpdateDeviceState, payload: i.change})
        });
        applied.current = changes.seq;
    }, [changes, dispatch]);

    return <>
        {children}
    </>
};

export default HubStateUpdates;
