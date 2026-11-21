// Opening the device page with a fake master, for the tests of DeviceShowPage.
import {ReactElement} from 'react';
import DeviceShowPage from './DeviceShowPage';
import IDeviceData from '../../types/device';
import {TestResultState} from '../../types/test.result.state.enum';
import {fakeApi, FakeApi, fakeSources, node, renderPage} from './testKit';

export const protocol = {
    id: 9, testRunId: 4, testName: 'Suite.Auth/Login', startedAt: new Date('2025-01-02T10:00:00Z'),
    endedAt: new Date('2025-01-02T10:00:30Z'), testResult: TestResultState.TestResultSuccess,
    testRun: {testId: 3, test: {appId: 2, name: 'Nightly'}},
};

export type Opened = ReturnType<typeof renderPage> & { log: ReturnType<typeof fakeSources> };

export const open = (stored: IDeviceData, api: FakeApi, live?: IDeviceData[]): Opened => {
    const log = fakeSources();
    const page: ReactElement = <DeviceShowPage device={stored} createLogSource={log.create}/>;
    const result = renderPage(page, {
        api, devices: live, nodes: [node()], route: 'device/:deviceId', url: `/project/p1/device/${stored.id}`,
    });
    return Object.assign(result, {log});
};

export const withHistory = (more: Record<string, unknown> = {}): FakeApi =>
    fakeApi({'GET /p1/device/1/protocols': [protocol], ...more});
