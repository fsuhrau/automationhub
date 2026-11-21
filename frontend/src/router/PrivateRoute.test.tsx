import {render, screen} from '@testing-library/react';
import React from 'react';
import {MemoryRouter, Route, Routes, useLocation} from 'react-router-dom';
import {describe, expect, it} from 'vitest';
import PrivateRoute from './PrivateRoute';
import {AuthProvider} from '../hooks/AuthProvider';
import {fakeAuthClient, noAuth, signedIn, signedOut} from '../app/testing/fakeAuth';
import {AuthClient} from '../app/authClient';

const Login: React.FC = () => <p>{`login from=${(useLocation().state as { from: string }).from}`}</p>;

const mount = (client: AuthClient): ReturnType<typeof render> => render(
    <MemoryRouter initialEntries={['/secret?x=1']}>
        <AuthProvider client={client}>
            <Routes>
                <Route path="/login" element={<Login/>}/>
                <Route element={<PrivateRoute/>}><Route path="/secret" element={<p>secret page</p>}/></Route>
            </Routes>
        </AuthProvider>
    </MemoryRouter>);

describe('PrivateRoute', () => {
    it('shows a spinner while the session is checked', () => {
        mount(fakeAuthClient(signedIn, {session: () => new Promise(() => undefined)}));
        expect(screen.getByRole('progressbar', {hidden: true})).toBeInTheDocument();
    });

    it('sends anonymous visitors to the sign in, which returns to the page', async () => {
        mount(fakeAuthClient(signedOut));
        expect(await screen.findByText('login from=/secret?x=1')).toBeInTheDocument();
    });

    it.each([['signed in', signedIn], ['without authentication', noAuth]])('shows the page when %s', async (_name, session) => {
        mount(fakeAuthClient(session));
        expect(await screen.findByText('secret page')).toBeInTheDocument();
    });
});
