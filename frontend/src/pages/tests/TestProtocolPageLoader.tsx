import React from 'react';
import {getRun, getTestProtocol} from '../../services/test.run.service';
import {useParams} from 'react-router-dom';
import TestProtocolPage from './TestProtocolPage';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {useApplicationContext} from "../../hooks/ApplicationProvider";
import {useLoader} from "../../hooks/useLoader";
import PageLoader from "../../components/PageLoader";

const TestProtocolPageLoader: React.FC = () => {
    const {projectIdentifier} = useProjectContext();
    const {testId, runId, protocolId} = useParams();
    const {appId} = useApplicationContext()

    // the protocol with its history and averages
    const {data: run, error, loading, retry} = useLoader(
        () => protocolId !== undefined
            ? getTestProtocol(projectIdentifier, appId as number, testId as string, runId as string, protocolId)
            : null,
        [projectIdentifier, testId, runId, protocolId, appId]);

    // the whole run, for the other tests of the run and the runner's
    // protocol; it stays loaded while moving between the tests of one run,
    // and the page works without it
    const {data: fullRun} = useLoader(
        () => getRun(projectIdentifier, appId as number, testId as string, runId as string),
        [projectIdentifier, testId, runId, appId]);

    if (!run || error) {
        return <PageLoader error={error} onRetry={retry}/>;
    }
    const protocol = run.protocols.find(p => p.id === Number(protocolId));
    if (!protocol) {
        // the endpoint answers with just the requested protocol, so another
        // one is the previous response while the next protocol loads
        const stale = loading || run.protocols.length > 0;
        return <PageLoader error={stale ? undefined : new Error(`protocol ${protocolId} not found in run ${runId}`)}/>;
    }
    const runProtocols = fullRun && fullRun.testRun.id === Number(runId) ? fullRun.testRun.protocols : null;
    return <TestProtocolPage key={protocol.id} run={run} protocol={protocol} runProtocols={runProtocols}/>;
};

export default TestProtocolPageLoader;
