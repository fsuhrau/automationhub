import React, {useEffect} from 'react';
import DeviceTableComponent from '../../components/NodeDevices';
import {Box} from "@mui/system";
import {TitleCard} from "../../components/title.card.component";
import {getNodes} from "../../services/settings.service";
import {HubStateActions} from "../../application/HubState";
import {useProjectContext} from "../../hooks/ProjectProvider";
import {useHubState} from "../../hooks/HubStateProvider";
import {useError} from "../../ErrorProvider";

const DevicesPage: React.FC = () => {
    const {dispatch} = useHubState()
    const {setError} = useError()
    const {projectIdentifier} = useProjectContext()

    useEffect(() => {
        getNodes(projectIdentifier).then(nodes => {
            dispatch({
                type: HubStateActions.NodesUpdate,
                payload: nodes
            })
        }).catch(ex => {
            setError(ex);
        });
    }, [projectIdentifier]);

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <TitleCard title={'Device Pool'}
                       subtitle={'All devices of the project, grouped by the node they are connected to. Open a device for its details, or take live control of it.'}>
                <DeviceTableComponent/>
            </TitleCard>
        </Box>
    );
};

export default DevicesPage;
