// The project overview without React: what to show where.
import {IOverviewRun} from '../../types/overview.run';
import {IProjectOverview} from '../../types/overview.project';
import {availability, DeviceLiveState} from '../devices/deviceDetailsModel';
import {describeController} from '../../components/DeviceController';

// "3 passed · 1 failed": the counts of a run that are not zero
export const runCounts = (r: IOverviewRun): string =>
    [r.passed && `${r.passed} passed`, r.failed && `${r.failed} failed`, r.unstable && `${r.unstable} unstable`, r.open && `${r.open} open`]
        .filter(Boolean).join(' · ');

export interface Readiness {
    ready: number,
    inUse: number,
}

// how many devices can run a test now, and how many someone holds
export const readiness = (devices: DeviceLiveState[]): Readiness => {
    const states = devices.map(d => availability(d, describeController));
    return {
        ready: states.filter(a => a.label === 'Ready').length,
        inUse: states.filter(a => a.canUnlock).length,
    };
};

// what the overview has to say first: the project needs apps, then tests, then it shows its runs
export type Stage = 'noApps' | 'noTests' | 'runs';

export const stageOf = (overview: IProjectOverview): Stage => {
    if (overview.apps === 0) return 'noApps';
    return overview.tests === 0 ? 'noTests' : 'runs';
};
