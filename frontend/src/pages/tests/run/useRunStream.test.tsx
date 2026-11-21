import {describe, expect, it} from 'vitest';
import {renderHook, RenderHookResult} from '@testing-library/react';
import {TestResultState as S} from '../../../types/test.result.state.enum';
import {FakeSse} from '../../../testing/fakeSse';
import {logEntry, protocol, run} from './harness/fixtures';
import {wrapperOf} from '../../../testing/renderWithProviders';
import ITestRunData from '../../../types/test.run';
import {RunStream, useRunStream} from './useRunStream';

interface Setup {
    sse: FakeSse,
    result: RenderHookResult<RunStream, unknown>['result'],
}

const setup = (over: Partial<ITestRunData> = {}): Setup => {
    const sse = new FakeSse();
    const {result} = renderHook(() => useRunStream(run({protocols: [protocol({id: 1})], log: [logEntry({id: 1})], ...over})), {wrapper: wrapperOf({sse})});
    return {sse, result};
};

describe('useRunStream', () => {
    it('starts with what the run holds', () => {
        const {result} = setup();
        expect(result.current.protocols.map(p => p.id)).toEqual([1]);
        expect(result.current.log.map(l => l.id)).toEqual([1]);
        expect(result.current.finished).toBe(false);
    });

    it('starts empty for a run without protocols and log', () => {
        const {result} = setup({protocols: undefined, log: undefined});
        expect(result.current).toMatchObject({protocols: [], log: []});
    });

    it('merges protocol updates', () => {
        const {sse, result} = setup();
        sse.emit('test_run_3_protocol', protocol({id: 1, testResult: S.TestResultFailed}));
        sse.emit('test_run_3_protocol', protocol({id: 2}));
        expect(result.current.protocols.map(p => [p.id, p.testResult])).toEqual([[1, S.TestResultFailed], [2, S.TestResultSuccess]]);
    });

    it('appends new log entries once', () => {
        const {sse, result} = setup();
        sse.emit('test_run_3_log', {testRunId: 3, entry: logEntry({id: 2, log: 'next'})});
        sse.emit('test_run_3_log', {testRunId: 3, entry: logEntry({id: 2, log: 'next'})});
        expect(result.current.log.map(l => l.id)).toEqual([1, 2]);
    });

    it('knows when the run finished', () => {
        const {sse, result} = setup();
        sse.emit('test_run_3_finished', {});
        expect(result.current.finished).toBe(true);
    });
});
