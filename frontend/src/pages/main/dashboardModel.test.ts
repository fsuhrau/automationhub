import {describe, expect, it} from 'vitest';
import {readiness, runCounts, stageOf} from './dashboardModel';
import {DeviceStateType} from '../../types/deviceStateType';
import {IOverviewRun} from '../../types/overview.run';
import {IProjectOverview} from '../../types/overview.project';

const run = (fields: Partial<IOverviewRun>): IOverviewRun =>
    ({id: 1, testId: 2, testName: 't', appId: 3, status: 'passed', passed: 0, failed: 0, unstable: 0, open: 0, ...fields} as IOverviewRun);

describe('runCounts', () => {
    it('lists the counts that are not zero', () => {
        expect(runCounts(run({passed: 3, failed: 1, unstable: 2, open: 4}))).toBe('3 passed · 1 failed · 2 unstable · 4 open');
        expect(runCounts(run({failed: 1}))).toBe('1 failed');
    });

    it('is empty for a run without counts', () => {
        expect(runCounts(run({}))).toBe('');
    });
});

describe('readiness', () => {
    const device = (status: DeviceStateType, extra = {}): { status: DeviceStateType, isLocked: boolean, lockedBy?: string } =>
        ({status, isLocked: false, ...extra});

    it('counts the devices that are ready and the ones someone holds', () => {
        const result = readiness([
            device(DeviceStateType.Booted),
            device(DeviceStateType.Booted),
            device(DeviceStateType.Booted, {lockedBy: 'mcp:claude:1'}),
            device(DeviceStateType.Locked, {isLocked: true}),
            device(DeviceStateType.Shutdown),
        ]);
        expect(result).toEqual({ready: 2, inUse: 2});
    });

    it('is zero without devices', () => {
        expect(readiness([])).toEqual({ready: 0, inUse: 0});
    });
});

describe('stageOf', () => {
    const overview = (apps: number, tests: number): IProjectOverview =>
        ({apps, tests, bundles: 0, bundleStorage: 0, failing: [], running: [], latest: []});

    it.each([[0, 0, 'noApps'], [0, 5, 'noApps'], [1, 0, 'noTests'], [1, 5, 'runs']] as const)('for %s apps and %s tests is %s', (apps, tests, stage) => {
        expect(stageOf(overview(apps, tests))).toBe(stage);
    });
});
