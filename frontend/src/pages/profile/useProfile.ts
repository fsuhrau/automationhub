import {useEffect, useState} from 'react';
import {Dayjs} from 'dayjs';
import {useError} from '../../ErrorProvider';
import {AsyncAction} from '../../hooks/useAsyncAction';
import {useLoader} from '../../hooks/useLoader';
import {errorMessage} from '../../lib/errorMessage';
import IProject from '../../project/project';
import {useProjectService} from '../../project/project.service';
import {useProfileService} from '../../services/profile.service';
import {ICreatedPersonalToken, IPersonalToken, IProfile} from '../../types/profile.token';
import {useReportedAction} from '../settings/useReportedAction';

export interface ProfileData {
    profile: IProfile | null,
    profileError: string | null,
    // null until loaded
    tokens: IPersonalToken[] | null,
    projects: IProject[],
    // the token created last, whose value is shown once
    created: ICreatedPersonalToken | null,
    dismissCreated: () => void,
    // resolves true when the token was created
    create: AsyncAction<[string, Dayjs | null], boolean>,
    // resolves true when the token was revoked
    revoke: AsyncAction<[IPersonalToken], boolean>,
}

// useProfile loads the signed-in person's profile, personal tokens and the
// projects (for the guide), and creates and revokes personal tokens
export function useProfile(authenticated: boolean): ProfileData {
    const service = useProfileService();
    const projectService = useProjectService();
    const {setError} = useError();
    const [created, setCreated] = useState<ICreatedPersonalToken | null>(null);
    const profile = useLoader(() => authenticated ? service.profile() : null, [authenticated]);
    const tokens = useLoader(() => profile.data ? service.tokens() : null, [profile.data]);
    const projects = useLoader(() => authenticated ? projectService.list() : null, [authenticated]);
    useEffect(() => {
        if (tokens.error) setError(errorMessage(tokens.error));
    }, [tokens.error, setError]);

    const create = useReportedAction(async (name: string, expiresAt: Dayjs | null): Promise<boolean> => {
        setCreated(await service.createToken(name, expiresAt));
        tokens.retry();
        return true;
    });
    const revoke = useReportedAction(async (token: IPersonalToken): Promise<boolean> => {
        await service.deleteToken(token.id);
        setCreated(current => current?.id === token.id ? null : current);
        tokens.retry();
        return true;
    });
    return {
        profile: profile.data, profileError: profile.error ? errorMessage(profile.error) : null,
        tokens: tokens.data, projects: projects.data ?? [], created, dismissCreated: () => setCreated(null), create, revoke,
    };
}
