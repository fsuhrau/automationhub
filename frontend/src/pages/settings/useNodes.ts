import {useState} from 'react';
import {HubStateActions} from '../../application/HubState';
import {AsyncAction} from '../../hooks/useAsyncAction';
import {useHubState} from '../../hooks/HubStateProvider';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useSettingsService} from '../../services/settings.service';
import {INodeData} from '../../types/node';
import {NodeWithToken} from '../../types/settings.requests';
import {useHubLoad} from './useHubLoad';
import {useReportedAction} from './useReportedAction';

export interface Nodes {
    nodes: INodeData[],
    // the node whose config (with its token) is shown, once
    setup: NodeWithToken | null,
    closeSetup: () => void,
    create: AsyncAction<[string], boolean>,
    reissue: AsyncAction<[INodeData], boolean>,
    remove: AsyncAction<[INodeData], boolean>,
}

// useNodes loads the nodes and adds, re-tokens and deletes them
export function useNodes(): Nodes {
    const {state, dispatch} = useHubState();
    const {projectIdentifier} = useProjectContext();
    const service = useSettingsService();
    const [setup, setSetup] = useState<NodeWithToken | null>(null);
    useHubLoad(() => service.nodes(projectIdentifier), payload => ({type: HubStateActions.NodesUpdate, payload}), [projectIdentifier]);

    const create = useReportedAction(async (name: string): Promise<boolean> => {
        const node = await service.createNode(projectIdentifier, {name});
        dispatch({type: HubStateActions.NodeAdd, payload: node});
        setSetup(node);
        return true;
    });
    const reissue = useReportedAction(async (node: INodeData): Promise<boolean> => {
        setSetup(await service.issueNodeToken(projectIdentifier, node.id));
        return true;
    });
    const remove = useReportedAction(async (node: INodeData): Promise<boolean> => {
        await service.deleteNode(projectIdentifier, node.id);
        dispatch({type: HubStateActions.NodeDelete, payload: node.id});
        setSetup(current => current?.id === node.id ? null : current);
        return true;
    });
    return {nodes: state.nodes ?? [], setup, closeSetup: () => setSetup(null), create, reissue, remove};
}
