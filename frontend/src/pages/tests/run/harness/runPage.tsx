// Setup shared by the tests of the test run page.
import {IAppBinaryData} from '../../../../types/app';
import {PlatformType} from '../../../../types/platform.type.enum';
import ITestRunData from '../../../../types/test.run';
import {TestResultState as S} from '../../../../types/test.result.state.enum';
import {fakeApi, FakeApi} from '../../../../testing/fakeApi';
import {FakeSse} from '../../../../testing/fakeSse';
import {at, device, logEntry, protocol, run} from './fixtures';
import {renderAt} from './render';

export const base = '/p1/app/1/test/2';
export const url = '/project/p1/app:1/test/2/run/3';

const bundle = {
    id: 5, name: 'Game', version: '1.4', identifier: 'com.game', additional: 'addon', hash: 'abc123', createdAt: at(0),
} as IAppBinaryData;

export const status = (deviceId: number, startupTime: number, histAvgStartupTime: number): ITestRunData['deviceStatus'][number] => ({
    id: deviceId, testRunId: 3, deviceId, device: device({id: deviceId}), startupTime, histAvgStartupTime,
    createdAt: at(0), updatedAt: at(0), deletedAt: at(0),
});

// a finished run with a passed and a failed test, one device, a bundle and parameters
export const fullRun = (over: Partial<ITestRunData> = {}): ITestRunData => run({
    protocols: [protocol({id: 1, testName: 'Game/Login'}), protocol({id: 2, testName: 'Game/Shop', testResult: S.TestResultFailed})],
    log: [logEntry({id: 1, log: 'run started'}), logEntry({id: 2, level: 'error', log: 'shop broke'})],
    deviceStatus: [status(7, 3000, 2000)],
    appBinary: bundle, parameter: 'A=1;B=', startUrl: 'game://start', ...over,
});

// a run that executes its first test
export const runningRun = (over: Partial<ITestRunData> = {}): ITestRunData => fullRun({
    protocols: [protocol({id: 1, testName: 'Game/Login', testResult: S.TestResultOpen, endedAt: undefined})], ...over,
});

export interface Setup {
    api: FakeApi,
    sse: FakeSse,
}

export interface OpenOptions {
    // fields of the response next to the run
    response?: Record<string, unknown>,
    // more answers of the fake API
    answers?: Record<string, unknown>,
    at?: string,
}

// the master's device 7 is an Android phone
const HUB_DEVICES = [device({id: 7, platformType: PlatformType.Android})];

export const openRun = (loaded: ITestRunData, {response = {}, answers = {}, at: where = url}: OpenOptions = {}): Setup => {
    const api = fakeApi({
        [`GET ${base}/run/3`]: {testRun: loaded, nextRunId: 4, prevRunId: 2, state: 'finished', devices: [], ...response},
        ...answers,
    });
    const sse = new FakeSse();
    renderAt(where, {api, sse, devices: HUB_DEVICES});
    return {api, sse};
};

