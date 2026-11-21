import React from 'react';
import {useParams} from 'react-router-dom';
import PageLoader from '../../components/PageLoader';
import {useApplicationContext} from '../../hooks/ApplicationProvider';
import {useLoader} from '../../hooks/useLoader';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useTestRunService} from '../../services/test.run.service';
import TestRunPage from './TestRunPage';

const TestRunPageLoader: React.FC = () => {
    const {projectIdentifier: projectId} = useProjectContext();
    const {appId} = useApplicationContext();
    const {testId, runId} = useParams();
    const service = useTestRunService();

    const {data, error, retry} = useLoader(
        () => {
            const test = {projectId, appId: appId as number, testId: testId as string};
            return runId !== undefined ? service.run({...test, runId}) : service.lastRun(test);
        },
        [service, projectId, testId, appId, runId]);

    if (!data || error) {
        return <PageLoader error={error} onRetry={retry}/>;
    }
    // key: a new run gets a fresh page (its state and SSE subscriptions)
    return <TestRunPage key={data.testRun.id} testRun={data.testRun} nextRunId={data.nextRunId} prevRunId={data.prevRunId}
        state={data.state} devices={data.devices}/>;
};

export default TestRunPageLoader;
