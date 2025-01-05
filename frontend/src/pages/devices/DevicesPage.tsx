import React from 'react';
import DeviceTableComponent from '../../components/device-table.component';
import {useNavigate} from 'react-router-dom';
import {Box} from "@mui/system";
import Button from "@mui/material/Button";
import Grid from "@mui/material/Grid2";
import {TitleCard} from "../../components/title.card.component";

const DevicesPage: React.FC = () => {
    const navigate = useNavigate();

    function onManageDevices(): void {
        navigate('/devices/manager');
    }

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <TitleCard title={'Nodes'}>
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
                        <Button variant={"text"} onClick={onManageDevices}>Manage</Button>
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
