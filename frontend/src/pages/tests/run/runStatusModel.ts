// How the run page words and colors a run and its results.
import IDeviceData from '../../../types/device';
import {TestResultState} from '../../../types/test.result.state.enum';
import {RunStatus} from '../testRunModel';

export type PaletteKey = 'success' | 'error' | 'warning' | 'info';

export const STATUS: Record<RunStatus, { label: string, color: PaletteKey | 'default', hint: string }> = {
    running: {label: 'Running', color: 'info', hint: 'The run is executing; results update live.'},
    failed: {label: 'Failed', color: 'error', hint: 'At least one test failed.'},
    unstable: {label: 'Unstable', color: 'warning', hint: 'No test failed, but some were unstable.'},
    interrupted: {label: 'Interrupted', color: 'default', hint: 'The run stopped before all tests finished (cancelled, crashed, or the master restarted).'},
    success: {label: 'Passed', color: 'success', hint: 'All tests passed.'},
    not_run: {label: 'Did not run', color: 'error', hint: 'No test ran; the execution log has the errors.'},
    empty: {label: 'No results', color: 'default', hint: 'The run has no test results.'},
};

// results in the order they matter when reading a run
export const RESULT_ORDER: TestResultState[] = [
    TestResultState.TestResultFailed,
    TestResultState.TestResultUnstable,
    TestResultState.TestResultOpen,
    TestResultState.TestResultSuccess,
];

export const resultLabel = (r: TestResultState, running: boolean): string => {
    switch (r) {
        case TestResultState.TestResultFailed:
            return 'Failed';
        case TestResultState.TestResultUnstable:
            return 'Unstable';
        case TestResultState.TestResultSuccess:
            return 'Passed';
        default:
            return running ? 'Running' : 'Not finished';
    }
};

export const resultPalette = (r: TestResultState): PaletteKey => {
    switch (r) {
        case TestResultState.TestResultFailed:
            return 'error';
        case TestResultState.TestResultUnstable:
            return 'warning';
        case TestResultState.TestResultSuccess:
            return 'success';
        default:
            return 'info';
    }
};

// the name a device is shown by: its alias when it has one
export const deviceLabel = (device: IDeviceData | null | undefined): string =>
    device ? (device.alias?.length > 0 ? device.alias : device.name) : '';
