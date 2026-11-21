import {useCallback} from 'react';
import {useNavigate} from 'react-router-dom';
import {AsyncAction, useAsyncAction} from '../../hooks/useAsyncAction';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useHubState} from '../../hooks/HubStateProvider';
import {HubStateActions} from '../../application/HubState';
import {routes} from '../../lib/routes';
import {useDeviceService} from '../../services/device.service';
import {applyEdit, EditForm, toRequest} from './deviceEditModel';

// useSaveDevice saves the form of a device, updates the device list and
// opens the device page; on a failure it stays and reports the error
export function useSaveDevice(deviceId: number): AsyncAction<[EditForm], void> {
    const service = useDeviceService();
    const navigate = useNavigate();
    const {projectIdentifier} = useProjectContext();
    const {state, dispatch} = useHubState();
    const save = useCallback(async (form: EditForm): Promise<void> => {
        const saved = await service.update(projectIdentifier, deviceId, toRequest(form));
        const listed = state.devices?.find(d => d.id === deviceId);
        if (listed && saved?.id) {
            dispatch({type: HubStateActions.UpdateDevice, payload: applyEdit(listed, saved)});
        }
        navigate(routes.project(projectIdentifier).device(deviceId).show());
    }, [service, projectIdentifier, deviceId, state.devices, dispatch, navigate]);
    return useAsyncAction(save);
}
