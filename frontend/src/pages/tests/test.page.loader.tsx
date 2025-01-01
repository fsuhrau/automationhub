import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Backdrop, CircularProgress } from '@mui/material';
import ITestData from '../../types/test';
import EditTestPage from './edit.test.content';
import { getTest } from '../../services/test.service';
import ShowTestPage from './show.test.content';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {useHubState} from "../../hooks/HubStateProvider";

interface TestPageProps {
    edit: boolean
}

const TestPageLoader: React.FC<TestPageProps> = (props) =>  {
    const { edit } = props;

    const {state, dispatch} = useHubState()
    const {project, projectIdentifier} = useProjectContext();

    const { testId } = useParams();

    const [test, setTest] = useState<ITestData>();


    useEffect(() => {
        if (state.appId !== null && testId !== undefined) {
            getTest(projectIdentifier, state.appId, testId).then(response => {
                setTest(response.data);
            }).catch(ex => {
                console.log(ex);
            });
        }
    }, [projectIdentifier, state.appId, testId]);
    return (
        <div>
            { test ? (edit ? (<EditTestPage test={ test }/>) : (<ShowTestPage test={ test }/>) ) : (<Backdrop open={true} ><CircularProgress color="inherit" /></Backdrop>) }
        </div>
    );
};

export default TestPageLoader;
