import {useCallback} from 'react';
import {AsyncAction, useAsyncAction} from '../../hooks/useAsyncAction';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useHubState} from '../../hooks/HubStateProvider';
import {HubStateActions} from '../../application/HubState';
import {useDeviceService} from '../../services/device.service';

// useUnlockDevice ends the test run or control session that holds a device
// and marks it free in the device list
export function useUnlockDevice(): AsyncAction<[number], void> {
    const service = useDeviceService();
    const {projectIdentifier} = useProjectContext();
    const {dispatch} = useHubState();
    const unlock = useCallback(async (deviceId: number): Promise<void> => {
        await service.unlock(projectIdentifier, deviceId);
        dispatch({type: HubStateActions.UnlockDevice, payload: {deviceId}});
    }, [service, projectIdentifier, dispatch]);
    return useAsyncAction(unlock);
}
