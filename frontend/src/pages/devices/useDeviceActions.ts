import {useCallback} from 'react';
import {useNavigate} from 'react-router-dom';
import {useAsyncAction} from '../../hooks/useAsyncAction';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {routes} from '../../lib/routes';
import {useDeviceService} from '../../services/device.service';
import {useUnlockDevice} from './useUnlockDevice';

export interface DeviceActions {
    remove: () => Promise<void | undefined>,
    unlock: () => Promise<void | undefined>,
    busy: boolean,
    error: string | null,
}

// useDeviceActions are the things the device page does to its device:
// delete it (and leave to the device list) or unlock it
export function useDeviceActions(deviceId: number): DeviceActions {
    const service = useDeviceService();
    const navigate = useNavigate();
    const {projectIdentifier} = useProjectContext();
    const unlock = useUnlockDevice();
    const removeDevice = useCallback(async (): Promise<void> => {
        await service.remove(projectIdentifier, deviceId);
        navigate(routes.project(projectIdentifier).devices());
    }, [service, projectIdentifier, deviceId, navigate]);
    const remove = useAsyncAction(removeDevice);
    return {
        remove: remove.run,
        unlock: () => unlock.run(deviceId),
        busy: remove.busy || unlock.busy,
        error: remove.error ?? unlock.error,
    };
}
