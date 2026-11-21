// Setup shared by the tests of the protocol page.
import {screen} from '@testing-library/react';
import ITestProtocolData from '../../../../types/test.protocol';
import {TestResultState as S} from '../../../../types/test.result.state.enum';
import {fakeApi, FakeApi} from '../../../../testing/fakeApi';
import {FakeSse} from '../../../../testing/fakeSse';
import {at, entry, performance, protocol, run} from './fixtures';
import {renderAt} from './render';

export const base = '/p1/app/1/test/2/run/3';
export const url = '/project/p1/app:1/test/2/run/3/1';

// the same test in the earlier run 2
export const earlier = protocol({
    id: 20, testRunId: 2, startedAt: at(-100),
    performance: [performance({checkpoint: 'menu', runtime: 1.5, fps: 60}), performance({checkpoint: 'game', runtime: 6, fps: 59})],
});

export const perf = [
    performance({id: 1, checkpoint: 'schedule', runtime: 0}),
    performance({id: 2, checkpoint: 'menu', runtime: 2, fps: 58}),
    performance({id: 3, checkpoint: 'game', runtime: 5, fps: 40}),
];

export const steps = [
    entry({id: 1, message: 'open menu', runtime: 1}),
    entry({id: 2, message: 'tap play\nsecond line\nthird line', runtime: 2.5}),
];

// a passed test with steps, performance and an earlier run
export const login = (over: Partial<ITestProtocolData> = {}): ITestProtocolData => protocol({
    id: 1, testName: 'Game.Tests/Login', entries: steps, performance: perf, avgFps: 55, histAvgFps: 60, testProtocolHistory: [earlier], ...over,
});

export const shop = protocol({id: 2, testName: 'Game.Tests/Shop', testResult: S.TestResultFailed, startedAt: at(11), endedAt: at(20)});

export interface Setup {
    api: FakeApi,
    sse: FakeSse,
}

// opens the page of `main`, whose run holds the other protocols too
export interface OpenOptions {
    // more answers of the fake API
    answers?: Record<string, unknown>,
    at?: string,
}

export const openProtocol = (main: ITestProtocolData, others: ITestProtocolData[] = [shop], {answers = {}, at: where = url}: OpenOptions = {}): Setup => {
    const all = [main, ...others].sort((a, b) => (a.id as number) - (b.id as number));
    const api = fakeApi({
        [`GET ${base}/${main.id}`]: run({protocols: [main]}),
        [`GET ${base}`]: {testRun: run({protocols: all}), nextRunId: null, prevRunId: null},
        ...answers,
    });
    const sse = new FakeSse();
    renderAt(where, {api, sse});
    return {api, sse};
};

export const tab = (name: RegExp): HTMLElement => screen.getByRole('tab', {name});
