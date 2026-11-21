// Builders of a recorded run for the replay tests. Times are ms since T0.
import ITesRunLogEntryData from '../../../types/test.run.log.entry';
import ITestRunData from '../../../types/test.run';
import ITestProtocolData from '../../../types/test.protocol';
import ITestRunRecordingData from '../../../types/test.run.recording';
import IProtocolEntryData from '../../../types/protocol.entry';
import {TestResultState} from '../../../types/test.result.state.enum';

export const T0 = Date.parse('2025-01-01T10:00:00Z');

export const at = (ms: number): Date => new Date(T0 + ms);

export const recording = (over: Partial<ITestRunRecordingData> = {}): ITestRunRecordingData => ({
    id: 1, testRunId: 7, deviceId: 5, codec: 'mp4', fileName: 'a b.mp4', width: 100, height: 200, size: 1,
    startedAt: at(0).toISOString(), endedAt: at(10000).toISOString(), ...over,
});

export const entry = (over: Partial<IProtocolEntryData>): IProtocolEntryData => ({
    id: 1, testProtocolId: 11, source: 'runner', level: 'info', message: 'msg', data: '', runtime: 0,
    timestamp: at(0), createdAt: at(0), updatedAt: at(0), deletedAt: at(0), ...over,
});

export const protocol = (over: Partial<ITestProtocolData> = {}): ITestProtocolData => ({
    id: 11, deviceId: 5, testName: 'Ns.Cls/Login', startedAt: at(1000), endedAt: at(6000),
    testResult: TestResultState.TestResultFailed, entries: [
        entry({id: 21, source: 'step', message: 'tap', timestamp: at(2000)}),
        entry({id: 22, source: 'screen', message: '', data: 'shot one.png', timestamp: at(3000)}),
    ], ...over,
} as ITestProtocolData);

interface LogSpec {
    id: number,
    log: string,
    // ms since T0, or any value for a broken time
    time: number | Date,
    level?: string,
}

export const logEntry = ({id, log, time, level = 'info'}: LogSpec): ITesRunLogEntryData =>
    ({id, testRunId: 7, level, log, createdAt: typeof time === 'number' ? at(time) : time} as ITesRunLogEntryData);

export const run = (over: Partial<ITestRunData> = {}): ITestRunData => ({
    id: 7, testId: 3, test: {name: 'Checkout'}, createdAt: at(0), recordings: [recording()], protocols: [protocol()],
    log: [logEntry({id: 1, log: 'boot', time: 500}), logEntry({id: 2, log: 'crash', time: 8000, level: 'error'})], ...over,
} as unknown as ITestRunData);
