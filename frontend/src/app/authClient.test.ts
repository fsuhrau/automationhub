import {afterEach, describe, expect, it, vi} from 'vitest';
import {createAuthClient} from './authClient';
import {aUser} from './testing/fakeAuth';

const json = (body: unknown, init: ResponseInit = {}): Response =>
    new Response(JSON.stringify(body), {...init, headers: {'content-type': 'application/json'}});
const html = (init: ResponseInit = {}): Response => new Response('<html/>', {...init, headers: {'content-type': 'text/html'}});

const clientAnswering = (response: () => Response | Promise<Response>): ReturnType<typeof createAuthClient> =>
    createAuthClient(vi.fn(async () => response()));

afterEach(() => vi.restoreAllMocks());

describe('AuthClient.session', () => {
    it('finds the signed in user', async () => {
        expect(await clientAnswering(() => json({user: aUser})).session()).toEqual({state: 'authenticated', user: aUser});
    });

    it.each([
        ['a master without authentication answering status ok', () => json({status: 'ok'})],
        ['a master without /auth routes', () => html()],
        ['an error answer', () => json({}, {status: 500})],
    ])('is disabled for %s', async (_name, response) => {
        expect(await clientAnswering(response).session()).toEqual({state: 'disabled'});
    });

    it('is anonymous with the sign in method of a 401', async () => {
        const client = clientAnswering(() => json({method: 'github', url: '/auth/github'}, {status: 401}));
        expect(await client.session()).toEqual({state: 'anonymous', signIn: {method: 'github', url: '/auth/github'}});
    });

    it('is anonymous without a method when the 401 does not say', async () => {
        expect(await clientAnswering(() => json({method: 'other'}, {status: 401})).session()).toEqual({state: 'anonymous', signIn: null});
        expect(await clientAnswering(() => html({status: 401})).session()).toEqual({state: 'anonymous', signIn: null});
    });

    it('does not lock people out when the master cannot be reached', async () => {
        vi.spyOn(console, 'error').mockImplementation(() => undefined);
        const client = clientAnswering(() => Promise.reject(new Error('offline')));
        expect(await client.session()).toEqual({state: 'disabled'});
    });
});

describe.each([['login', 'Invalid email or password', 'not enabled'], ['register', 'Registration failed (400)', 'not available']] as const)(
    'AuthClient.%s', (action, failure, notEnabled) => {
        const send = (client: ReturnType<typeof createAuthClient>): Promise<unknown> => client[action]({name: 'n', email: 'a@b.c', password: 'secret123'});

        it('returns the user', async () => {
            expect(await send(clientAnswering(() => json({user: aUser})))).toEqual({ok: true, user: aUser});
        });

        it('posts the data as json to the auth route', async () => {
            const fetchFn = vi.fn(async () => json({user: aUser}));
            await send(createAuthClient(fetchFn));
            const [path, init] = fetchFn.mock.calls[0] as unknown as [string, RequestInit];
            expect(path).toBe(`/auth/${action === 'login' ? 'login' : 'register'}`);
            expect(init).toMatchObject({method: 'POST', credentials: 'include'});
            expect(JSON.parse(init.body as string)).toMatchObject({email: 'a@b.c'});
        });

        it('returns the error of the master', async () => {
            expect(await send(clientAnswering(() => json({error: 'taken'}, {status: 400})))).toEqual({ok: false, error: 'taken'});
        });

        it('has its own text when the master gives none', async () => {
            expect(await send(clientAnswering(() => json({}, {status: 400})))).toEqual({ok: false, error: failure});
        });

        it('says when password authentication is not enabled', async () => {
            const result = await send(clientAnswering(() => html()));
            expect(result).toMatchObject({ok: false, error: expect.stringContaining(notEnabled)});
        });

        it('says when the server cannot be reached', async () => {
            vi.spyOn(console, 'error').mockImplementation(() => undefined);
            const result = await send(clientAnswering(() => Promise.reject(new Error('offline'))));
            expect(result).toMatchObject({ok: false, error: expect.stringContaining('could not be reached')});
        });
    });

describe('AuthClient.logout', () => {
    it('is done when the master answers', async () => {
        const fetchFn = vi.fn(async () => json({}));
        expect(await createAuthClient(fetchFn).logout()).toBe(true);
        expect(fetchFn).toHaveBeenCalledWith('/auth/logout', {method: 'POST', credentials: 'include'});
    });

    it('is not done when the master cannot be reached', async () => {
        vi.spyOn(console, 'error').mockImplementation(() => undefined);
        expect(await clientAnswering(() => Promise.reject(new Error('offline'))).logout()).toBe(false);
    });
});

describe('createAuthClient without a fetch function', () => {
    it('uses the global fetch', async () => {
        const fetchFn = vi.fn(async () => json({status: 'ok'}));
        vi.stubGlobal('fetch', fetchFn);
        expect(await createAuthClient().session()).toEqual({state: 'disabled'});
        expect(fetchFn).toHaveBeenCalledWith('/auth/session', {method: 'GET', credentials: 'include'});
        vi.unstubAllGlobals();
    });
});
