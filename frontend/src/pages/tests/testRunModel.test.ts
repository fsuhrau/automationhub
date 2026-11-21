import {describe, expect, it} from 'vitest';
import IProtocolKeyData from '../../types/protocol.key';
import {TestResultState as S} from '../../types/test.result.state.enum';
import {protocol, at} from './run/testkit/fixtures';
import {
    between, buildRows, countResults, filterRows, formatDuration, isTopLevel, leaves, mergeKey, mergeProtocol, parseParameters,
    ResultCounts, runEnd, runStatus, splitTestName, testResult,
} from './testRunModel';

const key = (k: string, value: string): IProtocolKeyData => ({testProtocolId: 1, key: k, value});

describe('isTopLevel', () => {
    it.each([[null, true], [undefined, true], [0, true], [4, false]])('parent %s gives %s', (parent, top) => {
        expect(isTopLevel(protocol({parentTestProtocolId: parent as number | null}))).toBe(top);
    });
});

describe('mergeProtocol', () => {
    it('ignores a missing update or one without id', () => {
        const list = [protocol()];
        expect(mergeProtocol(list, null)).toBe(list);
        expect(mergeProtocol(list, protocol({id: undefined}))).toBe(list);
        expect(mergeProtocol(list, protocol({id: null}))).toBe(list);
    });

    it('appends an unknown protocol', () => {
        expect(mergeProtocol([protocol()], protocol({id: 2})).map(p => p.id)).toEqual([1, 2]);
    });

    it('keeps the device and keys the update leaves out', () => {
        const existing = protocol({keys: [key('a', '1')]});
        const merged = mergeProtocol([existing], protocol({device: null, keys: undefined, testResult: S.TestResultFailed}))[0];
        expect(merged.device).toBe(existing.device);
        expect(merged.keys).toEqual(existing.keys);
        expect(merged.testResult).toBe(S.TestResultFailed);
    });

    it('does not let a late open update reopen a finished protocol', () => {
        const merged = mergeProtocol([protocol()], protocol({endedAt: undefined, testResult: S.TestResultOpen}))[0];
        expect(merged.endedAt).toEqual(at(10));
        expect(merged.testResult).toBe(S.TestResultSuccess);
    });
});

describe('splitTestName', () => {
    it('splits at the last slash', () => {
        expect(splitTestName('A.B/C/test')).toEqual({name: 'test', suite: 'A.B/C'});
    });

    it('has no suite without slash', () => {
        expect(splitTestName('test')).toEqual({name: 'test', suite: ''});
    });
});

describe('testResult', () => {
    it('takes the runner result when it is worse', () => {
        expect(testResult(protocol(), protocol({testResult: S.TestResultFailed}))).toBe(S.TestResultFailed);
    });

    it('keeps its own result against an open runner', () => {
        expect(testResult(protocol(), protocol({testResult: undefined}))).toBe(S.TestResultSuccess);
    });

    it('is open without result', () => {
        expect(testResult(protocol({testResult: undefined}))).toBe(S.TestResultOpen);
    });
});

describe('buildRows', () => {
    const parent = protocol({id: 1, testName: 'Run', testResult: S.TestResultSuccess});

    it('nests children under their parent', () => {
        const rows = buildRows([parent, protocol({id: 2, parentTestProtocolId: 1, testName: 'A'}), protocol({id: 3, parentTestProtocolId: 1, testName: 'B'})]);
        expect(rows).toHaveLength(1);
        expect(rows[0].children.map(c => c.name)).toEqual(['A', 'B']);
    });

    it('folds a runner with one child of the same test into one row', () => {
        const rows = buildRows([parent, protocol({id: 2, parentTestProtocolId: 1, testName: 'Run', startedAt: at(1)})]);
        expect(rows).toHaveLength(1);
        expect(rows[0].key).toBe(2);
        expect(rows[0].runner?.id).toBe(1);
        expect(rows[0].startedAt).toEqual(parent.startedAt);
    });

    it('lists tests without children as rows', () => {
        expect(leaves(buildRows([parent]))).toHaveLength(1);
    });

    it('leaves are the children of rows', () => {
        const rows = buildRows([parent, protocol({id: 2, parentTestProtocolId: 1, testName: 'A'}), protocol({id: 3, parentTestProtocolId: 1, testName: 'B'})]);
        expect(leaves(rows).map(r => r.name)).toEqual(['A', 'B']);
    });
});

const rows = buildRows([
    protocol({id: 1, testName: 'S/alpha', testResult: S.TestResultFailed}),
    protocol({id: 2, testName: 'S/beta', testResult: S.TestResultSuccess}),
    protocol({id: 3, testName: 'T/gamma', testResult: S.TestResultUnstable}),
    protocol({id: 4, testName: 'T/delta', testResult: undefined, endedAt: undefined}),
]);

