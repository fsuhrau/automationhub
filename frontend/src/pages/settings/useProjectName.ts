import {useState} from 'react';
import {HubStateActions} from '../../application/HubState';
import {useHubState} from '../../hooks/HubStateProvider';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useProjectService} from '../../project/project.service';
import {useReportedAction} from './useReportedAction';

export interface ProjectName {
    name: string,
    setName: (name: string) => void,
    changed: boolean,
    saving: boolean,
    save: () => void,
    undo: () => void,
}

// useProjectName edits the project's name and stores it when saved; while
// there is no draft the field shows the stored name
export function useProjectName(): ProjectName {
    const {dispatch} = useHubState();
    const {project, projectIdentifier} = useProjectContext();
    const service = useProjectService();
    const [draft, setDraft] = useState<string | null>(null);
    const name = draft ?? project.name;
    const trimmed = name.trim();
    const changed = trimmed !== project.name && trimmed !== '';
    const saveAction = useReportedAction(async (): Promise<void> => {
        await service.update(projectIdentifier, {...project, name: trimmed});
        dispatch({type: HubStateActions.ProjectAttributeUpdate, payload: {projectIdentifier, attribute: 'name', value: trimmed}});
        setDraft(null);
    });
    return {
        name, setName: setDraft, changed, saving: saveAction.busy,
        save: () => {
            if (changed) void saveAction.run();
        },
        undo: () => setDraft(null),
    };
}
