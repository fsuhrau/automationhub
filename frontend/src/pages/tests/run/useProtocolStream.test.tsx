import {describe, expect, it} from 'vitest';
import {renderHook, RenderHookResult} from '@testing-library/react';
import {TestResultState as S} from '../../../types/test.result.state.enum';
import {FakeSse} from '../../../testing/fakeSse';
import {at, entry, protocol} from './harness/fixtures';
import {wrapperOf} from '../../../testing/renderWithProviders';
import ITestProtocolData from '../../../types/test.protocol';
import {ProtocolStream, useProtocolStream} from './useProtocolStream';

interface Setup {
    sse: FakeSse,
    result: RenderHookResult<ProtocolStream, unknown>['result'],
}

const setup = (over: Partial<ITestProtocolData> = {}): Setup => {
    const sse = new FakeSse();
    const initial = protocol({id: 1, entries: [entry({id: 1})], keys: [{testProtocolId: 1, key: 'a', value: '1'}], ...over});
    const {result} = renderHook(() => useProtocolStream(3, initial), {wrapper: wrapperOf({sse})});
    return {sse, result};
};

describe('useProtocolStream', () => {
    it('starts with what the protocol holds', () => {
        const {result} = setup();
        expect(result.current.protocol.id).toBe(1);
        expect(result.current.entries.map(e => e.id)).toEqual([1]);
        expect(result.current.keys.map(k => k.key)).toEqual(['a']);
        expect(result.current.endedId).toBeNull();
    });

    it('starts empty for a protocol without entries and keys', () => {
        const {result} = setup({entries: undefined, keys: undefined});
        expect(result.current).toMatchObject({entries: [], keys: []});
    });

    it('follows updates of its own protocol only', () => {
        const {sse, result} = setup();
        sse.emit('test_run_3_protocol', protocol({id: 2, testResult: S.TestResultFailed}));
        expect(result.current.protocol.testResult).toBe(S.TestResultSuccess);
        sse.emit('test_run_3_protocol', protocol({id: 1, testResult: S.TestResultFailed}));
        expect(result.current.protocol.testResult).toBe(S.TestResultFailed);
    });

    it('appends new entries once', () => {
        const {sse, result} = setup();
        const payload = {testProtocolId: 1, entry: entry({id: 2})};
        sse.emit('test_protocol_1_log', payload);
        sse.emit('test_protocol_1_log', payload);
        expect(result.current.entries.map(e => e.id)).toEqual([1, 2]);
    });

    it('merges keys the app reports', () => {
        const {sse, result} = setup();
        sse.emit('test_protocol_1_keys', {testRunId: 3, testProtocolId: 1, key: {testProtocolId: 1, key: 'a', value: '2'}});
        expect(result.current.keys[0].value).toBe('2');
    });

    it('remembers the protocol that ended last', () => {
        const {sse, result} = setup();
        sse.emit('test_run_3_protocol', protocol({id: 9, endedAt: at(5)}));
        expect(result.current.endedId).toBe(9);
    });
});
