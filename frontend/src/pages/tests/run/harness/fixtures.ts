// Builders of the data the run and protocol pages show. Each takes the
// fields a test cares about and fills in the rest.
import IDeviceData from '../../../../types/device';
import {DeviceType} from '../../../../types/device.type.enum';
import {DeviceConnectionType} from '../../../../types/device.connection.type.enum';
import {DeviceStateType} from '../../../../types/deviceStateType';
import {PlatformType} from '../../../../types/platform.type.enum';
import IProtocolEntryData from '../../../../types/protocol.entry';
import IProtocolPerformanceEntryData from '../../../../types/protocol.performance.entry';
import ITestProtocolData from '../../../../types/test.protocol';
import ITestRunData from '../../../../types/test.run';
import ITesRunLogEntryData from '../../../../types/test.run.log.entry';
import {TestResultState} from '../../../../types/test.result.state.enum';

const T0 = '2024-05-01T10:00:00.000Z';
export const at = (seconds: number): Date => new Date(new Date(T0).valueOf() + seconds * 1000);

export const device = (over: Partial<IDeviceData> = {}): IDeviceData => ({
    id: 7, nodeId: 1, companyId: 1, deviceIdentifier: 'abc', deviceType: DeviceType.Phone, platformType: PlatformType.iOS,
    name: 'iPhone 12', alias: '', os: 'iOS', osVersion: '17.1', status: DeviceStateType.Booted, isLocked: false, manager: 'wda',
    isAcknowledged: true, connectionType: DeviceConnectionType.USB, connectionParameter: {connectionType: DeviceConnectionType.USB, ip: '', port: 0},
    customParameter: [], deviceParameter: [], ...over,
});

export const protocol = (over: Partial<ITestProtocolData> = {}): ITestProtocolData => ({
    id: 1, testRun: {} as ITestRunData, testRunId: 3, parentTestProtocolId: null, deviceId: 7, device: device(),
    testName: 'Game.Tests/Login', startedAt: at(0), endedAt: at(10), entries: [], testResult: TestResultState.TestResultSuccess,
    performance: [], avgFps: 0, avgMem: 0, avgCpu: 0, avgVertexCount: 0, avgTriangles: 0,
    histAvgFps: 0, histAvgMem: 0, histAvgCpu: 0, histAvgVertexCount: 0, histAvgTriangles: 0,
    testProtocolHistory: null, createdAt: at(0), updatedAt: at(0), deletedAt: at(0), childProtocols: [], ...over,
});

export const entry = (over: Partial<IProtocolEntryData> = {}): IProtocolEntryData => ({
    id: 1, testProtocolId: 1, source: 'step', level: 'info', message: 'tap play', data: '', runtime: 1,
    createdAt: at(1), updatedAt: at(1), deletedAt: at(1), ...over,
});

export const performance = (over: Partial<IProtocolPerformanceEntryData> = {}): IProtocolPerformanceEntryData => ({
    id: 1, testProtocolId: 1, checkpoint: 'menu', fps: 60, mem: 500, cpu: 20, vertexCount: 1000, triangles: 2000, other: '',
    runtime: 2, executionTime: 2, createdAt: at(2), updatedAt: at(2), deletedAt: at(2), ...over,
});

export const logEntry = (over: Partial<ITesRunLogEntryData> = {}): ITesRunLogEntryData => ({
    id: 1, testRunId: 3, level: 'info', log: 'started', createdAt: at(1), updatedAt: at(1), deletedAt: at(1), ...over,
});

export const run = (over: Partial<ITestRunData> = {}): ITestRunData => ({
    id: 3, testId: 2, test: {id: 2, appId: 1, name: 'Smoke', testConfig: {testId: 2} as ITestRunData['test']['testConfig']} as ITestRunData['test'],
    sessionId: 's', appBinaryId: 5, appBinary: null, startUrl: '', parameter: '', testResult: TestResultState.TestResultSuccess,
    protocols: [], log: [], deviceStatus: [], createdAt: at(0), updatedAt: at(0), deletedAt: at(0), ...over,
});
