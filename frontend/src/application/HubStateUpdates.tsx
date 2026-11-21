import React, {Dispatch, ReactNode, useEffect, useRef, useState} from 'react';
import {useSSE} from 'react-hooks-sse';
import {DeviceStatePayload, HubStateAction, HubStateActions} from "./HubStateActions";
import {useParams} from "react-router-dom";
import {useHubState} from "../hooks/HubStateProvider";
import {useDeviceService} from "../services/device.service";
import {useAppService} from "../services/app.service";
import {useSettingsService} from "../services/settings.service";
import {useProjectService} from "../project/project.service";
import IProject from "../project/project";
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
    const devicesApi = useDeviceService();
    useEffect(() => {
        if (events === 0) return;
        // a plugged in phone produces a few events at once; load the list once
        let cancelled = false;
        const timer = setTimeout(() => {
            devicesApi.list(projectIdentifier).then(devices => {
                if (!cancelled) {
                    dispatch({type: HubStateActions.UpdateDevices, payload: devices})
                }
            }).catch(() => undefined);
        }, DEVICE_RELOAD_DEBOUNCE_MS);
        return () => {
            cancelled = true;
            clearTimeout(timer);
        };
    }, [events, projectIdentifier, dispatch, devicesApi]);
};

export const useNodesLoad = (projectIdentifier: string, dispatch: HubDispatch, events: number): void => {
    const settings = useSettingsService();
    useEffect(() => {
        settings.nodes(projectIdentifier).then(nodes => {
            dispatch({type: HubStateActions.NodesUpdate, payload: nodes})
        }).catch(() => undefined);
    }, [events, projectIdentifier, dispatch, settings]);
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

// the project list is needed by every project page; load it once
export const useProjectsLoad = (projects: IProject[] | null, dispatch: HubDispatch, setError: (e: object) => void): void => {
    const projectApi = useProjectService();
    useEffect(() => {
        if (projects !== null) return;
        let cancelled = false;
        applyUnlessCancelled(projectApi.list(), () => cancelled, {
            apply: payload => dispatch({type: HubStateActions.ProjectsUpdate, payload}),
            fail: setError,
        });
        return () => {
            cancelled = true;
        };
    }, [projects, dispatch, setError, projectApi]);
};

// devices and apps belong to the project; reload both when it changes
export const useProjectDataLoad = (
    projectIdentifier: string,
    dispatch: HubDispatch,
    setError: (e: object) => void,
): void => {
    const appApi = useAppService();
    const devicesApi = useDeviceService();
    useEffect(() => {
        let cancelled = false;
        const isCancelled = (): boolean => cancelled;
        applyUnlessCancelled(appApi.list(projectIdentifier), isCancelled, {
            apply: payload => dispatch({type: HubStateActions.AppsUpdate, payload}),
            fail: setError,
        });
        applyUnlessCancelled(devicesApi.list(projectIdentifier), isCancelled, {
            apply: payload => dispatch({type: HubStateActions.UpdateDevices, payload}),
            fail: setError,
        });
        return () => {
            cancelled = true;
        };
    }, [projectIdentifier, dispatch, setError, appApi, devicesApi]);
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

// keeps the state of one project up to date: its nodes, devices and apps
const ProjectUpdates: React.FC<{ projectIdentifier: string }> = ({projectIdentifier}) => {
    const {dispatch} = useHubState()
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
    return null;
};

// HubStateUpdates sits above the project pages: it loads the project list,
// and then keeps the project of the URL up to date once it is known. It reads the project from
// the URL, because the project provider below waits for the list it loads.
const HubStateUpdates: React.FC<{ children: ReactNode }> = ({children}) => {
    const {state, dispatch} = useHubState()
    const {setError} = useError()
    const {project_identifier: projectIdentifier} = useParams();
    useProjectsLoad(state.projects, dispatch, setError);
    // like the project provider, wait for the project to be known
    const known = state.projects?.some(p => p.identifier === projectIdentifier);

    return <>
        {projectIdentifier && known && <ProjectUpdates projectIdentifier={projectIdentifier}/>}
        {children}
    </>
};

export default HubStateUpdates;
