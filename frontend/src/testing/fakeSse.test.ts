import {afterEach, describe, expect, it, vi} from 'vitest';
import {FakeSse, fakeSources, stubEventSource} from './fakeSse';

afterEach(() => vi.unstubAllGlobals());

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
        sse.emit('news');
        source.removeEventListener('news', listener);
        sse.emit('news', {a: 2});
        source.close();
        expect(seen).toEqual(['{"a":1}', '{}']);
    });

    it('is what new EventSource returns once stubbed', () => {
        const sse = new FakeSse();
        stubEventSource(sse);
        const seen: unknown[] = [];
        new EventSource('/x').addEventListener('news', e => seen.push((e as MessageEvent).data));
        sse.emit('news', 1);
        expect(seen).toEqual(['1']);
    });
});

describe('fakeSources', () => {
    it('creates a source per url that delivers events and closes', () => {
        const {create, sources} = fakeSources();
        const source = create('/log');
        const seen: string[] = [];
        source.addEventListener('lines', e => seen.push((e as MessageEvent<string>).data));
        sources[0].emit('lines', ['a']);
        sources[0].emit('other');
        source.close();
        expect(sources[0].url).toBe('/log');
        expect(seen).toEqual(['["a"]']);
        expect(sources[0].closed).toBe(true);
    });
});
