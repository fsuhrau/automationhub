import {useMemo} from 'react';
import {AppRoutes, routes} from '../../../lib/routes';
import {useApplicationContext} from '../../../hooks/ApplicationProvider';
import {useProjectContext} from '../../../hooks/ProjectProvider';

// useAppRoutes builds the links of the app picked in the menu
export const useAppRoutes = (): AppRoutes => {
    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    return useMemo(() => routes.project(projectIdentifier).app(String(appId)), [projectIdentifier, appId]);
};
