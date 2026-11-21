import {describe, expect, it} from 'vitest';
import {FakeApi} from './pageKit';

describe('FakeApi', () => {
    it('answers by method and url, and records the calls', async () => {
        const api = new FakeApi({'PUT /x': (data: unknown) => ({echo: data}), 'GET /y': 1});
        expect(await api.put('/x', {a: 1})).toEqual({echo: {a: 1}});
        expect(api.to('PUT', '/x')).toHaveLength(1);
        expect(api.to('GET', '/y')).toHaveLength(0);
    });

    it('rejects a call nobody expected', async () => {
        await expect(new FakeApi({}).get('/z')).rejects.toThrow('unexpected GET /z');
    });
});
