import React, {useEffect, useState} from 'react';
import {getLastRun, getRun} from '../../services/test.run.service';
import {useParams} from 'react-router-dom';
import ITestRunData from '../../types/test.run';
import TestRunPage from './TestRunPage';
import {Backdrop, CircularProgress} from '@mui/material';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {useApplicationContext} from "../../hooks/ApplicationProvider";
import {useError} from "../../ErrorProvider";

const TestRunPageLoader: React.FC = () => {

    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();

    const {testId, runId} = useParams();
    const {setError} = useError()

    const [uiState, setUiState] = useState<{
        testRun: ITestRunData | null,
        prevRunId: number | null,
        nextRunId: number | null
    }>({
        testRun: null,
        prevRunId: null,
        nextRunId: null,
    })

    useEffect(() => {
        if (runId !== undefined) {
            getRun(projectIdentifier, appId as number, testId as string, runId).then(response => {
                setUiState(prevState => ({
                    ...prevState,
                    testRun: response.data.TestRun,
                    nextRunId: response.data.NextRunId,
                    prevRunId: response.data.PrevRunId
                }))
            }).catch(ex => {
                setError(ex);
            });
        } else {
            getLastRun(projectIdentifier, appId as number, testId as string).then(response => {
                setUiState(prevState => ({
                    ...prevState,
                    testRun: response.data.TestRun,
                    nextRunId: response.data.NextRunId,
                    prevRunId: response.data.PrevRunId
                }))
            }).catch(ex => {
                setError(ex);
            });
        }
    }, [projectIdentifier, testId, appId, runId]);

    return (
        uiState.testRun
            ? <TestRunPage testRun={uiState.testRun} nextRunId={uiState.nextRunId} prevRunId={uiState.prevRunId}/>
            : <Backdrop sx={{zIndex: 1, color: '#fff'}} open={true}><CircularProgress color="inherit"/></Backdrop>
    );
};

export default TestRunPageLoader;
