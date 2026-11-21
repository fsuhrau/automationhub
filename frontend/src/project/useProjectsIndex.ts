import {useMemo, useState} from 'react';
import {useNavigate} from 'react-router-dom';
import {HubStateActions} from '../application/HubState';
import {useHubState} from '../hooks/HubStateProvider';
import {useLoader} from '../hooks/useLoader';
import {errorMessage} from '../lib/errorMessage';
import {routes} from '../lib/routes';
import IProject from './project';
import {lastProject, visibleProjects} from './projectsModel';
import {useProjectService} from './project.service';

export interface ProjectsIndex {
    // null while the projects load
    projects: IProject[] | null,
    visible: IProject[],
    last: string | null,
    error: string | null,
    retry: () => void,
    search: string,
    setSearch: (search: string) => void,
    // creates the project and opens its settings to add an app; throws on failure
    create: (name: string) => Promise<boolean>,
}

// useProjectsIndex loads the projects into the hub state (when they are not
// there yet), filters them by the search, and creates a new one
export function useProjectsIndex(): ProjectsIndex {
    const {state, dispatch} = useHubState();
    const service = useProjectService();
    const navigate = useNavigate();
    const [search, setSearch] = useState('');
    const [last] = useState(lastProject);
    const {error, retry} = useLoader(() => state.projects === null
        ? service.list().then(payload => dispatch({type: HubStateActions.ProjectsUpdate, payload}))
        : null, [state.projects]);
    const visible = useMemo(() => visibleProjects(state.projects ?? [], last, search), [state.projects, last, search]);

    const create = async (name: string): Promise<boolean> => {
        const project = await service.create({name});
        dispatch({type: HubStateActions.ProjectsUpdate, payload: [...(state.projects ?? []), project]});
        // a new project needs an app first
        navigate(routes.project(project.identifier).settings('apps'));
        return true;
    };
    return {projects: state.projects, visible, last, error: error ? errorMessage(error) : null, retry, search, setSearch, create};
}
