import {AuthClient, SessionResult} from '../app/authClient';
import {IUser} from '../types/user';

export const aUser: IUser = {
    id: 1, name: 'Ada Lovelace', email: 'ada@example.com', login: 'ada', company: '', role: 'admin', url: '', avatar: '',
};

// A fake AuthClient for tests: a session of the given kind, and sign in, sign up
// and logout that succeed unless overridden.
export const fakeAuthClient = (session: SessionResult, overrides: Partial<AuthClient> = {}): AuthClient => ({
    session: async () => session,
    login: async () => ({ok: true, user: aUser}),
    register: async () => ({ok: true, user: aUser}),
    logout: async () => true,
    ...overrides,
});

export const signedIn: SessionResult = {state: 'authenticated', user: aUser};
export const signedOut: SessionResult = {state: 'anonymous', signIn: {method: 'password'}};
export const noAuth: SessionResult = {state: 'disabled'};
