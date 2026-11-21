import axios, {AxiosAdapter, AxiosError, AxiosInstance, InternalAxiosRequestConfig} from 'axios';
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import axiosInstance from '../http-common';
import {ApiClient, createApiClient} from './ApiClient';

type Respond = (config: InternalAxiosRequestConfig) => { status: number, data: unknown };

// an adapter answers in place of the network; it records the requests
const fakeAdapter = (respond: Respond, seen: InternalAxiosRequestConfig[] = []): AxiosAdapter => async (config) => {
    seen.push(config);
    const {status, data} = respond(config);
    const response = {data, status, statusText: '', headers: {}, config};
    if (status >= 400) {
        throw new AxiosError('failed', 'ERR_BAD_REQUEST', config, null, response);
    }
    return response;
};

const instanceAnswering = (respond: Respond, seen?: InternalAxiosRequestConfig[]): AxiosInstance =>
    axios.create({baseURL: '/api', adapter: fakeAdapter(respond, seen)});

describe('createApiClient', () => {
    const seen: InternalAxiosRequestConfig[] = [];
    let client: ApiClient;

    beforeEach(() => {
        seen.length = 0;
        client = createApiClient(instanceAnswering(c => ({status: 200, data: {method: c.method, url: c.url}}), seen));
    });

    it.each([
        ['get', () => client.get('/things')],
        ['post', () => client.post('/things', {a: 1})],
        ['put', () => client.put('/things', {a: 1})],
        ['delete', () => client.delete('/things')],
    ])('%s resolves with the response body', async (method, call) => {
        await expect(call()).resolves.toEqual({method, url: '/things'});
    });

    it.each([
        ['post', () => client.post('/things', {a: 1})],
        ['put', () => client.put('/things', {a: 1})],
    ])('%s sends the data as JSON', async (_method, call) => {
        await call();
        expect(JSON.parse(seen[0].data)).toEqual({a: 1});
    });

    it('passes the config on', async () => {
        await client.get('/things', {params: {page: 2}});
        expect(seen[0].params).toEqual({page: 2});
    });

    it.each([
        ['get', (c: ApiClient) => c.get('/x')],
        ['post', (c: ApiClient) => c.post('/x')],
        ['put', (c: ApiClient) => c.put('/x')],
        ['delete', (c: ApiClient) => c.delete('/x')],
    ])('%s rejects with the request error', async (_method, call) => {
        const failing = createApiClient(instanceAnswering(() => ({status: 500, data: {message: 'boom'}})));
        await expect(call(failing)).rejects.toMatchObject({response: {status: 500, data: {message: 'boom'}}});
    });
});

describe('the 401 redirect of the shared axios instance', () => {
    const originalAdapter = axiosInstance.defaults.adapter;
    const originalLocation = window.location;
    let location: { href: string };

    beforeEach(() => {
        location = {href: '/start'};
        Object.defineProperty(window, 'location', {value: location, configurable: true});
    });

    afterEach(() => {
        axiosInstance.defaults.adapter = originalAdapter;
        Object.defineProperty(window, 'location', {value: originalLocation, configurable: true});
        vi.restoreAllMocks();
    });

    const answerWith = (status: number, data: unknown): ApiClient => {
        axiosInstance.defaults.adapter = fakeAdapter(() => ({status, data}));
        return createApiClient(axiosInstance);
    };

    it('sends the browser to the url of a 401 and still rejects', async () => {
        const client = answerWith(401, {url: '/auth/login'});
        await expect(client.get('/me')).rejects.toMatchObject({response: {status: 401}});
        expect(location.href).toBe('/auth/login');
    });

    it('stays on the page for a 401 without url', async () => {
        const client = answerWith(401, {});
        await expect(client.get('/me')).rejects.toBeDefined();
        expect(location.href).toBe('/start');
    });

    it('stays on the page for other errors', async () => {
        const client = answerWith(500, {url: '/auth/login'});
        await expect(client.get('/me')).rejects.toBeDefined();
        expect(location.href).toBe('/start');
    });

    it('passes a successful response through', async () => {
        const client = answerWith(200, {ok: true});
        await expect(client.get('/me')).resolves.toEqual({ok: true});
    });
});
