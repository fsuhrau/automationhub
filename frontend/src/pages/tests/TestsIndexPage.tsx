import React from 'react';
import TestsTable from '../../components/tests-table.component';
import {useNavigate} from 'react-router-dom';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {TitleCard} from "../../components/title.card.component";
import {useApplicationContext} from "../../hooks/ApplicationProvider";
import {Box} from "@mui/system";
import Button from "@mui/material/Button";
import {AddRounded} from "@mui/icons-material";
import {useHubState} from "../../hooks/HubStateProvider";

const TestsIndexPage: React.FC = () => {

    const {projectIdentifier} = useProjectContext();
    const {state} = useHubState();
    const {appId} = useApplicationContext();
    const navigate = useNavigate();

    const app = state.apps?.find(a => a.id === appId);

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <TitleCard title={'Tests'}
                       subtitle={`Test configurations of ${app?.name ?? 'this app'}. Open one for its details and runs, or start a run.`}
                       actions={<Button variant={"contained"} size={'small'} startIcon={<AddRounded/>}
                                        onClick={() => navigate(`/project/${projectIdentifier}/app:${appId}/test/new`)}>
                           New test
                       </Button>}>
                <TestsTable appId={appId}/>
            </TitleCard>
        </Box>
    );
};

export default TestsIndexPage;
