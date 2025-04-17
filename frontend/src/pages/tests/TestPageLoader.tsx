import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Backdrop, CircularProgress } from '@mui/material';
import ITestData from '../../types/test';
import EditTestPage from './EditTestPage';
import { getTest } from '../../services/test.service';
import ShowTestPage from './ShowTestPage';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {useApplicationContext} from "../../hooks/ApplicationProvider";
import {useError} from "../../ErrorProvider";

interface TestPageProps {
    edit: boolean
}

const TestPageLoader: React.FC<TestPageProps> = (props) =>  {
    const { edit } = props;

    const {projectIdentifier} = useProjectContext();
    const {setError} = useError()

    const { testId } = useParams();
    const { appId } = useApplicationContext()

    const [test, setTest] = useState<ITestData>();


    useEffect(() => {
        if (appId !== null && testId !== undefined) {
            getTest(projectIdentifier, appId, testId).then(response => {
                setTest(response.data);
            }).catch(ex => {
                setError(ex);
            });
        }
    }, [projectIdentifier, appId, testId]);
    return (
        <div>
            { test ? (edit ? (<EditTestPage test={ test }/>) : (<ShowTestPage test={ test }/>) ) : (<Backdrop open={true} ><CircularProgress color="inherit" /></Backdrop>) }
        </div>
    );
};

export default TestPageLoader;
