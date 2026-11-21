import {describe, expect, it} from 'vitest';
import {fakeApi, FakeRequest, FakeSse} from './fakes';

describe('fakeApi', () => {
    it('answers by method and url and records the requests', async () => {
        const api = fakeApi({'GET /a': 1, 'POST /b': ({data}: FakeRequest) => ({echo: data}), 'PUT /c': 3, 'DELETE /d': 4});
        expect(await api.get('/a')).toBe(1);
        expect(await api.post('/b', {x: 1})).toEqual({echo: {x: 1}});
        expect(await api.put('/c', {})).toBe(3);
        expect(await api.delete('/d')).toBe(4);
        expect(api.requests.map(r => r.key)).toEqual(['GET /a', 'POST /b', 'PUT /c', 'DELETE /d']);
    });

    it('rejects an Error answer and an unknown request', async () => {
        const api = fakeApi({'GET /a': new Error('boom')});
        await expect(api.get('/a')).rejects.toThrow('boom');
        await expect(api.get('/b')).rejects.toThrow('no answer for GET /b');
    });
});

describe('FakeSse', () => {
    it('sends events as JSON to the listeners of a name until they are removed', () => {
        const sse = new FakeSse();
        const source = sse.source();
        const seen: string[] = [];
        const listener = (e: Event): void => {
            seen.push((e as MessageEvent<string>).data);
        };
        source.addEventListener('news', listener);
        sse.emit('news', {a: 1});
        source.removeEventListener('news', listener);
        sse.emit('news', {a: 2});
        source.close();
        expect(seen).toEqual(['{"a":1}']);
    });
});
