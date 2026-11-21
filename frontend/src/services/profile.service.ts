import {useMemo} from 'react';
import {Dayjs} from 'dayjs';
import {ApiClient} from '../api/ApiClient';
import {useApiClient} from '../api/ApiClientProvider';
import {ICreatedPersonalToken, IPersonalToken, IProfile} from '../types/profile.token';

export interface ProfileService {
    profile: () => Promise<IProfile>,
    tokens: () => Promise<IPersonalToken[]>,
    // an expiry is the end of that day
    createToken: (name: string, expiresAt: Dayjs | null) => Promise<ICreatedPersonalToken>,
    deleteToken: (id: number) => Promise<void>,
}

export const profileService = (api: ApiClient): ProfileService => ({
    profile: () => api.get('/me'),
    tokens: () => api.get('/me/tokens'),
    createToken: (name, expiresAt) =>
        api.post('/me/tokens', {name, expiresAt: expiresAt ? expiresAt.endOf('day').toISOString() : null}),
    deleteToken: async (id) => {
        await api.delete(`/me/tokens/${id}`);
    },
});

export const useProfileService = (): ProfileService => {
    const api = useApiClient();
    return useMemo(() => profileService(api), [api]);
};
