import React, { useEffect, useState } from 'react';
import { getLastRun, getRun } from '../../services/test.run.service';
import { useParams } from 'react-router-dom';
import ITestRunData from '../../types/test.run';
import TestRunContent from '../../components/testrun-content.component';
import { Backdrop, CircularProgress } from '@mui/material';
import { useProjectAppContext } from "../../project/app.context";

const TestRunPageLoader: React.FC = () => {

    const {projectId, appId} = useProjectAppContext();

    const { testId, runId } = useParams();

    const [testRun, setTestRun] = useState<ITestRunData>();
    const [nextRunId, setNextRunId] = useState<number>();
    const [prevRunId, setPrevRunId] = useState<number>();

    useEffect(() => {
        if (runId !== undefined) {
            getRun(projectId, appId, testId as string, runId).then(response => {
                setTestRun(response.data.TestRun);
                setNextRunId(response.data.NextRunId);
                setPrevRunId(response.data.PrevRunId);
            }).catch(ex => {
                console.log(ex);
            });
        } else {
            getLastRun(projectId, appId, testId as string).then(response => {
                setTestRun(response.data.TestRun);
                setNextRunId(response.data.NextRunId);
                setPrevRunId(response.data.PrevRunId);
            }).catch(ex => {
                console.log(ex);
            });
        }
    }, [testId, runId]);

    return (
        testRun
            ? <TestRunContent testRun={ testRun } nextRunId={ nextRunId } prevRunId={ prevRunId }/>
            : <Backdrop sx={ { zIndex: 1, color: '#fff' } } open={ true }><CircularProgress color="inherit"/></Backdrop>
    );
};

export default TestRunPageLoader;
