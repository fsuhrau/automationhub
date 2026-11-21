import {describe, expect, it} from 'vitest';
import {TestResultState as S} from '../../types/test.result.state.enum';
import {at, protocol} from './run/testkit/fixtures';
import {earlierRuns, neighbours, runnerOf} from './protocolModel';
import {buildRows, leaves} from './testRunModel';

const tests = leaves(buildRows([
    protocol({id: 1, testName: 'A', testResult: S.TestResultSuccess}),
    protocol({id: 2, testName: 'B', testResult: S.TestResultFailed}),
    protocol({id: 3, testName: 'C', testResult: S.TestResultSuccess}),
    protocol({id: 4, testName: 'D', testResult: S.TestResultUnstable}),
]));

describe('neighbours', () => {
    it('places a protocol among the tests', () => {
        const place = neighbours(tests, 2);
        expect([place.position, place.total, place.prev?.name, place.next?.name, place.nextProblem?.name]).toEqual([2, 4, 'A', 'C', 'D']);
    });

    it('has no neighbours at the ends and none for an unknown protocol', () => {
        expect(neighbours(tests, 1).prev).toBeNull();
        expect(neighbours(tests, 4)).toMatchObject({next: null, nextProblem: null});
        expect(neighbours(tests, 99)).toEqual({position: 0, total: 4, prev: null, next: null, nextProblem: null});
    });

    it('finds a test by its runner', () => {
        const wrapped = leaves(buildRows([protocol({id: 1, testName: 'A'}), protocol({id: 2, parentTestProtocolId: 1, testName: 'A'})]));
        expect(neighbours(wrapped, 1).position).toBe(1);
    });
});

describe('runnerOf', () => {
    const wrapped = leaves(buildRows([protocol({id: 1, testName: 'A'}), protocol({id: 2, parentTestProtocolId: 1, testName: 'A'})]));

    it('is the runner that wrapped the test', () => {
        expect(runnerOf(wrapped, neighbours(wrapped, 2), 2)?.id).toBe(1);
    });

    it('is none for the runner itself, a test without runner, or an unknown protocol', () => {
        expect(runnerOf(wrapped, neighbours(wrapped, 1), 1)).toBeNull();
        expect(runnerOf(tests, neighbours(tests, 1), 1)).toBeNull();
        expect(runnerOf(tests, neighbours(tests, 99), 99)).toBeNull();
    });
});

describe('earlierRuns', () => {
    it('lists the history of other runs, newest first', () => {
        const history = [
            protocol({id: 1, testRunId: 3}),
            protocol({id: 8, testRunId: 3}),
            protocol({id: 5, testRunId: 1, startedAt: at(-200)}),
            protocol({id: 6, testRunId: 2, startedAt: at(-100)}),
        ];
        expect(earlierRuns(protocol({testProtocolHistory: history})).map(h => h.id)).toEqual([6, 5]);
    });

    it('is empty without history', () => {
        expect(earlierRuns(protocol({testProtocolHistory: null}))).toEqual([]);
    });
});
