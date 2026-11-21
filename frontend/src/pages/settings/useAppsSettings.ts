import {useMemo, useState} from 'react';
import {HubStateActions} from '../../application/HubState';
import {AsyncAction} from '../../hooks/useAsyncAction';
import {useHubState} from '../../hooks/HubStateProvider';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useAppsSettingsService} from '../../services/settings.service';
import {AppParameter, IAppData} from '../../types/app';
import {NewAppRequest} from '../../types/settings.requests';
import {ParameterForm, parameterRequest, sortApps} from './settingsModel';
import {useHubLoad} from './useHubLoad';
import {useReportedAction} from './useReportedAction';

export interface AppsSettings {
    apps: IAppData[],
    // the selected app, the first one until another is picked
    app: IAppData | null,
    select: (id: number) => void,
    rename: AsyncAction<[string], boolean>,
    removeParameter: AsyncAction<[AppParameter], boolean>,
    // the dialogs show a failure themselves, so these throw and resolve true
    createApp: (request: NewAppRequest) => Promise<boolean>,
    saveParameter: (parameter: AppParameter | null, form: ParameterForm) => Promise<boolean>,
}

// useAppsSettings loads the project's apps and changes the selected one: its
// name and parameters; a new app becomes the selected one
export function useAppsSettings(): AppsSettings {
    const {state, dispatch} = useHubState();
    const {projectIdentifier: pid} = useProjectContext();
    const service = useAppsSettingsService();
    const [selectedId, select] = useState<number | null>(null);
    useHubLoad(() => service.list(pid), payload => ({type: HubStateActions.AppsUpdate, payload}), [pid]);
    const apps = useMemo(() => sortApps(state.apps ?? []), [state.apps]);
    const app = apps.find(a => a.id === selectedId) ?? apps[0] ?? null;

    const rename = useReportedAction(async (name: string): Promise<boolean> => {
        if (!app) return false;
        await service.update({projectId: pid, appId: app.id}, {...app, name});
        dispatch({type: HubStateActions.AppAttributeUpdate, payload: {appId: app.id, attribute: 'name', value: name}});
        return true;
    });
    const removeParameter = useReportedAction(async (parameter: AppParameter): Promise<boolean> => {
        if (!app) return false;
        await service.removeParameter({projectId: pid, appId: app.id}, parameter.id);
        dispatch({type: HubStateActions.DeleteAppParameter, payload: {appId: app.id, id: parameter.id}});
        return true;
    });
    const createApp = async (request: NewAppRequest): Promise<boolean> => {
        const created = await service.create(pid, request);
        dispatch({type: HubStateActions.AppAdd, payload: created});
        select(created.id);
        return true;
    };
    const saveParameter = async (parameter: AppParameter | null, form: ParameterForm): Promise<boolean> => {
        if (!app) return false;
        const body = parameterRequest(form);
        if (parameter) {
            const updated = await service.updateParameter({projectId: pid, appId: app.id}, parameter.id, body);
            dispatch({type: HubStateActions.UpdateAppParameter, payload: {...updated, appId: app.id}});
        } else {
            const added = await service.addParameter({projectId: pid, appId: app.id}, body);
            dispatch({type: HubStateActions.AddAppParameter, payload: {...added, appId: app.id}});
        }
        return true;
    };
    return {apps, app, select, rename, removeParameter, createApp, saveParameter};
}
