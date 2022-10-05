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
            <Grid item={ true } xs={ 12 }>
                <Typography variant={ "h1" }>Device Pool</Typography>
            </Grid>
            <Grid item={ true } xs={ 12 }>
                <Divider/>
            </Grid>
            <Grid item={ true } container={ true } xs={ 12 } alignItems={ "center" } justifyContent={ "center" }>
                <Grid
                    item={ true }
                    xs={ 12 }
                    style={ {maxWidth: 1000} }
                >
                    <TitleCard title={ "Devices" }>
                        <Paper sx={ {margin: 'auto', overflow: 'hidden'} }>
                            <Grid container={ true }>
                                <Grid item={ true } xs={ 12 } container={ true } justifyContent={ "flex-end" } sx={ {
                                    padding: 1,
                                    backgroundColor: '#fafafa',
                                    borderBottom: '1px solid rgba(0, 0, 0, 0.12)'
                                } }>
                                    <Button variant={ "contained" }>Manage</Button>
                                </Grid>
                                <Grid item={true} xs={12}>
                                    <DeviceTableComponent/>
                                </Grid>
                            </Grid>
                        </Paper>
                    </TitleCard>
                </Grid>
            </Grid>
        </Grid>
    );
};

export default Devices;
