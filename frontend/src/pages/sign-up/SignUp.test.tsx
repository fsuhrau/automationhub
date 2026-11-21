import {act, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import {Route, Routes, useLocation} from 'react-router-dom';
import {describe, expect, it, vi} from 'vitest';
import SignUp from './SignUp';
import {renderWithProviders} from '../../testing/renderWithProviders';
import {aUser, fakeAuthClient, noAuth, signedIn, signedOut} from '../../testing/fakeAuth';
import {AuthClient, SessionResult} from '../../app/authClient';

const Where: React.FC<{ name: string }> = ({name}) => <p>{`${name} at ${useLocation().pathname}`}</p>;

const mount = async (session: SessionResult, overrides: Partial<AuthClient> = {}): Promise<void> => {
    renderWithProviders(<Routes>
        <Route path="/signup" element={<SignUp/>}/>
        <Route path="/login" element={<Where name="login"/>}/>
        <Route path="/" element={<Where name="projects"/>}/>
    </Routes>, {url: '/signup', auth: fakeAuthClient(session, overrides)});
    await act(async () => undefined);
};

const fill = async (name: string, email: string, password: string): Promise<void> => {
    await userEvent.type(screen.getByLabelText('Name'), name);
    await userEvent.type(screen.getByLabelText('Email'), email);
    await userEvent.type(screen.getByLabelText('Password'), password);
    await userEvent.click(screen.getByRole('button', {name: 'Create account'}));
};

describe('SignUp session states', () => {
    it('shows progress while the session is checked', async () => {
        await mount(signedOut, {session: () => new Promise(() => undefined)});
        expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    it('sends signed in users on', async () => {
        await mount(signedIn);
        expect(screen.getByText('projects at /')).toBeInTheDocument();
    });

    it('needs no account without authentication', async () => {
        await mount(noAuth);
        expect(screen.getByRole('heading', {name: 'No account needed'})).toBeInTheDocument();
    });

    it('sends people to the sign in where accounts come from a provider', async () => {
        await mount({state: 'anonymous', signIn: {method: 'github', url: '/x'}});
        expect(screen.getByText('login at /login')).toBeInTheDocument();
    });
});

describe('SignUp form', () => {
    it('asks for the missing input and sends nothing', async () => {
        const register = vi.fn();
        await mount(signedOut, {register});
        expect(screen.getByText('At least 8 characters.')).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Create account'}));
        expect(screen.getByText(/Enter your name/)).toBeInTheDocument();
        expect(screen.getByText(/Enter an email address/)).toBeInTheDocument();
        expect(register).not.toHaveBeenCalled();
    });

    it('creates the account and opens the project list', async () => {
        const register = vi.fn(async () => ({ok: true as const, user: aUser}));
        await mount(signedOut, {register});
        await fill(' Ada ', ' ada@example.com ', 'longenough');
        await waitFor(() => expect(screen.getByText('projects at /')).toBeInTheDocument());
        expect(register).toHaveBeenCalledWith({name: 'Ada', email: 'ada@example.com', password: 'longenough'});
    });

    it('shows the error of the master', async () => {
        await mount(signedOut, {register: async () => ({ok: false, error: 'Email taken'})});
        await fill('Ada', 'ada@example.com', 'longenough');
        expect(await screen.findByText('Email taken')).toBeInTheDocument();
        expect(screen.getByRole('button', {name: 'Create account'})).toBeEnabled();
    });

    it('leads back to the sign in', async () => {
        await mount(signedOut);
        await userEvent.click(screen.getByRole('link', {name: 'Sign in'}));
        expect(screen.getByText('login at /login')).toBeInTheDocument();
    });
});
