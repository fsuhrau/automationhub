import {describe, expect, it} from 'vitest';
import IProtocolKeyData from '../../../types/protocol.key';
import {IBaseline} from '../../../types/protocol.baseline';
import {entry, performance, protocol} from './harness/fixtures';
import {LogEntry, summarize} from './logModel';
import {comparisonOf, mergedKeys, ownErrors, phasesOf, pickedBaseline, relatedKeys, tabFromQuery} from './protocolViewModel';

const key = (k: string, value: string): IProtocolKeyData => ({testProtocolId: 1, key: k, value});

describe('tabFromQuery', () => {
    it.each([
        ['log', 'log'], ['screenshots', 'screenshots'], ['performance', 'performance'], ['checkpoint', 'performance'], ['x', 'overview'], [null, 'overview'],
    ])(
        'maps %s to %s', (query, tab) => {
            expect(tabFromQuery(query)).toBe(tab);
        },
    );
});

describe('relatedKeys', () => {
    const parent = protocol({id: 1, keys: [key('player', '1'), key('level', '2')]});
    const child = protocol({id: 2, parentTestProtocolId: 1, keys: [key('player', '9')]});
    const other = protocol({id: 3, keys: [key('other', 'x')]});

    it('takes the keys of the parent', () => {
        expect(relatedKeys([parent, child, other], child).map(k => `${k.key}=${k.value}`)).toEqual(['level=2', 'player=1']);
    });

    it('takes the keys of the children', () => {
        expect(relatedKeys([parent, child, other], parent).map(k => `${k.key}=${k.value}`)).toEqual(['player=9']);
    });

    it('is empty without run protocols or keys', () => {
        expect(relatedKeys(null, parent)).toEqual([]);
        expect(relatedKeys([protocol({id: 5, parentTestProtocolId: 1})], protocol({id: 1}))).toEqual([]);
        expect(relatedKeys([other], protocol({id: null, parentTestProtocolId: null}))).toEqual([]);
    });
});

describe('mergedKeys', () => {
    it('lets the own value win', () => {
        expect(mergedKeys([key('a', 'own')], [key('a', 'related'), key('b', 'x')]).map(k => `${k.key}=${k.value}`)).toEqual(['a=own', 'b=x']);
    });
});

describe('phasesOf', () => {
    it('counts setup and teardown entries, leaving out empty phases', () => {
        const log: LogEntry[] = [{...entry(), phase: 'setup'}, {...entry({id: 2}), phase: 'setup'}, entry({id: 3})];
        expect(phasesOf(log)).toEqual([{phase: 'setup', count: 2}]);
    });
});

describe('ownErrors', () => {
    it('drops runner errors that repeat the test error', () => {
        const own = [entry({id: 1, level: 'error', message: 'Test X Failed'})];
        const all: LogEntry[] = [...own, {...entry({id: 2, level: 'error', message: 'Test X Failed'}), phase: 'teardown'},
            {...entry({id: 3, level: 'error', message: 'crash'}), phase: 'teardown'}];
        expect(ownErrors(summarize(all), summarize(own)).map(e => e.id)).toEqual([1, 3]);
    });
});

describe('comparisonOf and pickedBaseline', () => {
    const baseline = {testRunId: 4, performance: [performance()]} as IBaseline;
    const earlier = protocol({testRunId: 2, performance: [performance()]});

    it('compares with the picked run once loaded', () => {
        expect(comparisonOf('4', baseline, earlier)).toEqual({runId: 4, performance: baseline.performance, averages: baseline});
        expect(comparisonOf('4', null, earlier)).toBeNull();
        expect(comparisonOf('4', {testRunId: 4} as IBaseline, null)?.performance).toEqual([]);
    });

    it('compares with the latest earlier run without a pick', () => {
        expect(comparisonOf(null, null, earlier)).toEqual({runId: 2, performance: earlier.performance, averages: null});
        expect(comparisonOf(null, null, protocol({testRunId: 2, performance: undefined as never}))?.performance).toEqual([]);
        expect(comparisonOf(null, null, null)).toBeNull();
    });

    it('takes a loaded baseline as picked only for the asked run', () => {
        expect(pickedBaseline('4', baseline)).toBe(baseline);
        expect(pickedBaseline('5', baseline)).toBeNull();
        expect(pickedBaseline(null, baseline)).toBeNull();
        expect(pickedBaseline('4', null)).toBeNull();
    });
});
