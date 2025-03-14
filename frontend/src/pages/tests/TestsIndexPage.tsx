import React, {useEffect} from 'react';
import TestsTable from '../../components/tests-table.component';
import {useNavigate} from 'react-router-dom';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {TitleCard} from "../../components/title.card.component";
import {useApplicationContext} from "../../hooks/ApplicationProvider";
import {Box} from "@mui/system";
import Grid from "@mui/material/Grid2";
import Button from "@mui/material/Button";

const TestsIndexPage: React.FC = () => {

    const {project, projectIdentifier} = useProjectContext();

    const {appId} = useApplicationContext();

    const navigate = useNavigate();

    function newTestClick(): void {
        navigate(`/project/${projectIdentifier}/app:${appId}/test/new`)
    }

    useEffect(() => {
        if (appId === 0) {
            const value = project.Apps === undefined ? null : project.Apps.length === 0 ? null : project.Apps[0].ID;
            if (value !== null) {
                navigate(`/project/${projectIdentifier}/app:${appId}/tests`)
            }
        }
    }, [project.Apps, appId])

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <TitleCard title={'Tests'}>
                <Grid container={true}>
                    <Grid size={{xs: 6}} container={true} sx={{
                        padding: 2,
                        borderBottom: '1px solid rgba(0, 0, 0, 0.12)'
                    }}>
                    </Grid>
                    <Grid size={{xs: 6}} container={true} justifyContent={"flex-end"} sx={{
                        padding: 1,
                        borderBottom: '1px solid rgba(0, 0, 0, 0.12)'
                    }}>
                        <Button variant={"contained"} size={'small'} onClick={newTestClick}>Add new Test</Button>
                    </Grid>
                    <Grid size={{xs: 12}}>
                        <TestsTable appId={appId}/>
                    </Grid>
                </Grid>
            </TitleCard>
        </Box>
    );
};

export default TestsIndexPage;