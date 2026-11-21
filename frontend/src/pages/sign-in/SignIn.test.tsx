import {act, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import {Route, Routes, useLocation} from 'react-router-dom';
import {describe, expect, it, vi} from 'vitest';
import SignIn from './SignIn';
import SignUp from '../sign-up/SignUp';
import {renderInApp} from '../../app/testing/renderInApp';
import {aUser, fakeAuthClient, noAuth, signedIn, signedOut} from '../../app/testing/fakeAuth';
import {AuthClient, SessionResult} from '../../app/authClient';

const Where: React.FC<{ name: string }> = ({name}) => <p>{`${name} at ${useLocation().pathname}`}</p>;

const mount = async (session: SessionResult, overrides: Partial<AuthClient> = {}, path: object | string = '/login'): Promise<AuthClient> => {
    const auth = fakeAuthClient(session, overrides);
    renderInApp(<Routes>
        <Route path="/login" element={<SignIn/>}/>
        <Route path="/signup" element={<SignUp/>}/>
        <Route path="/secret" element={<Where name="secret"/>}/>
        <Route path="/" element={<Where name="projects"/>}/>
    </Routes>, {path, auth});
    await act(async () => undefined);
    return auth;
};

const fill = async (email: string, password: string): Promise<void> => {
    if (email) await userEvent.type(screen.getByLabelText('Email'), email);
    if (password) await userEvent.type(screen.getByLabelText('Password'), password);
    await userEvent.click(screen.getByRole('button', {name: 'Sign in'}));
};

describe('SignIn session states', () => {
    it('shows progress while the session is checked', async () => {
        await mount(signedOut, {session: () => new Promise(() => undefined)});
        expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    it('returns signed in users to where they came from', async () => {
        await mount(signedIn, {}, {pathname: '/login', state: {from: '/secret'}});
        expect(screen.getByText('secret at /secret')).toBeInTheDocument();
    });

    it('returns signed in users to the project list by default', async () => {
        await mount(signedIn);
        expect(screen.getByText('projects at /')).toBeInTheDocument();
    });

    it('lets people continue when no sign in is needed', async () => {
        await mount(noAuth);
        expect(screen.getByRole('heading', {name: 'No sign in needed'})).toBeInTheDocument();
        await userEvent.click(screen.getByRole('link', {name: 'Continue'}));
        expect(screen.getByText('projects at /')).toBeInTheDocument();
    });
});

describe('SignIn with a provider', () => {
    it.each([
        ['github', 'Sign in with GitHub'],
        ['oauth2', 'Sign in with your company account'],
    ] as const)('links to the sign in of %s', async (method, label) => {
        await mount({state: 'anonymous', signIn: {method, url: '/auth/start'}});
        expect(screen.getByRole('link', {name: label})).toHaveAttribute('href', '/auth/start');
        expect(screen.queryByLabelText('Email')).toBeNull();
    });

    it('shows the form when the provider gave no address', async () => {
        await mount({state: 'anonymous', signIn: {method: 'github'}});
        expect(screen.getByLabelText('Email')).toBeInTheDocument();
    });
});

describe('SignIn with a password', () => {
    it('asks for the missing input and sends nothing', async () => {
        const login = vi.fn();
        await mount(signedOut, {login});
        expect(screen.queryByText('Enter your password.')).toBeNull();
        await fill('not an email', '');
        expect(screen.getByText('Enter the email address of your account.')).toBeInTheDocument();
        expect(screen.getByText('Enter your password.')).toBeInTheDocument();
        expect(login).not.toHaveBeenCalled();
    });

    it('signs in and returns to where the user came from', async () => {
        const login = vi.fn(async () => ({ok: true as const, user: aUser}));
        await mount(signedOut, {login}, {pathname: '/login', state: {from: '/secret'}});
        await fill(' ada@example.com ', 'secret');
        await waitFor(() => expect(screen.getByText('secret at /secret')).toBeInTheDocument());
        expect(login).toHaveBeenCalledWith({email: 'ada@example.com', password: 'secret'});
    });

    it('shows the error of the master and lets the user try again', async () => {
        await mount(signedOut, {login: async () => ({ok: false, error: 'Invalid email or password'})});
        await fill('ada@example.com', 'wrong');
        expect(await screen.findByText('Invalid email or password')).toBeInTheDocument();
        expect(screen.getByRole('button', {name: 'Sign in'})).toBeEnabled();
    });

});

describe('SignIn password field and sign up link', () => {
    it('can show the password', async () => {
        await mount(signedOut);
        const field = screen.getByLabelText('Password');
        expect(field).toHaveAttribute('type', 'password');
        await userEvent.click(screen.getByRole('button', {name: 'Show the password'}));
        expect(field).toHaveAttribute('type', 'text');
        await userEvent.click(screen.getByRole('button', {name: 'Hide the password'}));
        expect(field).toHaveAttribute('type', 'password');
    });

    it('leads on to the sign up, keeping where the user came from', async () => {
        await mount(signedOut, {}, {pathname: '/login', state: {from: '/secret'}});
        await userEvent.click(screen.getByRole('link', {name: 'Create one'}));
        expect(screen.getByRole('heading', {name: 'Create an account'})).toBeInTheDocument();
        await userEvent.type(screen.getByLabelText('Name'), 'Ada');
        await userEvent.type(screen.getByLabelText('Email'), 'ada@example.com');
        await userEvent.type(screen.getByLabelText('Password'), 'longenough');
        await userEvent.click(screen.getByRole('button', {name: 'Create account'}));
        await waitFor(() => expect(screen.getByText('secret at /secret')).toBeInTheDocument());
    });
});
