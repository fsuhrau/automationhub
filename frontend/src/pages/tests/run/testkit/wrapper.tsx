import React from 'react';
import {MemoryRouter} from 'react-router-dom';
import {SSEProvider} from 'react-hooks-sse';
import {ApiClientProvider} from '../../../../api/ApiClientProvider';
import {ErrorProvider} from '../../../../ErrorProvider';
import {fakeApi, FakeApi, FakeSse} from './fakes';

interface Fakes {
    api?: FakeApi,
    sse?: FakeSse,
    url?: string,
}

// wrapperOf is the wrapper of renderHook for hooks that need the router, the
// API client, the event stream and the error banner
export const wrapperOf = ({api = fakeApi({}), sse = new FakeSse(), url = '/'}: Fakes = {}): React.FC<{ children: React.ReactNode }> => {
    const Wrapper: React.FC<{ children: React.ReactNode }> = ({children}) =>
        <MemoryRouter initialEntries={[url]}><ErrorProvider><ApiClientProvider client={api}>
            <SSEProvider source={sse.source}>{children}</SSEProvider>
        </ApiClientProvider></ErrorProvider></MemoryRouter>;
    return Wrapper;
};
