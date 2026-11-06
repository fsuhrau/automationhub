import React from 'react';
import {getLastRun, getRun} from '../../services/test.run.service';
import {useParams} from 'react-router-dom';
import TestRunPage from './TestRunPage';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {useApplicationContext} from "../../hooks/ApplicationProvider";
import {useLoader} from "../../hooks/useLoader";
import PageLoader from "../../components/PageLoader";

const TestRunPageLoader: React.FC = () => {
    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const {testId, runId} = useParams();

    const {data, error, retry} = useLoader(
        () => runId !== undefined
            ? getRun(projectIdentifier, appId as number, testId as string, runId)
            : getLastRun(projectIdentifier, appId as number, testId as string),
        [projectIdentifier, testId, appId, runId]);

    if (!data || error) {
        return <PageLoader error={error} onRetry={retry}/>;
    }
    // key: a new run gets a fresh page (its state and SSE subscriptions)
    return <TestRunPage key={data.testRun.id} testRun={data.testRun} nextRunId={data.nextRunId} prevRunId={data.prevRunId}/>;
};

export default TestRunPageLoader;
