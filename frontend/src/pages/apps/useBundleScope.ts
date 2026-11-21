import {useMemo} from 'react';
import {useApplicationContext} from '../../hooks/ApplicationProvider';
import {useHubState} from '../../hooks/HubStateProvider';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {AppRoutes, routes} from '../../lib/routes';
import {IAppData} from '../../types/app';
import {AppRef} from '../../types/app.ref';

export interface BundleScope {
    projectId: string,
    app: IAppData | undefined,
    // null until the app is known
    ref: AppRef | null,
    routes: AppRoutes,
}

// useBundleScope is the project and app the bundle pages work in
export const useBundleScope = (): BundleScope => {
    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const {state} = useHubState();
    const apps = state.apps;
    return useMemo(() => ({
        projectId: projectIdentifier,
        app: apps?.find(a => a.id === appId),
        ref: appId === null ? null : {projectId: projectIdentifier, appId},
        routes: routes.project(projectIdentifier).app(appId ?? 0),
    }), [projectIdentifier, appId, apps]);
};
