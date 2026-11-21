import {render, screen} from '@testing-library/react';
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import App from './App';

beforeEach(() => {
    window.history.pushState({}, '', '/login');
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({method: 'password'}), {status: 401})));
});
afterEach(() => vi.unstubAllGlobals());

describe('App', () => {
    it('shows the sign in at /login, themed, with the real providers and a session lookup', async () => {
        const {container} = render(<App/>);
        expect(await screen.findByLabelText('Email')).toBeInTheDocument();
        expect(screen.getByRole('heading', {name: 'Sign in'})).toBeInTheDocument();
        expect(container.querySelector('.App')).not.toBeNull();
        expect(fetch).toHaveBeenCalledWith('/auth/session', expect.objectContaining({method: 'GET'}));
    });

    it('sends anonymous visitors of a private page to the sign in', async () => {
        window.history.pushState({}, '', '/profile');
        render(<App/>);
        expect(await screen.findByLabelText('Email')).toBeInTheDocument();
        expect(window.location.pathname).toBe('/login');
    });
});
