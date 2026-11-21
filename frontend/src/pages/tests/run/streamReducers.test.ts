import {describe, expect, it} from 'vitest';
import {TestResultState as S} from '../../../types/test.result.state.enum';
import {at, entry, protocol} from './harness/fixtures';
import {addEntry, addKey, lastEndedId, mergeOwnProtocol} from './streamReducers';

describe('addEntry', () => {
    it('appends a new entry', () => {
        const next = addEntry([entry({id: 1})], {entry: entry({id: 2})});
        expect(next.map(e => e.id)).toEqual([1, 2]);
    });

    it('ignores a known id and an event without entry', () => {
        const state = [entry({id: 1})];
        expect(addEntry(state, {entry: entry({id: 1})})).toBe(state);
        expect(addEntry(state, null)).toBe(state);
    });
});

describe('mergeOwnProtocol', () => {
    it('merges an update of the same protocol', () => {
        expect(mergeOwnProtocol(protocol({id: 1}), protocol({id: 1, testResult: S.TestResultFailed})).testResult).toBe(S.TestResultFailed);
    });

    it('keeps the state for another protocol or no update', () => {
        const state = protocol({id: 1});
        expect(mergeOwnProtocol(state, protocol({id: 2}))).toBe(state);
        expect(mergeOwnProtocol(state, null)).toBe(state);
    });
});

describe('lastEndedId', () => {
    it('is the id of a protocol that ended', () => {
        expect(lastEndedId(null, protocol({id: 4, endedAt: at(1)}))).toBe(4);
    });

    it('stays for updates of running protocols', () => {
        expect(lastEndedId(3, protocol({id: 4, endedAt: undefined}))).toBe(3);
        expect(lastEndedId(3, null)).toBe(3);
    });
});

describe('addKey', () => {
    it('merges the key of the event', () => {
        const key = {testProtocolId: 1, key: 'a', value: '1'};
        expect(addKey([], {testRunId: 3, testProtocolId: 1, key})).toEqual([key]);
        expect(addKey([key], null)).toEqual([key]);
    });
});
