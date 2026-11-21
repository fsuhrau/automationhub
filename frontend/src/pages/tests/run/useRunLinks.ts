import {useNavigate} from 'react-router-dom';
import {routes} from '../../../lib/routes';
import ITestRunData from '../../../types/test.run';
import {TestRef} from '../../../types/test.run.refs';
import {TestRow} from '../testRunModel';
import {replayQuery} from './runUrlModel';

export interface RunLinks {
    runUrl: (runId: number) => string,
    replayUrl: (time?: Date, deviceId?: number | null) => string,
    openRun: (runId: number) => void,
    openProtocol: (protocolId: number) => void,
    // opens the replay where the test of the row started
    openReplay: (row: TestRow) => void,
}

// useRunLinks builds the urls around a run and navigates to them
export const useRunLinks = (run: ITestRunData, test: TestRef): RunLinks => {
    const navigate = useNavigate();
    const testRoutes = routes.project(test.projectId).app(test.appId).test(test.testId);
    const runUrl = (runId: number): string => testRoutes.run(runId).show();
    const replayUrl = (time?: Date, deviceId?: number | null): string =>
        `${testRoutes.run(run.id as number).replay()}${replayQuery(time, deviceId)}`;
    return {
        runUrl,
        replayUrl,
        openRun: runId => navigate(runUrl(runId)),
        openProtocol: protocolId => navigate(testRoutes.run(run.id as number).protocol(protocolId)),
        openReplay: row => navigate(replayUrl(row.startedAt, row.protocol.deviceId)),
    };
};
