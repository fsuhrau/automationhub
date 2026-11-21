import {AxiosResponse, InternalAxiosRequestConfig} from 'axios';
import axiosInstance from '../../../../http-common';
import {ApiClient, createApiClient} from '../../../../api/ApiClient';

export interface FakeRequest {
    method: string,
    url: string,
    data: unknown,
    params: unknown,
    config: InternalAxiosRequestConfig,
}

// what a route answers with: the body, or a function that returns it (or
// throws an error such as failure(...) to reject)
export type Reply = ((request: FakeRequest) => unknown) | object | string | number | boolean | null;

export interface FakeApi {
    client: ApiClient,
    calls: FakeRequest[],
    // the requests of one route, e.g. calledWith('POST /p/app/1/test/3/run')
    calledWith: (route: string) => FakeRequest[],
    restore: () => void,
}

// failure is what axios rejects with for an error response
export const failure = (status: number, message: string): object =>
    ({response: {status, data: {message}}, message: `Request failed with status code ${status}`});

const parse = (data: unknown): unknown => typeof data === 'string' ? JSON.parse(data) : data;

// createFakeApi answers the routes ("GET /p/apps") of the shared axios
// instance, so both an injected ApiClient and the services that still use
// the instance directly get the same fake master
export const createFakeApi = (routes: Record<string, Reply>): FakeApi => {
    const calls: FakeRequest[] = [];
    const original = axiosInstance.defaults.adapter;
    axiosInstance.defaults.adapter = async (config: InternalAxiosRequestConfig): Promise<AxiosResponse> => {
        const [url, query] = (config.url ?? '').split('?');
        const params = config.params ?? (query ? Object.fromEntries(new URLSearchParams(query)) : undefined);
        const request = {method: (config.method ?? 'get').toUpperCase(), url, data: parse(config.data), params, config};
        calls.push(request);
        const reply = routes[`${request.method} ${request.url}`];
        if (reply === undefined) throw failure(404, `no fake for ${request.method} ${request.url}`);
        const data = typeof reply === 'function' ? await reply(request) : reply;
        return {data, status: 200, statusText: 'OK', headers: {}, config};
    };
    return {
        client: createApiClient(axiosInstance),
        calls,
        calledWith: route => calls.filter(c => `${c.method} ${c.url}` === route),
        restore: () => {
            axiosInstance.defaults.adapter = original;
        },
    };
};
