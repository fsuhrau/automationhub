import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Backdrop, CircularProgress } from '@mui/material';
import ITestData from '../../types/test';
import EditTestPage from './edit.test.content';
import { getTest } from '../../services/test.service';
import ShowTestPage from './show.test.content';
import { ApplicationProps } from "../../application/ApplicationProps";
import {useProjectContext} from "../../hooks/ProjectProvider";

interface TestPageProps extends ApplicationProps {
    edit: boolean
}

const TestPageLoader: React.FC<TestPageProps> = (props) =>  {
    const { edit, appState } = props;

    debugger;

    const { testId } = useParams();
    const {project, projectId} = useProjectContext();

    const [test, setTest] = useState<ITestData>();


    useEffect(() => {
        if (appState.appId !== null && testId !== undefined) {
            getTest(projectId as string, appState.appId, testId).then(response => {
                debugger;
                setTest(response.data);
            }).catch(ex => {
                console.log(ex);
            });
        }
    }, [projectId, appState.appId, testId]);
    return (
        <div>
            { test ? (edit ? (<EditTestPage appState={appState} test={ test }/>) : (<ShowTestPage test={ test }/>) ) : (<Backdrop open={true} ><CircularProgress color="inherit" /></Backdrop>) }
        </div>
    );
};

export default TestPageLoader;
