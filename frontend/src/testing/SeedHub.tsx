import React, {ReactNode, useEffect} from 'react';
import {HubState, HubStateAction, HubStateActions} from '../application/HubState';
import {useHubState} from '../hooks/HubStateProvider';
import IProject from '../project/project';
import {IAppData} from '../types/app';
import IDeviceData from '../types/device';
import {INodeData} from '../types/node';

// the lists a test puts into the hub state, as the loaders of the app would
export interface HubSeed {
    projects?: IProject[],
    apps?: IAppData[],
    devices?: IDeviceData[],
    nodes?: INodeData[],
}

const actionsOf = (seed: HubSeed): HubStateAction[] => {
    const actions: HubStateAction[] = [];
    if (seed.projects) actions.push({type: HubStateActions.ProjectsUpdate, payload: seed.projects});
    if (seed.apps) actions.push({type: HubStateActions.AppsUpdate, payload: seed.apps});
    if (seed.devices) actions.push({type: HubStateActions.UpdateDevices, payload: seed.devices});
    if (seed.nodes) actions.push({type: HubStateActions.NodesUpdate, payload: seed.nodes});
    return actions;
};

// every list that was given is in the state
const seeded = (seed: HubSeed, state: HubState): boolean =>
    (!seed.projects || !!state.projects) && (!seed.apps || !!state.apps) && (!seed.devices || !!state.devices) && (!seed.nodes || !!state.nodes);

// SeedHub puts the given lists into the hub state once mounted and shows its
// children when they are all in place. What is not given stays unloaded (null).
export const SeedHub: React.FC<HubSeed & { children: ReactNode }> = ({children, ...seed}) => {
    const {state, dispatch} = useHubState();
    const {projects, apps, devices, nodes} = seed;
    useEffect(() => {
        actionsOf({projects, apps, devices, nodes}).forEach(dispatch);
    }, [projects, apps, devices, nodes, dispatch]);
    return seeded(seed, state) ? <>{children}</> : null;
};
