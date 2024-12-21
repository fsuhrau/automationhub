import React from 'react';
import Paper from '@mui/material/Paper';
import DeviceTableComponent from '../../components/device-table.component';
import { Divider, FormControl, IconButton, MenuItem, Select, Typography } from '@mui/material';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Grid from '@mui/material/Grid';
import { useNavigate } from 'react-router-dom';
import { Add } from '@mui/icons-material';
import { TitleCard } from "../../components/title.card.component";
import Button from "@mui/material/Button";

const Devices: React.FC = () => {
    const navigate = useNavigate();

    function onManageDevices(): void {
        navigate('/devices/manager');
    }

    return (
        <Grid container={ true } spacing={ 2 }>
            <Grid item={ true } container={true} xs={ 12 }>
                <Grid item={true} xs={true}>
                    <Typography variant={ "h1" }>Device Pool</Typography>
                </Grid>
                <Grid item={true}>
                    <Button variant={"text"}>Manage</Button>
                </Grid>
            </Grid>
            <Grid item={ true } xs={ 12 }>
                <Divider/>
            </Grid>
            <Grid item={ true } container={ true } xs={ 12 } alignItems={ "center" } justifyContent={ "center" }>
                <Grid
                    item={ true }
                    xs={ 12 }
                >
                    <DeviceTableComponent/>
                </Grid>
            </Grid>
        </Grid>
    );
};

export default Devices;