describe('countResults', () => {
    it('counts every result', () => {
        expect(countResults(rows)).toEqual({
            [S.TestResultOpen]: 1, [S.TestResultUnstable]: 1, [S.TestResultFailed]: 1, [S.TestResultSuccess]: 1,
        });
    });
});

describe('filterRows', () => {
    it('keeps everything without filter', () => {
        expect(filterRows(rows, 'all', ' ')).toHaveLength(4);
    });

    it('filters by result and by name or suite', () => {
        expect(filterRows(rows, S.TestResultFailed, '').map(r => r.name)).toEqual(['alpha']);
        expect(filterRows(rows, 'all', 'GAM').map(r => r.name)).toEqual(['gamma']);
        expect(filterRows(rows, 'all', 'S').map(r => r.name)).toEqual(['alpha', 'beta']);
    });

    it('keeps the matching children of a row with children', () => {
        const nested = buildRows([
            protocol({id: 1, testName: 'Run'}),
            protocol({id: 2, parentTestProtocolId: 1, testName: 'A'}),
            protocol({id: 3, parentTestProtocolId: 1, testName: 'B'}),
        ]);
        const found = filterRows(nested, 'all', 'b');
        expect(found[0].children.map(c => c.name)).toEqual(['B']);
        expect(filterRows(nested, 'all', 'zzz')).toEqual([]);
    });
});

describe('runStatus', () => {
    const counts = ([open, unstable, failed, success]: number[]): ResultCounts =>
        ({[S.TestResultOpen]: open, [S.TestResultUnstable]: unstable, [S.TestResultFailed]: failed, [S.TestResultSuccess]: success});

    it.each([
        {lifecycle: 'running', c: counts([0, 0, 0, 5]), errors: 0, status: 'running'},
        {lifecycle: 'finished', c: counts([0, 0, 0, 0]), errors: 2, status: 'not_run'},
        {lifecycle: 'finished', c: counts([0, 0, 0, 0]), errors: 0, status: 'empty'},
        {lifecycle: 'finished', c: counts([1, 1, 1, 1]), errors: 0, status: 'failed'},
        {lifecycle: 'finished', c: counts([1, 1, 0, 1]), errors: 0, status: 'unstable'},
        {lifecycle: 'finished', c: counts([1, 0, 0, 1]), errors: 0, status: 'interrupted'},
        {lifecycle: 'finished', c: counts([0, 0, 0, 1]), errors: 0, status: 'success'},
        {lifecycle: 'finished', c: {...counts([0, 0, 0, 0]), 9: 1} as ResultCounts, errors: 0, status: 'empty'},
    ] as const)('is $status for a $lifecycle run with $errors log errors', ({lifecycle, c, errors, status}) => {
        expect(runStatus(lifecycle, c, errors)).toBe(status);
    });
});

describe('runEnd', () => {
    it('is null while running', () => {
        expect(runEnd('running', [protocol()], [])).toBeNull();
    });

    it('is the latest end of a protocol or log line', () => {
        expect(runEnd('finished', [protocol({endedAt: at(10)}), protocol({endedAt: undefined})], [at(30)])).toEqual(at(30));
    });

    it('is null without any time', () => {
        expect(runEnd('finished', [], [])).toBeNull();
    });
});

describe('parseParameters', () => {
    it('splits pairs by semicolon or line', () => {
        expect(parseParameters('A=1;B=x=y\nC')).toEqual([{name: 'A', value: '1'}, {name: 'B', value: 'x=y'}, {name: 'C', value: ''}]);
    });

    it.each([[null], [undefined], ['']])('has no parameters for %s', (text) => {
        expect(parseParameters(text)).toEqual([]);
    });
});

describe('formatDuration', () => {
    it.each([[-5, '0s'], [7000, '7s'], [464000, '7m 44s'], [3_900_000, '1h 5m']])('formats %d ms as %s', (ms, text) => {
        expect(formatDuration(ms)).toBe(text);
    });
});

describe('between', () => {
    it('is the span in ms, null without end', () => {
        expect(between(at(0), at(2))).toBe(2000);
        expect(between(at(0), null)).toBeNull();
    });
});

describe('mergeKey', () => {
    it('ignores a missing key', () => {
        const list = [key('a', '1')];
        expect(mergeKey(list, null)).toBe(list);
        expect(mergeKey(list, key('', 'x'))).toBe(list);
    });

    it('keeps the list for the same value and replaces a changed one, sorted', () => {
        const list = [key('b', '1'), key('c', '1')];
        expect(mergeKey(list, key('b', '1'))).toBe(list);
        expect(mergeKey(list, key('a', '1')).map(k => k.key)).toEqual(['a', 'b', 'c']);
        expect(mergeKey(list, key('b', '2')).map(k => `${k.key}${k.value}`)).toEqual(['b2', 'c1']);
    });
});
