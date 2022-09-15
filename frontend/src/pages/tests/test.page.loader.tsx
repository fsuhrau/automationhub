import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Backdrop, CircularProgress } from '@mui/material';
import ITestData from '../../types/test';
import EditTestPage from './edit.test.content';
import { getTest } from '../../services/test.service';
import ShowTestPage from './show.test.content';

interface TestPageProps {
    edit: boolean
}

const TestPageLoader: React.FC<TestPageProps> = (props) =>  {
    const { edit } = props;
    const { project_id, app_id, testId } = useParams();
    const [test, setTest] = useState<ITestData>();

    useEffect(() => {
        getTest(project_id as string, +app_id as number, testId).then(response => {
            setTest(response.data);
        }).catch(ex => {
            console.log(ex);
        });
    }, [testId]);
    return (
        <div>
            { test ? (edit ? (<EditTestPage test={ test }/>) : (<ShowTestPage test={ test }/>) ) : (<Backdrop open={true} ><CircularProgress color="inherit" /></Backdrop>) }
        </div>
    );
};

export default TestPageLoader;
