import {describe, expect, it} from 'vitest';
import {TestExecutionType} from '../../types/test.execution.type.enum';
import {TestType} from '../../types/test.type.enum';
import {UnityTestCategory} from '../../types/unity.test.category.type.enum';
import {aSummary, aTest} from './definitions/testing/fixtures';
import {countHealth, describeConfig, filterTests, health, runStats, TestConfigLike} from './testsListModel';

const config = (over: Partial<TestConfigLike> = {}): TestConfigLike =>
    ({type: TestType.Unity, executionType: TestExecutionType.Concurrent, allDevices: true, unity: null, ...over});
const unity = (categoryType: UnityTestCategory, rest: object = {}): TestConfigLike['unity'] => ({testCategoryType: categoryType, ...rest});
const run = (status: Parameters<typeof aSummary>[0] = {}): ReturnType<typeof aSummary> => aSummary(status);

describe('health', () => {
    it.each([
        ['no runs', undefined, 'never'],
        ['running', [run({status: 'running'}), run({status: 'failed'})], 'running'],
        ['passed', [run({status: 'passed'})], 'passing'],
        ['failed', [run({status: 'failed'})], 'failing'],
        ['empty after passed', [run({status: 'empty'}), run({status: 'running'}), run({status: 'passed'})], 'passing'],
        ['interrupted after failed', [run({status: 'interrupted'}), run({status: 'failed'})], 'failing'],
        ['only interrupted', [run({status: 'interrupted'})], 'never'],
    ])('%s is %s', (_, runs, expected) => {
        expect(health(runs)).toBe(expected);
    });
});

describe('countHealth and filterTests', () => {
    const tests = [
        aTest({id: 1, name: 'Alpha', recentRuns: [run({status: 'passed'})]}),
        aTest({id: 2, name: 'Beta', recentRuns: [run({status: 'failed'})]}),
        aTest({id: 3, name: 'Gamma'}),
    ];

    it('counts the tests by health', () => {
        expect(countHealth(tests)).toEqual({all: 3, failing: 1, passing: 1, running: 0, never: 1});
    });

    it('filters by health and by text', () => {
        expect(filterTests(tests, 'failing', '').map(t => t.id)).toEqual([2]);
        expect(filterTests(tests, 'all', ' ALP ').map(t => t.id)).toEqual([1]);
        expect(filterTests(tests, 'never', 'alp')).toEqual([]);
    });
});

describe('describeConfig', () => {
    it.each([
        [config({unity: unity(UnityTestCategory.RunAllTests)}), 'Unity · all tests · on all devices · one device after the other'],
        [config({unity: unity(UnityTestCategory.RunAllOfCategory, {categories: 'fast'})}),
            'Unity · category fast · on all devices · one device after the other'],
        [config({unity: unity(UnityTestCategory.RunAllOfCategory)}), 'Unity · a category · on all devices · one device after the other'],
        [config({unity: unity(UnityTestCategory.RunSelectedTestsOnly, {testFunctions: [1]})}), 'Unity · 1 test · on all devices · one device after the other'],
        [config({unity: unity(UnityTestCategory.RunSelectedTestsOnly)}), 'Unity · 0 tests · on all devices · one device after the other'],
        [config({allDevices: false, devices: [1, 2], executionType: TestExecutionType.Simultaneously}), 'Unity · on 2 devices · devices at the same time'],
        [config({allDevices: false, devices: [1]}), 'Unity · on 1 device · one device after the other'],
        [config({allDevices: false}), 'Unity · on 0 devices · one device after the other'],
        [config({type: 99 as TestType}), 'Test · on all devices · one device after the other'],
    ])('describes %#', (c, expected) => {
        expect(describeConfig(c)).toBe(expected);
    });
});

describe('runStats', () => {
    const at = (s: number): string => new Date(Date.UTC(2026, 0, 1, 0, 0, s)).toISOString();

    it('sums up the runs with a verdict', () => {
        const stats = runStats([
            {createdAt: at(0), endedAt: at(30), status: 'passed'},
            {createdAt: at(0), endedAt: at(10), status: 'passed'},
            {createdAt: at(0), status: 'failed'},
            {createdAt: at(0), status: 'running'},
            {createdAt: at(0), endedAt: at(0), status: 'interrupted'},
        ]);
        expect(stats).toEqual({judged: 3, passed: 2, lastPassedAt: at(0), typicalMs: 10000});
    });

    it('has nothing without runs, and ignores negative durations', () => {
        expect(runStats([])).toEqual({judged: 0, passed: 0, lastPassedAt: null, typicalMs: null});
        expect(runStats([{createdAt: at(10), endedAt: at(0), status: 'passed'}]).typicalMs).toBeNull();
    });
});
