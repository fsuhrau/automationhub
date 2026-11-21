import {describe, expect, it} from 'vitest';
import {Call, failure, fakeApi} from './fakeApi';

describe('fakeApi', () => {
    it('answers by method and url and records every call', async () => {
        const api = fakeApi({'GET /a': 1, 'POST /b': ({data}: Call) => ({echo: data}), 'PUT /c': 3, 'DELETE /d': 4});
        expect(await api.get('/a', {params: {q: 1}})).toBe(1);
        expect(await api.post('/b', {x: 1})).toEqual({echo: {x: 1}});
        expect(await api.put('/c', {})).toBe(3);
        expect(await api.delete('/d')).toBe(4);
        expect(api.calls.map(c => c.key)).toEqual(['GET /a', 'POST /b', 'PUT /c', 'DELETE /d']);
        expect(api.calls[0].params).toEqual({q: 1});
    });

    it('finds the calls and bodies of one route', async () => {
        const api = fakeApi({'POST /b': null, 'GET /a': null});
        await api.post('/b', {n: 1});
        await api.get('/a');
        await api.post('/b', {n: 2});
        expect(api.called('POST /b')).toHaveLength(2);
        expect(api.bodies('POST /b')).toEqual([{n: 1}, {n: 2}]);
        expect(api.called('GET /none')).toEqual([]);
    });

    it('rejects an Error answer, a throwing function and a route nobody expects', async () => {
        const api = fakeApi({
            'GET /a': new Error('boom'),
            'GET /b': () => {
                throw failure('nope', 409);
            },
        });
        await expect(api.get('/a')).rejects.toThrow('boom');
        await expect(api.get('/b')).rejects.toMatchObject({response: {status: 409, data: {message: 'nope'}}});
        await expect(api.get('/c')).rejects.toThrow('no fake answer for GET /c');
    });

    it('waits for an answer that is a promise', async () => {
        const api = fakeApi({'GET /a': async () => 'late'});
        expect(await api.get('/a')).toBe('late');
    });

    it('starts without routes', async () => {
        await expect(fakeApi().get('/a')).rejects.toThrow('no fake answer');
    });
});

describe('failure', () => {
    it('is what axios rejects with, 500 unless said otherwise', () => {
        expect(failure('down')).toEqual({response: {status: 500, data: {message: 'down'}}, message: 'Request failed'});
    });
});
