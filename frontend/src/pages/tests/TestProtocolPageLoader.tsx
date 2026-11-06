import React from 'react';
import {getTestProtocol} from '../../services/test.run.service';
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

    const {data: run, error, retry} = useLoader(
        () => protocolId !== undefined
            ? getTestProtocol(projectIdentifier, appId as number, testId as string, runId as string, protocolId)
            : null,
        [projectIdentifier, testId, runId, protocolId, appId]);

    if (!run || error) {
        return <PageLoader error={error} onRetry={retry}/>;
    }
    const protocol = run.protocols.find(p => p.id === Number(protocolId));
    if (!protocol) {
        return <PageLoader error={new Error(`protocol ${protocolId} not found in run ${runId}`)}/>;
    }
    return <TestProtocolPage key={protocol.id} run={run} protocol={protocol}/>;
};

export default TestProtocolPageLoader;
