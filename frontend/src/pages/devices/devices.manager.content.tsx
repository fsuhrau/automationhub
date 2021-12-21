import { FC, useEffect, useState } from 'react';
import Paper from '@mui/material/Paper';
import { IconButton, Typography } from '@mui/material';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Grid from '@mui/material/Grid';
import IDeviceData from '../../types/device';
import { getAllDevices } from '../../services/device.service';
import TableRow from '@mui/material/TableRow';
import TableCell from '@mui/material/TableCell';
import { Add, ArrowForward } from '@mui/icons-material';
import TableContainer from '@mui/material/TableContainer';
import Table from '@mui/material/Table';
import TableHead from '@mui/material/TableHead';
import TableBody from '@mui/material/TableBody';
import { useHistory } from 'react-router-dom';

const DevicesManagerContent: FC = () => {

    const history = useHistory();

    const [devices, setDevices] = useState<IDeviceData[]>([]);

    function openDetails(id: number): void {
        history.push(`/web/device/${ id }`);
    }

    useEffect(() => {
        getAllDevices().then(response => {
            setDevices(response.data);
        }).catch(e => {
            console.log(e);
        });
    }, []);

    return (
        <Paper sx={{ maxWidth: 1200, margin: 'auto', overflow: 'hidden' }}>
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
                                Device Manager
                            </Typography>
                        </Grid>
                        <Grid item={ true } xs={ true }>
                        </Grid>
                        <Grid item={ true }>
                            <IconButton color="primary" size={ 'small' }
                                onClick={ (e) => {

                                } }>
                                <Add/>
                            </IconButton>

                        </Grid>
                    </Grid>
                </Toolbar>
            </AppBar>
            <TableContainer component={ Paper }>
                <Table sx={{ maxWidth: 1200, margin: 'auto', overflow: 'hidden' }} size="small" aria-label="a dense table">
                    <TableHead>
                        <TableRow>
                            <TableCell>Name</TableCell>
                            <TableCell>Model</TableCell>
                            <TableCell>RAM</TableCell>
                            <TableCell>SOC</TableCell>
                            <TableCell>Status</TableCell>
                            <TableCell align="right"></TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        { devices.map((device) => <TableRow key={ device.ID }>
                            <TableCell component="th" scope="row">
                                { device.Name }
                            </TableCell>
                            <TableCell>
                                { device.HardwareModel }
                            </TableCell>
                            <TableCell>
                                { device.RAM }
                            </TableCell>
                            <TableCell>
                                { device.SOC }
                            </TableCell>
                            <TableCell>
                                { device.Status }
                            </TableCell>
                            <TableCell align="right">
                                <IconButton color="primary" size={ 'small' }
                                    onClick={ (e) => {
                                        openDetails(device.ID);
                                    } }>
                                    <ArrowForward/>
                                </IconButton>
                            </TableCell>
                        </TableRow>) }
                    </TableBody>
                </Table>
            </TableContainer>
        </Paper>
    );
};

export default DevicesManagerContent;
