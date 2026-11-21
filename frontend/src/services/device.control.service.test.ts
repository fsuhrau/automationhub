import {describe, expect, it, Mock, vi} from 'vitest';
import {DeviceControlApi} from '../types/device.control.api';
import {ApiClient} from '../api/ApiClient';
import {deviceControlService, LEASE_HEADER, logStreamUrl, screenshotUrl, streamUrl} from './device.control.service';

interface FakeApi {
    api: ApiClient,
    calls: unknown[][],
    answer: Mock<() => Promise<unknown>>,
}

const fakeApi = (): FakeApi => {
    const calls: unknown[][] = [];
    const answer = vi.fn(async (): Promise<unknown> => ({}));
    const record = (method: string) => (...args: unknown[]) => {
        calls.push([method, ...args]);
        return answer();
    };
    const api = {get: record('get'), post: record('post'), put: record('put'), delete: record('delete')} as unknown as ApiClient;
    return {api, calls, answer};
};

const target = {projectId: 'my project', deviceId: 7};
const lease = {headers: {[LEASE_HEADER]: 'tok'}};
const base = '/my%20project/device/7';

describe('urls', () => {
    it('builds the stream urls below /api', () => {
        expect(streamUrl('p', 7, 3)).toBe('/api/p/device/7/stream?fps=3');
        expect(streamUrl('p', 7)).toBe('/api/p/device/7/stream?fps=2');
        expect(logStreamUrl('p', 7)).toBe('/api/p/device/7/logs/stream?lines=300');
        expect(screenshotUrl('p', 7)).toMatch(/^\/api\/p\/device\/7\/screenshot\?t=\d+$/);
    });
});

type Case = [string, (s: DeviceControlApi) => Promise<unknown>, unknown[]];

// every call of the service: what it asks of the hub
const CASES: Case[] = [
    ['device', s => s.device(), ['get', base]],
    ['apps', s => s.apps(), ['get', '/my project/apps']],
    ['bundles', s => s.bundles(3), ['get', '/my project/app/3/bundles']],
    ['lockStatus', s => s.lockStatus(), ['get', `${base}/lock`]],
    ['lock', s => s.lock(), ['post', `${base}/lock`]],
    ['release', s => s.release('tok'), ['delete', `${base}/lock`, lease]],
    ['tap', s => s.tap('tok', {x: 1, y: 2}), ['post', `${base}/tap`, {x: 1, y: 2}, lease]],
    ['swipe', s => s.swipe('tok', {from: {x: 1, y: 2}, to: {x: 3, y: 4}, durationMs: 200}),
        ['post', `${base}/swipe`, {fromX: 1, fromY: 2, toX: 3, toY: 4, durationMs: 200}, lease]],
    ['text', s => s.text('tok', 'hi'), ['post', `${base}/text`, {text: 'hi'}, lease]],
    ['key', s => s.key('tok', 'home'), ['post', `${base}/key`, {key: 'home'}, lease]],
    ['touchElement', s => s.touchElement('tok', {xpath: '//A'}), ['post', `${base}/element/touch`, {xpath: '//A'}, lease]],
    ['appAction', s => s.appAction('tok', 'stop', {binaryId: 5, clearData: true}),
        ['post', `${base}/app/stop`, {binaryId: 5, clearData: true}, lease]],
    ['appAction without clearing', s => s.appAction('tok', 'start', {binaryId: 5}),
        ['post', `${base}/app/start`, {binaryId: 5, clearData: false}, lease]],
];

describe('deviceControlService', () => {
    it.each(CASES)('%s calls the hub', async (_, call, expected) => {
        const {api, calls} = fakeApi();
        await call(deviceControlService(api, target));
        expect(calls).toEqual([expected]);
    });
});

describe('deviceControlService reads', () => {
    it('reads the hierarchy as text', async () => {
        const {api, calls} = fakeApi();
        await deviceControlService(api, target).hierarchy();
        const [, url, config] = calls[0] as [string, string, { responseType: string, transformResponse: (r: string) => string }];
        expect(url).toBe(`${base}/hierarchy`);
        expect(config.responseType).toBe('text');
        expect(config.transformResponse('<a/>')).toBe('<a/>');
    });

    it('reads the log lines', async () => {
        const {api, calls, answer} = fakeApi();
        answer.mockResolvedValue({lines: ['a', 'b']});
        expect(await deviceControlService(api, target).logs(300)).toEqual(['a', 'b']);
        expect(calls[0]).toEqual(['get', `${base}/logs`, {params: {lines: 300}}]);
    });

    it('resolves actions without a body', async () => {
        const {api} = fakeApi();
        await expect(deviceControlService(api, target).tap('tok', {x: 0, y: 0})).resolves.toBeUndefined();
    });
});
