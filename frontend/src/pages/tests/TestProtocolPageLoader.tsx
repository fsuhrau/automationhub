import React from 'react';
import {getRun, getTestProtocol} from '../../services/test.run.service';
import {Navigate, useLocation, useParams} from 'react-router-dom';
import TestProtocolPage from './TestProtocolPage';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {useApplicationContext} from "../../hooks/ApplicationProvider";
import {useLoader} from "../../hooks/useLoader";
import PageLoader from "../../components/PageLoader";
import {buildRows, leaves} from "./testRunModel";

const TestProtocolPageLoader: React.FC = () => {
    const {projectIdentifier} = useProjectContext();
    const {testId, runId, protocolId} = useParams();
    const {appId} = useApplicationContext()
    const location = useLocation();

    // the protocol with its history and averages
    const {data: run, error, loading, retry} = useLoader(
        () => protocolId !== undefined
            ? getTestProtocol(projectIdentifier, appId as number, testId as string, runId as string, protocolId)
            : null,
        [projectIdentifier, testId, runId, protocolId, appId]);

    // the whole run, for the other tests of the run and the runner's
    // protocol; it stays loaded while moving between the tests of one run,
    // and the page works without it
    const {data: fullRun, retry: reloadRun} = useLoader(
        () => getRun(projectIdentifier, appId as number, testId as string, runId as string),
        [projectIdentifier, testId, runId, appId]);

    // a runner's protocol is shown on the page of the test it wrapped; old
    // links to the runner go there
    const runProtocols = fullRun && fullRun.testRun.id === Number(runId) ? fullRun.testRun.protocols : null;
    const wrapped = runProtocols
        ? leaves(buildRows(runProtocols)).find(r => r.runner?.id === Number(protocolId) && r.protocol.id !== r.runner.id)
        : undefined;
    if (wrapped) {
        const path = location.pathname.replace(/\/[^/]+\/?$/, `/${wrapped.protocol.id}`);
        return <Navigate replace to={`${path}${location.search}`}/>;
    }

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
    return <TestProtocolPage key={protocol.id} run={run} protocol={protocol} runProtocols={runProtocols}
                             onRunChanged={reloadRun}/>;
};

export default TestProtocolPageLoader;
