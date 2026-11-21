import {act, render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {ErrorProvider, useError} from './ErrorProvider';

const Raiser: React.FC<{ error: string | object | null }> = ({error}) => {
    const {setError} = useError();
    return <button onClick={() => setError(error)}>raise</button>;
};

const raise = (error: string | object | null): void => {
    render(<ErrorProvider><Raiser error={error}/></ErrorProvider>);
    act(() => screen.getByText('raise').click());
};

beforeEach(() => vi.useFakeTimers());
afterEach(() => vi.useRealTimers());

describe('ErrorProvider', () => {
    it.each([
        ['text', 'plain text', 'plain text'],
        ['the message the server sent', {response: {data: {message: 'server says no'}}, message: 'Request failed'}, 'server says no'],
        ['the message of an error', new Error('boom'), 'boom'],
        ['a fallback for an object without a message', {}, 'An unknown error occurred'],
    ])('shows %s', (_name, error, text) => {
        raise(error);
        expect(screen.getByRole('alert')).toHaveTextContent(text);
    });

    it('shows nothing without an error', () => {
        raise(null);
        expect(screen.queryByRole('alert')).toBeNull();
    });

    it('hides the error after a while', () => {
        raise('x');
        act(() => {
            vi.advanceTimersByTime(8000);
        });
        expect(screen.queryByRole('alert')).toBeNull();
    });

    it('hides the error on close', () => {
        vi.useRealTimers();
        raise('x');
        return userEvent.click(screen.getByRole('button', {name: /close/i})).then(() => expect(screen.queryByRole('alert')).toBeNull());
    });
});

describe('useError', () => {
    it('needs an ErrorProvider', () => {
        vi.spyOn(console, 'error').mockImplementation(() => undefined);
        expect(() => render(<Raiser error="x"/>)).toThrow('useError must be used within an ErrorProvider');
    });
});
