import {render, screen, waitFor} from '@testing-library/react';
import {describe, expect, it, vi} from 'vitest';
import React from 'react';
import {InternalAxiosRequestConfig} from 'axios';
import axiosInstance from '../http-common';
import {ApiClient} from './ApiClient';
import {ApiClientProvider, useApiClient} from './ApiClientProvider';

const fakeClient = (): ApiClient => ({
    get: vi.fn().mockResolvedValue('from fake'),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
});

const Consumer: React.FC = () => {
    const client = useApiClient();
    const [text, setText] = React.useState('loading');
    React.useEffect(() => {
        client.get<string>('/x').then(setText).catch(() => setText('failed'));
    }, [client]);
    return <p>{text}</p>;
};

describe('useApiClient', () => {
    it('returns the injected client inside the provider', async () => {
        const client = fakeClient();
        render(<ApiClientProvider client={client}><Consumer/></ApiClientProvider>);
        await waitFor(() => expect(screen.getByText('from fake')).toBeInTheDocument());
        expect(client.get).toHaveBeenCalledWith('/x');
    });

    it('defaults to the client on the shared axios instance', async () => {
        const adapter = vi.fn(async (config: InternalAxiosRequestConfig) =>
            ({data: 'from shared', status: 200, statusText: '', headers: {}, config}));
        const original = axiosInstance.defaults.adapter;
        axiosInstance.defaults.adapter = adapter;
        render(<Consumer/>);
        await waitFor(() => expect(screen.getByText('from shared')).toBeInTheDocument());
        expect(adapter.mock.calls[0][0].url).toBe('/x');
        axiosInstance.defaults.adapter = original;
    });
});
