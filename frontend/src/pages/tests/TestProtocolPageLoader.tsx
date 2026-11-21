import React from 'react';
import {Navigate, useLocation, useParams} from 'react-router-dom';
import PageLoader from '../../components/PageLoader';
import {useApplicationContext} from '../../hooks/ApplicationProvider';
import {useLoader} from '../../hooks/useLoader';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useTestRunService} from '../../services/test.run.service';
import ITestProtocolData from '../../types/test.protocol';
import {RunRef} from '../../types/test.run.refs';
import TestProtocolPage from './TestProtocolPage';
import {buildRows, leaves} from './testRunModel';

// the protocol a runner wrapped; the runner's own page redirects to it
const wrappedProtocolId = (runProtocols: ITestProtocolData[] | null, protocolId: string | undefined): number | undefined =>
    runProtocols
        ? leaves(buildRows(runProtocols)).find(r => r.runner?.id === Number(protocolId) && r.protocol.id !== r.runner.id)?.protocol.id as number | undefined
        : undefined;

const TestProtocolPageLoader: React.FC = () => {
    const {projectIdentifier: projectId} = useProjectContext();
    const {testId, runId, protocolId} = useParams();
    const {appId} = useApplicationContext();
    const location = useLocation();
    const service = useTestRunService();
    const ref: RunRef = {projectId, appId: appId as number, testId: testId as string, runId: runId as string};

    // the protocol with its history and averages
    const {data: run, error, loading, retry} = useLoader(
        () => protocolId !== undefined ? service.protocol({...ref, protocolId}) : null,
        [service, projectId, testId, runId, protocolId, appId]);

    // the whole run, for the other tests of the run and the runner's
    // protocol; it stays loaded while moving between the tests of one run,
    // and the page works without it
    const {data: fullRun, retry: reloadRun} = useLoader(
        () => service.run(ref), [service, projectId, testId, runId, appId]);

    const runProtocols = fullRun && fullRun.testRun.id === Number(runId) ? fullRun.testRun.protocols : null;
    const wrapped = wrappedProtocolId(runProtocols, protocolId);
    if (wrapped !== undefined) {
        const path = location.pathname.replace(/\/[^/]+\/?$/, `/${wrapped}`);
        return <Navigate replace={true} to={`${path}${location.search}`}/>;
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
