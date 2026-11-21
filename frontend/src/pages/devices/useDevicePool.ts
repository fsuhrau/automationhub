import {useEffect} from 'react';
import IDeviceData from '../../types/device';
import {Loaded, useLoader} from '../../hooks/useLoader';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useHubState} from '../../hooks/HubStateProvider';
import {HubStateActions} from '../../application/HubState';
import {useDeviceService} from '../../services/device.service';

// useDevicePool loads the devices of the project into the hub state, where
// the live updates keep them current; the result carries error and retry
export function useDevicePool(): Loaded<IDeviceData[]> {
    const service = useDeviceService();
    const {projectIdentifier} = useProjectContext();
    const {dispatch} = useHubState();
    const loaded = useLoader(() => service.list(projectIdentifier), [service, projectIdentifier]);
    const {data} = loaded;
    useEffect(() => {
        if (data) dispatch({type: HubStateActions.UpdateDevices, payload: data});
    }, [data, dispatch]);
    return loaded;
}
