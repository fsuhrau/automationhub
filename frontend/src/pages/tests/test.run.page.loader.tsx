import { FC, useEffect, useState } from 'react';
import { getLastRun, getRun } from '../../services/test.run.service';
import { useParams } from 'react-router-dom';
import ITestRunData from '../../types/test.run';
import TestRunContent from '../../components/testrun-content.component';
import { Backdrop, CircularProgress } from '@mui/material';
import { makeStyles } from '@mui/styles';

const useStyles = makeStyles(theme => ({
    backdrop: {
        zIndex: theme.zIndex.drawer + 1,
        color: '#fff',
    },
}));

interface ParamTypes {
    testId: string
    runId: string
}

const TestRunPageLoader: FC = () => {
    const classes = useStyles();
    const { testId, runId } = useParams<ParamTypes>();

    const [testRun, setTestRun] = useState<ITestRunData>();
    const [nextRunId, setNextRunId] = useState<number>();
    const [prevRunId, setPrevRunId] = useState<number>();

    useEffect(() => {
        if (runId !== undefined) {
            getRun(testId, runId).then(response => {
                setTestRun(response.data.TestRun);
                setNextRunId(response.data.NextRunId);
                setPrevRunId(response.data.PrevRunId);
            }).catch(ex => {
                console.log(ex);
            });
        } else {
            getLastRun(testId).then(response => {
                setTestRun(response.data.TestRun);
                setNextRunId(response.data.NextRunId);
                setPrevRunId(response.data.PrevRunId);
            }).catch(ex => {
                console.log(ex);
            });
        }
    }, [testId, runId]);

    return (
        <div>
            { testRun
                ? <TestRunContent testRun={ testRun } nextRunId={ nextRunId } prevRunId={ prevRunId }/>
                : <Backdrop className={ classes.backdrop } open={ true }>
                    <CircularProgress color="inherit"/>
                </Backdrop>
            }
        </div>
    );
};

export default TestRunPageLoader;
