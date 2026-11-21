import {AxiosRequestConfig} from 'axios';
import {ApiClient} from '../api/ApiClient';

export type Method = 'GET' | 'POST' | 'PUT' | 'DELETE';

// one request the fake master received
export interface Call {
    method: Method,
    url: string,
    // "GET /p1/devices", the key of the reply
    key: string,
    data?: unknown,
    config?: AxiosRequestConfig,
    // the query parameters the request was sent with
    params?: unknown,
}

// what a route answers with: the body, an Error (rejects), or a function that
// gets the call and returns the body or throws
export type Reply = unknown | ((call: Call) => unknown);

export interface FakeApi extends ApiClient {
    // every request in the order it was made
    calls: Call[],
    // the requests of one route, e.g. called('POST /p1/app')
    called: (key: string) => Call[],
    // the bodies sent to one route
    bodies: (key: string) => unknown[],
}

// failure is what axios rejects with for an error response: the server's message
export const failure = (message: string, status = 500): unknown =>
    ({response: {status, data: {message}}, message: 'Request failed'});

const settle = async (reply: Reply, call: Call): Promise<unknown> => {
    const body = typeof reply === 'function' ? await (reply as (c: Call) => unknown)(call) : reply;
    if (body instanceof Error) throw body;
    return body;
};

// fakeApi is the ApiClient of the tests. It answers by "METHOD /url" and
// records every call; a request nobody expects rejects like a missing route.
export const fakeApi = (replies: Record<string, Reply> = {}): FakeApi => {
    const calls: Call[] = [];
    const request = async <T, >({method, url, data, config}: Pick<Call, 'method' | 'url' | 'data' | 'config'>): Promise<T> => {
        const call: Call = {method, url, key: `${method} ${url}`, data, config, params: config?.params};
        calls.push(call);
        if (!(call.key in replies)) throw new Error(`no fake answer for ${call.key}`);
        return await settle(replies[call.key], call) as T;
    };
    const called = (key: string): Call[] => calls.filter(c => c.key === key);
    return {
        calls,
        called,
        bodies: key => called(key).map(c => c.data),
        get: (url, config) => request({method: 'GET', url, config}),
        post: (url, data, config) => request({method: 'POST', url, data, config}),
        put: (url, data, config) => request({method: 'PUT', url, data, config}),
        delete: (url, config) => request({method: 'DELETE', url, config}),
    };
};
