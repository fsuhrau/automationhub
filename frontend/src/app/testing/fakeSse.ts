import {act} from '@testing-library/react';
import {createElement, Fragment, ReactNode, useReducer} from 'react';

// A fake of react-hooks-sse for tests: `send(channel, data)` feeds the
// stateReducer that a component registered with useSSE, like the real hook
// does for a server event. Use it with
//   vi.mock('react-hooks-sse', async () => (await import('<path>/app/testing/fakeSse')).fakeSseModule);
const channels = new Map<string, (data: unknown) => void>();

interface SseOptions<S, A> {
    stateReducer: (state: S, action: { data: A }) => S,
}

export const sendSse = (channel: string, data: unknown = {}): void => {
    const sink = channels.get(channel);
    if (!sink) throw new Error(`no listener on ${channel}`);
    act(() => sink(data));
};

export const resetSse = (): void => channels.clear();

export const fakeSseModule = {
    useSSE: <S, A>(channel: string, initial: S, options: SseOptions<S, A>): S => {
        const [state, dispatch] = useReducer(options.stateReducer, initial);
        channels.set(channel, data => dispatch({data: data as A}));
        return state;
    },
    SSEProvider: ({children}: { children: ReactNode }) => createElement(Fragment, null, children),
};
