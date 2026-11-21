import {useState} from 'react';
import {Dayjs} from 'dayjs';
import {HubStateActions} from '../../application/HubState';
import {AsyncAction} from '../../hooks/useAsyncAction';
import {useHubState} from '../../hooks/HubStateProvider';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useSettingsService} from '../../services/settings.service';
import IAccessTokenData from '../../types/access.token';
import {useHubLoad} from './useHubLoad';
import {useReportedAction} from './useReportedAction';

export interface AccessTokens {
    tokens: IAccessTokenData[],
    // the token created last, whose value the page shows in a command
    created: IAccessTokenData | null,
    // resolves true when the token was created
    create: AsyncAction<[string, Dayjs | null], boolean>,
    // resolves true when the token was revoked
    revoke: AsyncAction<[IAccessTokenData], boolean>,
}

// useAccessTokens loads the project's access tokens and creates and revokes them
export function useAccessTokens(): AccessTokens {
    const {state, dispatch} = useHubState();
    const {projectIdentifier} = useProjectContext();
    const service = useSettingsService();
    const [created, setCreated] = useState<IAccessTokenData | null>(null);
    useHubLoad(() => service.accessTokens(projectIdentifier), payload => ({type: HubStateActions.AccessTokensUpdate, payload}), [projectIdentifier]);

    const create = useReportedAction(async (name: string, expiresAt: Dayjs | null): Promise<boolean> => {
        const token = await service.createAccessToken(projectIdentifier, {name, expiresAt});
        dispatch({type: HubStateActions.AccessTokenAdd, payload: token});
        setCreated(token);
        return true;
    });
    const revoke = useReportedAction(async (token: IAccessTokenData): Promise<boolean> => {
        await service.deleteAccessToken(projectIdentifier, token.id);
        dispatch({type: HubStateActions.AccessTokenDelete, payload: token.id});
        setCreated(current => current?.id === token.id ? null : current);
        return true;
    });
    return {tokens: state.accessTokens ?? [], created, create, revoke};
}
