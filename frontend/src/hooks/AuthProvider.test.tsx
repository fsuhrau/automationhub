import {act, render, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import {MemoryRouter, useLocation} from 'react-router-dom';
import {describe, expect, it, vi} from 'vitest';
import {AuthProvider, useAuth} from './AuthProvider';
import {AuthClient, SessionResult} from '../app/authClient';
import {aUser, fakeAuthClient, noAuth, signedIn, signedOut} from '../testing/fakeAuth';

const Probe: React.FC = () => {
    const auth = useAuth();
    const location = useLocation();
    return <>
        <p>{`state=${auth.authState} user=${auth.user?.name ?? '-'} method=${auth.signIn?.method ?? '-'} at=${location.pathname}`}</p>
        <button onClick={() => auth.loginAction({email: 'a@b.c', password: 'x'}, m => screen.getByTestId('err').append(m), '/after')}>login</button>
        <button onClick={() => auth.loginAction({email: 'a@b.c', password: 'x'}, () => undefined)}>login-default</button>
        <button onClick={() => auth.registerAction({name: 'n', email: 'a@b.c', password: 'x'}, m => screen.getByTestId('err').append(m))}>register</button>
        <button onClick={() => auth.logOut()}>logout</button>
        <span data-testid="err"/>
    </>;
};

const mount = async (client: AuthClient): Promise<void> => {
    render(<MemoryRouter initialEntries={['/start']}><AuthProvider client={client}><Probe/></AuthProvider></MemoryRouter>);
    await act(async () => undefined);
};

const status = (): string => screen.getByText(/^state=/).textContent ?? '';

describe('AuthProvider session', () => {
    it.each<[string, SessionResult, string]>([
        ['an existing session', signedIn, 'state=authenticated user=Ada Lovelace method=- at=/start'],
        ['nobody signed in', signedOut, 'state=anonymous user=- method=password at=/start'],
        ['a master without authentication', noAuth, 'state=disabled user=- method=- at=/start'],
    ])('looks up %s', async (_name, session, expected) => {
        await mount(fakeAuthClient(session));
        expect(status()).toBe(expected);
    });

    it('is checking until the lookup answers', () => {
        const client = fakeAuthClient(signedIn, {session: () => new Promise(() => undefined)});
        render(<MemoryRouter><AuthProvider client={client}><Probe/></AuthProvider></MemoryRouter>);
        expect(status()).toContain('state=checking');
    });

    it('ignores an answer that arrives after unmount', async () => {
        let answer: (r: SessionResult) => void = () => undefined;
        const client = fakeAuthClient(signedIn, {session: () => new Promise(resolve => {
            answer = resolve;
        })});
        const {unmount} = render(<MemoryRouter><AuthProvider client={client}><Probe/></AuthProvider></MemoryRouter>);
        unmount();
        await act(async () => answer(signedIn));
        expect(screen.queryByText(/^state=/)).toBeNull();
    });

    it('works with the real client by default', async () => {
        vi.stubGlobal('fetch', vi.fn(async () => new Response('{}', {status: 401})));
        render(<MemoryRouter><AuthProvider><Probe/></AuthProvider></MemoryRouter>);
        await waitFor(() => expect(status()).toContain('state=anonymous'));
        vi.unstubAllGlobals();
    });
});

describe('AuthProvider sign in and register', () => {
    it('signs in, then opens the page that asked for it', async () => {
        await mount(fakeAuthClient(signedOut));
        await userEvent.click(screen.getByText('login'));
        await waitFor(() => expect(status()).toBe('state=authenticated user=Ada Lovelace method=- at=/after'));
    });

    it('opens the project list after signing in without a target', async () => {
        await mount(fakeAuthClient(signedOut));
        await userEvent.click(screen.getByText('login-default'));
        await waitFor(() => expect(status()).toMatch(/at=\/$/));
    });

    it('registers and opens the project list', async () => {
        await mount(fakeAuthClient(signedOut));
        await userEvent.click(screen.getByText('register'));
        await waitFor(() => expect(status()).toBe('state=authenticated user=Ada Lovelace method=- at=/'));
    });

});

describe('AuthProvider failures and logout', () => {
    it('reports a failed sign in and stays', async () => {
        await mount(fakeAuthClient(signedOut, {login: async () => ({ok: false, error: 'Invalid email or password'})}));
        await userEvent.click(screen.getByText('login'));
        await waitFor(() => expect(screen.getByTestId('err')).toHaveTextContent('Invalid email or password'));
        expect(status()).toContain('state=anonymous');
    });

    it('logs out and opens the sign in', async () => {
        let loggedOut = false;
        await mount(fakeAuthClient(signedIn, {
            session: async () => loggedOut ? signedOut : signedIn,
            logout: async () => {
                loggedOut = true;
                return true;
            },
        }));
        await userEvent.click(screen.getByText('logout'));
        await waitFor(() => expect(status()).toBe('state=anonymous user=- method=password at=/login'));
    });

    it('opens the sign in after a logout', async () => {
        await mount(fakeAuthClient(signedIn));
        await userEvent.click(screen.getByText('logout'));
        await waitFor(() => expect(status()).toContain('at=/login'));
    });

    it('stays signed in when the logout did not reach the master', async () => {
        await mount(fakeAuthClient(signedIn, {logout: async () => false}));
        await userEvent.click(screen.getByText('logout'));
        expect(status()).toBe(`state=authenticated user=${aUser.name} method=- at=/start`);
    });
});

describe('useAuth without a provider', () => {
    it('is checking, with actions that do nothing', async () => {
        const Bare: React.FC = () => {
            const auth = useAuth();
            React.useEffect(() => {
                void auth.loginAction({email: '', password: ''}, () => undefined);
                void auth.registerAction({name: '', email: '', password: ''}, () => undefined);
                void auth.logOut();
            }, [auth]);
            return <p>{auth.authState}</p>;
        };
        render(<Bare/>);
        expect(screen.getByText('checking')).toBeInTheDocument();
    });
});
