import {useCallback} from 'react';
import {useAsyncAction} from '../../hooks/useAsyncAction';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {DeviceTestRequest} from '../../types/device.requests';
import {useDeviceService} from '../../services/device.service';
import {useUnlockDevice} from './useUnlockDevice';

export interface DeviceListActions {
    unlock: (deviceId: number) => Promise<void | undefined>,
    runTest: (deviceId: number, request: DeviceTestRequest) => Promise<void | undefined>,
    error: string | null,
}

// useDeviceListActions are the things the device list does to a device:
// unlock it, or start a test on it
export function useDeviceListActions(): DeviceListActions {
    const service = useDeviceService();
    const {projectIdentifier} = useProjectContext();
    const unlock = useUnlockDevice();
    const start = useCallback(async (deviceId: number, request: DeviceTestRequest): Promise<void> => {
        await service.runTest(projectIdentifier, deviceId, request);
    }, [service, projectIdentifier]);
    const runTest = useAsyncAction(start);
    return {unlock: unlock.run, runTest: runTest.run, error: unlock.error ?? runTest.error};
}
