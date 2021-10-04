import { FC, useEffect, useState } from 'react';
import { getLastRun, getRun } from '../../services/test.run.service';
import { useParams } from 'react-router-dom';
import ITestRunData from '../../types/test.run';
import TestRunContent from '../../components/testrun-content.component';
import { Typography } from '@material-ui/core';

const TestRun: FC = () => {
    const { testId } = useParams();
    const { runId } = useParams<number>();
    const [testRun, setTestRun] = useState<ITestRunData>();

    useEffect(() => {
        if (runId !== undefined) {
            getRun(testId, runId).then(response => {
                setTestRun(response.data);
            }).catch(ex => {
                console.log(ex);
            });
        } else {
            getLastRun(testId).then(response => {
                setTestRun(response.data);
            }).catch(ex => {
                console.log(ex);
            });
        }
    }, [testId, runId]);

    return (
        <div>
            { testRun
                ? <TestRunContent testRun={ testRun } />
                : <Typography variant={ 'h1' }>Loading</Typography>
            }
        </div>
    );
};

export default TestRun;
