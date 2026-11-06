import http from '../http-common';
import {Dayjs} from 'dayjs';

export interface IProfile {
    id: number,
    name: string,
    login: string,
    email: string,
}

export interface IPersonalToken {
    id: number,
    name: string,
    // last characters of the token, for recognizing it
    hint: string,
    createdAt: string,
    expiresAt: string | null,
    lastUsedAt: string | null,
}

// the value is only returned once, when the token is created
export interface ICreatedPersonalToken extends IPersonalToken {
    token: string,
}

export const getProfile = (): Promise<IProfile> =>
    http.get('/me').then(r => r.data);

export const getPersonalTokens = (): Promise<IPersonalToken[]> =>
    http.get('/me/tokens').then(r => r.data);

export const createPersonalToken = (name: string, expiresAt: Dayjs | null): Promise<ICreatedPersonalToken> =>
    http.post('/me/tokens', {name, expiresAt: expiresAt ? expiresAt.endOf('day').toISOString() : null}).then(r => r.data);

export const deletePersonalToken = (id: number): Promise<void> =>
    http.delete(`/me/tokens/${id}`).then(() => undefined);
