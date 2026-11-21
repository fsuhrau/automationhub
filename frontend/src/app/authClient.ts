import {IUser} from '../types/user';

export interface RegisterData {
    name: string,
    email: string,
    password: string,
}

export interface LoginData {
    email: string,
    password: string,
}

// how people sign in to this master: an email and password form, or a
// provider whose sign in starts at url
export interface SignInMethod {
    method: 'password' | 'github' | 'oauth2',
    url?: string,
}

// what the session lookup found out
export type SessionResult =
    | { state: 'authenticated', user: IUser }
    | { state: 'anonymous', signIn: SignInMethod | null }
    | { state: 'disabled' };

export type CredentialsResult = { ok: true, user: IUser } | { ok: false, error: string };

// What the auth provider needs from the master's /auth routes; tests inject a fake.
export interface AuthClient {
    session: () => Promise<SessionResult>,
    login: (data: LoginData) => Promise<CredentialsResult>,
    register: (data: RegisterData) => Promise<CredentialsResult>,
    // false when the master could not be reached
    logout: () => Promise<boolean>,
}

type FetchFn = typeof fetch;

const METHODS = ['password', 'github', 'oauth2'];

const isJson = (response: Response): boolean => (response.headers.get('content-type') ?? '').includes('application/json');

const toSignIn = (body: { method?: string, url?: string }): SignInMethod | null =>
    METHODS.includes(body.method ?? '') ? {method: body.method as SignInMethod['method'], url: body.url} : null;

const readSession = async (response: Response): Promise<SessionResult> => {
    if (response.status === 401) {
        const body = await response.json().catch(() => ({}));
        return {state: 'anonymous', signIn: toSignIn(body)};
    }
    // no /auth routes: the master runs without authentication
    if (!response.ok || !isJson(response)) return {state: 'disabled'};
    const res = await response.json();
    // a master without authentication answers {"status": "ok"}
    return res.user ? {state: 'authenticated', user: res.user} : {state: 'disabled'};
};

// what differs between signing in and registering
interface CredentialsSpec {
    path: string,
    notEnabled: string,
    failed: (status: number) => string,
    unreachable: string,
}

const LOGIN: CredentialsSpec = {
    path: '/auth/login',
    notEnabled: 'Sign in with email and password is not enabled on this master (check auth.password in its config).',
    failed: () => 'Invalid email or password',
    unreachable: 'Sign in failed: the server could not be reached.',
};

const REGISTER: CredentialsSpec = {
    path: '/auth/register',
    notEnabled: 'Registration is not available: the master is not configured for password authentication (auth.password in its config).',
    failed: status => `Registration failed (${status})`,
    unreachable: 'Registration failed: the server could not be reached.',
};

const postJson = (fetchFn: FetchFn, path: string, data: unknown): Promise<Response> => fetchFn(path, {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(data),
    credentials: 'include',
});

const submit = async (fetchFn: FetchFn, spec: CredentialsSpec, data: unknown): Promise<CredentialsResult> => {
    try {
        const response = await postJson(fetchFn, spec.path, data);
        if (!isJson(response)) return {ok: false, error: spec.notEnabled};
        const res = await response.json();
        return res.user ? {ok: true, user: res.user} : {ok: false, error: res.error ?? spec.failed(response.status)};
    } catch (err) {
        console.error(err);
        return {ok: false, error: spec.unreachable};
    }
};

export const createAuthClient = (fetchFn: FetchFn = (input, init) => fetch(input, init)): AuthClient => ({
    session: async () => {
        try {
            return await readSession(await fetchFn('/auth/session', {method: 'GET', credentials: 'include'}));
        } catch (err) {
            // network problem: don't lock the user out, API calls still get 401
            console.error(err);
            return {state: 'disabled'};
        }
    },
    login: data => submit(fetchFn, LOGIN, data),
    register: data => submit(fetchFn, REGISTER, data),
    logout: async () => {
        try {
            await fetchFn('/auth/logout', {method: 'POST', credentials: 'include'});
            return true;
        } catch (err) {
            console.error(err);
            return false;
        }
    },
});
