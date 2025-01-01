import React from 'react';
import DeviceTableComponent from '../../components/device-table.component';
import {Typography} from '@mui/material';
import {useNavigate} from 'react-router-dom';
import {Box} from "@mui/system";
import Button from "@mui/material/Button";
import Grid from "@mui/material/Grid2";
import {TitleCard} from "../../components/title.card.component";

const Devices: React.FC = () => {
    const navigate = useNavigate();

    function onManageDevices(): void {
        navigate('/devices/manager');
    }

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <Typography
                component="h2" variant="h6" sx={{mb: 2}}>
                Device Pool
            </Typography>
            <Grid
                container
                spacing={2}
                columns={12}
                sx={{mb: (theme) => theme.spacing(2)}}
            >

            </Grid>
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

export default Devices;
