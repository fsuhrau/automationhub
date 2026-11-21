import React, {Dispatch, ReactNode, useEffect, useRef, useState} from 'react';
import {useSSE} from 'react-hooks-sse';
import {DeviceStatePayload, HubStateAction, HubStateActions} from "./HubStateActions";
import {useProjectContext} from "../hooks/ProjectProvider";
import {useHubState} from "../hooks/HubStateProvider";
import {getAllDevices} from "../services/device.service";
import {getAllApps} from "../services/app.service";
import {getNodes} from "../services/settings.service";
import {useError} from "../ErrorProvider";
import {SSE_CONNECTED_EVENT} from "./sseSource";

type HubDispatch = Dispatch<HubStateAction>;

// the SSE reducer keeps a short numbered backlog, so device changes that
// arrive within one render are all applied
interface DeviceChanges {
    seq: number,
    items: { seq: number, change: DeviceStatePayload }[],
}

const BACKLOG = 50;

const DEVICE_RELOAD_DEBOUNCE_MS = 200;

// counts how often the event stream reconnected (e.g. after a master
// restart), because status changes sent while it was down are lost
export const useReconnects = (): number => {
    const [reconnects, setReconnects] = useState(0);
    useEffect(() => {
        let first = true;
        const onConnected = (): void => {
            if (first) {
                // the initial connect; the loaders load anyway
                first = false;
                return;
            }
            setReconnects(n => n + 1);
        };
        window.addEventListener(SSE_CONNECTED_EVENT, onConnected);
        return () => window.removeEventListener(SSE_CONNECTED_EVENT, onConnected);
    }, []);
    return reconnects;
};

const countEvents = {stateReducer: (count: number): number => count + 1};

// reloads the device list once after a burst of events; 0 events do nothing
export const useDeviceListReload = (projectIdentifier: string, dispatch: HubDispatch, events: number): void => {
    useEffect(() => {
        if (events === 0) return;
        // a plugged in phone produces a few events at once; load the list once
        let cancelled = false;
        const timer = setTimeout(() => {
            getAllDevices(projectIdentifier).then(devices => {
                if (!cancelled) {
                    dispatch({type: HubStateActions.UpdateDevices, payload: devices})
                }
            }).catch(() => undefined);
        }, DEVICE_RELOAD_DEBOUNCE_MS);
        return () => {
            cancelled = true;
            clearTimeout(timer);
        };
    }, [events, projectIdentifier, dispatch]);
};

export const useNodesLoad = (projectIdentifier: string, dispatch: HubDispatch, events: number): void => {
    useEffect(() => {
        getNodes(projectIdentifier).then(nodes => {
            dispatch({type: HubStateActions.NodesUpdate, payload: nodes})
        }).catch(() => undefined);
    }, [events, projectIdentifier, dispatch]);
};

// applies the result of a load unless the effect that started it is gone
const applyUnlessCancelled = <T, >(
    load: Promise<T>,
    isCancelled: () => boolean,
    handlers: { apply: (value: T) => void, fail: (e: object) => void },
): void => {
    load.then(value => {
        if (!isCancelled()) handlers.apply(value);
    }).catch(e => {
        if (!isCancelled()) handlers.fail(e);
    });
};

// devices and apps belong to the project; reload both when it changes
export const useProjectDataLoad = (
    projectIdentifier: string,
    dispatch: HubDispatch,
    setError: (e: object) => void,
): void => {
    useEffect(() => {
        let cancelled = false;
        const isCancelled = (): boolean => cancelled;
        applyUnlessCancelled(getAllApps(projectIdentifier), isCancelled, {
            apply: payload => dispatch({type: HubStateActions.AppsUpdate, payload}),
            fail: setError,
        });
        applyUnlessCancelled(getAllDevices(projectIdentifier), isCancelled, {
            apply: payload => dispatch({type: HubStateActions.UpdateDevices, payload}),
            fail: setError,
        });
        return () => {
            cancelled = true;
        };
    }, [projectIdentifier, dispatch, setError]);
};

const collectChanges = (state: DeviceChanges, action: { data: DeviceStatePayload | null }): DeviceChanges => {
    if (!action.data) return state;
    const seq = state.seq + 1;
    return {seq, items: [...state.items, {seq, change: action.data}].slice(-BACKLOG)};
};

// applies the status changes of the "devices" channel in place
export const useDeviceStateChanges = (dispatch: HubDispatch): void => {
    const changes = useSSE<DeviceChanges, DeviceStatePayload | null>('devices', {seq: 0, items: []}, {
        stateReducer: collectChanges,
    });
    const applied = useRef(0);

    useEffect(() => {
        changes.items.filter(i => i.seq > applied.current).forEach(i => {
            dispatch({type: HubStateActions.UpdateDeviceState, payload: i.change})
        });
        applied.current = changes.seq;
    }, [changes, dispatch]);
};

const HubStateUpdates: React.FC<{ children: ReactNode }> = ({children}) => {
    const {dispatch} = useHubState()
    const {projectIdentifier} = useProjectContext()
    const {setError} = useError()
    const reconnects = useReconnects();
    // a node connected or disconnected: its status and its devices changed
    const nodeEvents = useSSE<number, unknown>('nodes', 0, countEvents);
    // a device was added or its name, OS or parameters changed (status
    // changes come on the "devices" channel and are applied in place)
    const deviceListEvents = useSSE<number, unknown>('device_list', 0, countEvents);

    useDeviceListReload(projectIdentifier, dispatch, reconnects + nodeEvents + deviceListEvents);
    useNodesLoad(projectIdentifier, dispatch, reconnects + nodeEvents);
    useProjectDataLoad(projectIdentifier, dispatch, setError);
    useDeviceStateChanges(dispatch);

    return <>
        {children}
    </>
};

export default HubStateUpdates;
