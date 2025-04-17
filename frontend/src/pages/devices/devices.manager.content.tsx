import React, {useEffect, useState} from 'react';
import Paper from '@mui/material/Paper';
import {Button, IconButton, Typography} from '@mui/material';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Grid from '@mui/material/Grid';
import IDeviceData from '../../types/device';
import {getAllDevices, postUnlockDevice} from '../../services/device.service';
import TableRow from '@mui/material/TableRow';
import TableCell from '@mui/material/TableCell';
import {Add, ArrowForward} from '@mui/icons-material';
import TableContainer from '@mui/material/TableContainer';
import Table from '@mui/material/Table';
import TableHead from '@mui/material/TableHead';
import TableBody from '@mui/material/TableBody';
import {useNavigate} from 'react-router-dom';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {useError} from "../../ErrorProvider";
import {deviceState, DeviceStateType} from "../../types/deviceStateType";

const DevicesManagerContent: React.FC = () => {

    const {projectIdentifier} = useProjectContext();

    const navigate = useNavigate();
    const {setError} = useError()

    const [devices, setDevices] = useState<IDeviceData[]>([]);

    function openDetails(id: number): void {
        navigate(`/device/${id}`);
    }

    useEffect(() => {
        getAllDevices(projectIdentifier).then(response => {
            setDevices(response.data);
        }).catch(e => {
            setError(e);
        });
    }, [projectIdentifier]);

    const unlockDevice = (deviceId: number) => {
        postUnlockDevice(projectIdentifier, deviceId).then(response => {
            setDevices(devices.map(d => d.ID === deviceId ? response.data : d) as IDeviceData[]);
        }).catch(ex => setError(ex));
    }


    return (
        <Paper sx={{maxWidth: 1200, margin: 'auto', overflow: 'hidden'}}>
            <AppBar
                position="static"
                color="default"
                elevation={0}
                sx={{borderBottom: '1px solid rgba(0, 0, 0, 0.12)'}}
            >
                <Toolbar>
                    <Grid container={true} spacing={2} alignItems="center">
                        <Grid>
                            <Typography variant={'h6'}>
                                Device Manager
                            </Typography>
                        </Grid>
                        <Grid>
                        </Grid>
                        <Grid>
                            <IconButton color="primary" size={'small'}
                                        onClick={(e) => {

                                        }}>
                                <Add/>
                            </IconButton>

                        </Grid>
                    </Grid>
                </Toolbar>
            </AppBar>
            <TableContainer component={Paper}>
                <Table sx={{maxWidth: 1200, margin: 'auto', overflow: 'hidden'}} size="small"
                       aria-label="a dense table">
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
                        {devices.map((device) => <TableRow key={`device_table_row_${device.ID}`}>
                            <TableCell component="th" scope="row">
                                {device.Alias.length > 0 ? device.Alias : device.Name}
                            </TableCell>
                            <TableCell>
                            </TableCell>
                            <TableCell>
                            </TableCell>
                            <TableCell>
                            </TableCell>
                            <TableCell>
                                {deviceState(device.Status)}
                                {device.Status === DeviceStateType.Locked &&
                                    <Button onClick={() => unlockDevice(device.ID)}>Unlock</Button>}
                            </TableCell>
                            <TableCell align="right">
                                <IconButton color="primary" size={'small'}
                                            onClick={(e) => {
                                                openDetails(device.ID);
                                            }}>
                                    <ArrowForward/>
                                </IconButton>
                            </TableCell>
                        </TableRow>)}
                    </TableBody>
                </Table>
            </TableContainer>
        </Paper>
    );
};

export default DevicesManagerContent;
