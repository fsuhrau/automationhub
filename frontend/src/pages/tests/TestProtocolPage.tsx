import React, {useState} from 'react';
import {Alert, Box} from '@mui/material';
import {useNavigate} from 'react-router-dom';
import {useApplicationContext} from '../../hooks/ApplicationProvider';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {routes} from '../../lib/routes';
import ITestProtocolData from '../../types/test.protocol';
import ITestRunData from '../../types/test.run';
import {ProtocolRef} from '../../types/test.run.refs';
import {ProtocolBody} from './run/parts/ProtocolBody';
import {ProtocolHeader} from './run/parts/ProtocolHeader';
import {ProtocolTabs} from './run/parts/ProtocolTabs';
import {ScreenViewer} from './run/parts/ScreenViewer';
import {useComparison} from './run/useComparison';
import {useLogView} from './run/useLogView';
import {useProtocolQuery} from './run/useProtocolQuery';
import {useProtocolStream} from './run/useProtocolStream';
import {useProtocolView} from './run/useProtocolView';
import {TestRow} from './testRunModel';

interface TestProtocolPageProps {
    run: ITestRunData,
    protocol: ITestProtocolData,
    // all protocols of the run, once loaded
    runProtocols: ITestProtocolData[] | null,
    // loads the run again, for the runner's log once it has ended
    onRunChanged: () => void,
}

// TestProtocolPage shows one test of a run: why it failed, what it did, and
// how it performed. The loader mounts it with key={protocol.id}.
const TestProtocolPage: React.FC<TestProtocolPageProps> = ({run, protocol, runProtocols, onRunChanged}) => {
    const {projectIdentifier: projectId} = useProjectContext();
    const {appId} = useApplicationContext();
    const navigate = useNavigate();
    const refs: ProtocolRef = {projectId, appId: appId as number, testId: run.testId, runId: run.id as number, protocolId: protocol.id as number};
    const query = useProtocolQuery();
    const view = useProtocolView(useProtocolStream(run.id, protocol), runProtocols, onRunChanged);
    const comparison = useComparison(refs, view.earlier, query);
    const logView = useLogView(view.log);
    const [viewer, setViewer] = useState<number | null>(null);
    const screens = view.logSummary.screens;
    const openTest = (row: TestRow): void => {
        const url = routes.project(projectId).app(refs.appId).test(run.testId).run(refs.runId).protocol(row.protocol.id as number);
        navigate(`${url}${query.keep()}`);
    };
    const counts = {log: view.log.length, screenshots: screens.length, performance: (view.protocol.performance ?? []).length};
    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
            <ProtocolHeader run={refs} view={view} onOpen={openTest}/>
            <ProtocolTabs tab={query.tab} onTab={query.setTab} counts={counts}/>
            <ProtocolBody refs={refs} view={view} query={query} comparison={comparison} logView={logView} onViewer={setViewer}/>
            {view.running && <Alert severity="info" sx={{mt: 2}}>The test is running; the log and result update live.</Alert>}
            <ScreenViewer screens={screens} index={viewer} onIndex={setViewer}/>
        </Box>
    );
};

export default TestProtocolPage;
