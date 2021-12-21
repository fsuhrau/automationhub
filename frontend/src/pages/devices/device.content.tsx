import { FC } from 'react';
import Paper from '@mui/material/Paper';
import DeviceTableComponent from '../../components/device-table.component';
import { IconButton, Typography } from '@mui/material';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Grid from '@mui/material/Grid';
import { useHistory } from 'react-router-dom';
import { Add } from '@mui/icons-material';

const Devices: FC = () => {
    const history = useHistory();

    function onManageDevices(): void {
        history.push('/web/devices/manager');
    }

    return (
        <Paper variant={'paper_content'} sx={{ maxWidth: 1200, margin: 'auto', overflow: 'hidden' }}>
            <AppBar
                position="static"
                color="default"
                elevation={0}
                sx={{ borderBottom: '1px solid rgba(0, 0, 0, 0.12)' }}
            >
                <Toolbar>
                    <Grid container={ true } spacing={ 2 } alignItems="center">
                        <Grid item={ true }>
                            <Typography variant={ 'h6' }>
                                Devices
                            </Typography>
                        </Grid>
                        <Grid item={ true } xs={ true }>
                        </Grid>
                        <Grid item={ true }>
                            <IconButton color="primary" size={'small'}
                                onClick={ (e) => {
                                } }>
                                <Add/>
                            </IconButton>
                        </Grid>
                    </Grid>
                </Toolbar>
            </AppBar>
            <DeviceTableComponent/>
        </Paper>
    );
};

export default Devices;
