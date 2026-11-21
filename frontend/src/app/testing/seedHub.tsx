import React, {ReactNode, useEffect} from 'react';
import IProject from '../../project/project';
import IDeviceData from '../../types/device';
import {IAppData} from '../../types/app';
import {HubStateActions} from '../../application/HubStateActions';
import {useHubState} from '../../hooks/HubStateProvider';
import {PlatformType} from '../../types/platform.type.enum';

export const aProject = (identifier: string, name = identifier): IProject =>
    ({id: 1, identifier, name, companyId: 1, accessTokens: [], apps: []});

export const anApp = (id: number, name = `app ${id}`): IAppData =>
    ({id, name, identifier: `app.${id}`, projectId: 1, platform: PlatformType.Android} as IAppData);

export const aDevice = (fields: Partial<IDeviceData>): IDeviceData => ({id: 1, name: 'phone', alias: '', ...fields} as IDeviceData);

interface SeedProps {
    projects?: IProject[],
    apps?: IAppData[],
    devices?: IDeviceData[],
    children: ReactNode,
}

// Puts the given lists into the hub state once mounted, like the loaders of the
// app do. What is not given stays unloaded (null).
export const SeedHub: React.FC<SeedProps> = ({projects, apps, devices, children}) => {
    const {dispatch} = useHubState();
    useEffect(() => {
        if (projects) dispatch({type: HubStateActions.ProjectsUpdate, payload: projects});
        if (apps) dispatch({type: HubStateActions.AppsUpdate, payload: apps});
        if (devices) dispatch({type: HubStateActions.UpdateDevices, payload: devices});
    }, [projects, apps, devices, dispatch]);
    return <>{children}</>;
};
