import React, {useEffect, useRef, useState} from 'react';
import {useSSE} from 'react-hooks-sse';
import {HubStateActions} from "./HubState";
import {useProjectContext} from "../hooks/ProjectProvider";
import {useHubState} from "../hooks/HubStateProvider";
import {getAllDevices} from "../services/device.service";
import {getAllApps} from "../services/app.service";
import {useError} from "../ErrorProvider";
import {SSE_CONNECTED_EVENT} from "./sseSource";

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

    // reload devices whenever the event stream reconnects (e.g. after a master
    // restart), because status changes sent while it was down are lost
    const [reconnects, setReconnects] = useState(0);
    useEffect(() => {
        let first = true;
        const onConnected = () => {
            if (first) {
                // the initial connect; the effect below loads anyway
                first = false;
                return;
            }
            setReconnects(n => n + 1);
        };
        window.addEventListener(SSE_CONNECTED_EVENT, onConnected);
        return () => window.removeEventListener(SSE_CONNECTED_EVENT, onConnected);
    }, []);

    useEffect(() => {
        if (reconnects === 0) return;
        getAllDevices(projectIdentifier).then(devices => {
            dispatch({type: HubStateActions.UpdateDevices, payload: devices})
        }).catch(() => undefined);
    }, [reconnects, projectIdentifier, dispatch]);

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
