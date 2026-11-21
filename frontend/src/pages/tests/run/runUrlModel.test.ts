import {describe, expect, it} from 'vitest';
import {at} from './harness/fixtures';
import {replayQuery} from './runUrlModel';

describe('replayQuery', () => {
    it('has no query without time and device', () => {
        expect(replayQuery()).toBe('');
        expect(replayQuery(undefined, null)).toBe('');
    });

    it('carries the time in ms and the device', () => {
        expect(replayQuery(at(0), 7)).toBe(`?t=${at(0).valueOf()}&device=7`);
        expect(replayQuery(undefined, 7)).toBe('?device=7');
    });
});
