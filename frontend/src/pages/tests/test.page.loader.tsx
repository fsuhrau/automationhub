import { FC, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Backdrop, CircularProgress, Typography } from '@material-ui/core';
import ITestData from '../../types/test';
import EditTestPage from './edit.test.content';
import { getTest } from '../../services/test.service';
import ShowTestPage from './show.test.content';

interface TestPageProps {
    edit: boolean
}

interface ParamTypes {
    testId: string
}

const TestPageLoader: FC<TestPageProps> = (props) =>  {
    const { edit } = props;
    const { testId } = useParams<ParamTypes>();
    const [test, setTest] = useState<ITestData>();

    useEffect(() => {
        getTest(testId).then(response => {
            setTest(response.data);
        }).catch(ex => {
            console.log(ex);
        });
    }, [testId]);
    return (
        <div>
            { test
                ? (edit ? <EditTestPage test={ test }/> : <ShowTestPage test={ test }/> )
                : <Backdrop open={true} >
                    <CircularProgress color="inherit" />
                </Backdrop>
            }
        </div>
    );
};

export default TestPageLoader;
