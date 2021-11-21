import { FC, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Typography } from '@material-ui/core';
import ITestData from '../../types/test';
import EditTestPage from './edit.test.content';
import { getTest } from '../../services/test.service';
import ShowTestPage from './show.test.content';

interface TestPageProps {
    edit: boolean
}

const TestPage: FC<TestPageProps> = (props) =>  {
    const { edit } = props;
    const { testId } = useParams();
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
                : <Typography variant={ 'h1' }>Loading</Typography>
            }
        </div>
    );
};

export default TestPage;
