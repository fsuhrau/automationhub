import React, {useEffect} from 'react';
import DeviceTableComponent from '../../components/NodeDevices';
import {useNavigate} from 'react-router-dom';
import {Box} from "@mui/system";
import Grid from "@mui/material/Grid";
import {TitleCard} from "../../components/title.card.component";
import {getNodes} from "../../services/settings.service";
import {HubStateActions} from "../../application/HubState";
import {useProjectContext} from "../../hooks/ProjectProvider";
import {useHubState} from "../../hooks/HubStateProvider";
import {useError} from "../../ErrorProvider";

const DevicesPage: React.FC = () => {
    const navigate = useNavigate();
    const {dispatch} = useHubState()
    const {setError} = useError()
    const {projectIdentifier} = useProjectContext()

    useEffect(() => {
        getNodes(projectIdentifier).then(response => {
            dispatch({
                type: HubStateActions.NodesUpdate,
                payload: response.data
            })
        }).catch(ex => {
            setError(ex);
        });
    }, [projectIdentifier]);

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <TitleCard title={'Nodes'}>
                <Grid container={true}>
                    <Grid size={{xs: 6}} container={true} sx={{
                        padding: 2,
                    }}>
                    </Grid>
                    <Grid size={{xs: 6}} container={true} justifyContent={"flex-end"} sx={{
                        padding: 1,
                    }}>
                    </Grid>
                    <Grid size={{xs: 12}}>
                        <DeviceTableComponent/>
                    </Grid>
                </Grid>
            </TitleCard>
        </Box>
    );
};

export default DevicesPage;
