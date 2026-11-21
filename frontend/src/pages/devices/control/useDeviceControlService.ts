import {useMemo} from 'react';
import {useParams} from 'react-router-dom';
import {useApiClient} from '../../../api/ApiClientProvider';
import {useProjectContext} from '../../../hooks/ProjectProvider';
import {deviceControlService} from '../../../services/device.control.service';
import {DeviceControlApi, DeviceTarget} from '../../../types/device.control.api';

// the device of the route /project/:id/device/:deviceId/control
export const useDeviceTarget = (): DeviceTarget => {
    const {projectIdentifier} = useProjectContext();
    const {deviceId} = useParams();
    return useMemo(() => ({projectId: projectIdentifier, deviceId: Number(deviceId)}), [projectIdentifier, deviceId]);
};

export const useDeviceControlService = (target: DeviceTarget): DeviceControlApi => {
    const api = useApiClient();
    return useMemo(() => deviceControlService(api, target), [api, target]);
};
