import React from 'react';
import Paper from '@mui/material/Paper';
import { Box, ButtonGroup, Checkbox, Divider, FormControlLabel, Typography } from '@mui/material';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Grid from '@mui/material/Grid';
import Button from '@mui/material/Button';
import IDeviceData from '../../types/device';
import { deleteDevice } from '../../services/device.service';
import { useNavigate } from 'react-router-dom';
import { DeviceType } from '../../types/device.type.enum';
import { DeviceConnectionType } from '../../types/device.connection.type.enum';
import { useProjectContext } from "../../project/project.context";

interface DeviceShowProps {
    device: IDeviceData
}

const DeviceShowContent: React.FC<DeviceShowProps> = (props: DeviceShowProps) => {

    const {projectId} = useProjectContext();

    const navigate = useNavigate();

    const { device } = props;
    const [expanded, setExpanded] = React.useState<string | false>(false);

    const handleChange = (panel: string) => (event: React.ChangeEvent<{}>, isExpanded: boolean) => {
        setExpanded(isExpanded ? panel : false);
    };

    return (
        <Paper sx={ { maxWidth: 1200, margin: 'auto', overflow: 'hidden' } }>
            <AppBar
                position="static"
                color="default"
                elevation={ 0 }
                sx={ { borderBottom: '1px solid rgba(0, 0, 0, 0.12)' } }
            >
                <Toolbar>
                    <Grid container={ true } spacing={ 2 } alignItems="center">
                        <Grid item={ true }>
                            <Typography variant={ 'h6' }>
                                Device: { device.Name } ({ device.DeviceIdentifier })
                            </Typography>
                        </Grid>
                        <Grid item={ true } xs={ true }>
                        </Grid>
                        <Grid item={ true }>
                            <ButtonGroup variant="text" aria-label="text button group">
                                <Button href={ `${ device.ID }/edit` }>Edit</Button>
                                <Button color="secondary" onClick={ () => {
                                    deleteDevice(projectId as string, device.ID as number).then((result) =>
                                        navigate('/devices'),
                                    );
                                } }> Delete</Button>
                            </ButtonGroup>
                        </Grid>
                    </Grid>
                </Toolbar>
            </AppBar>
            <Box sx={ { p: 2, m: 2 } }>
                <Grid container={ true }>
                    <Grid item={ true } xs={ 12 }>
                        <Typography variant={ 'h1' }>Device Infos</Typography>
                        <Divider/>
                        <br/>
                        <Grid container={ true }>
                            <Grid item={ true } xs={ 2 }>
                                ID:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { device.ID }
                            </Grid>
                            <Grid item={ true } xs={ 2 }>
                                Name:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { device.Name }
                            </Grid>
                            <Grid item={ true } xs={ 2 }>
                                Model:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { device.HardwareModel }
                            </Grid>
                            <Grid item={ true } xs={ 2 }>
                                Identifier:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { device.DeviceIdentifier }
                            </Grid>
                            <Grid item={ true } xs={ 2 }>
                                Type:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { DeviceType[ device.DeviceType ] }
                            </Grid>
                            <Grid item={ true } xs={ 2 }>
                                Operation System:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { device.OS }<br/>
                                { device.OSVersion }
                            </Grid>
                            <Grid item={ true } xs={ 2 }>
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                <FormControlLabel
                                    control={ <Checkbox readOnly={ true } checked={ device.IsAcknowledged }
                                        name="ack"/> }
                                    label="Acknowledged"
                                />
                            </Grid>
                        </Grid>
                    </Grid>
                </Grid>
                <Grid container={ true }>
                    <Grid item={ true } xs={ 12 }>
                        <Typography variant={ 'h6' }>Connection</Typography>
                        <Divider/>
                        <br/>
                        <Grid container={ true }>
                            <Grid item={ true } xs={ 2 }>
                                Type
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { DeviceConnectionType[ device.ConnectionParameter.ConnectionType ] }
                            </Grid>
                            {
                                device.ConnectionParameter.ConnectionType == DeviceConnectionType.Remote && (<>
                                    <Grid item={ true } xs={ 2 }>
                                        IP:
                                    </Grid>
                                    <Grid item={ true } xs={ 10 }>
                                        { device.ConnectionParameter.IP }
                                    </Grid>
                                    <Grid item={ true } xs={ 2 }>
                                        Port:
                                    </Grid>
                                    <Grid item={ true } xs={ 10 }>
                                        { device.ConnectionParameter.Port }
                                    </Grid>
                                </>
                                )
                            }
                        </Grid>
                    </Grid>
                </Grid>
                <Grid container={ true }>
                    <Grid item={ true } xs={ 12 }>
                        <Typography variant={ 'h6' }>Parameter</Typography>
                        <Divider/>
                        <br/>
                        <Grid container={ true }>
                            { device.Parameter.map(value => (
                                <>
                                    <Grid item={ true } xs={ 2 }>
                                        { value.Key }
                                    </Grid>
                                    <Grid item={ true } xs={ 10 }>
                                        { value.Value }
                                    </Grid>
                                </>
                            ),
                            ) }
                        </Grid>
                    </Grid>
                </Grid>
                <br/>
                <Grid container={ true }>
                    <Grid item={ true } xs={ 12 }>
                        <Typography variant={ 'h1' }>Hardware</Typography>
                        <Divider/>
                        <br/>
                        <Grid container={ true }>
                            <Grid item={ true } xs={ 2 }>
                                RAM:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { device.RAM }
                            </Grid>
                            <Grid item={ true } xs={ 2 }>
                                SOC:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { device.SOC }
                            </Grid>
                            <Grid item={ true } xs={ 2 }>
                                GPU:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { device.GPU }
                            </Grid>
                            <Grid item={ true } xs={ 2 }>
                                ABI:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { device.ABI }
                            </Grid>
                        </Grid>
                    </Grid>
                </Grid>
                <br/>
                <Grid container={ true }>
                    <Grid item={ true } xs={ 12 }>
                        <Typography variant={ 'h1' }>Graphic</Typography>
                        <Divider/>
                        <br/>
                        <Grid container={ true }>
                            <Grid item={ true } xs={ 2 }>
                                Display:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { device.DisplaySize }
                            </Grid>
                            <Grid item={ true } xs={ 2 }>
                                DPI:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { device.DPI }
                            </Grid>
                            <Grid item={ true } xs={ 2 }>
                                OpenGL Es Version
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { device.OpenGLESVersion }
                            </Grid>
                        </Grid>
                    </Grid>
                </Grid>
                <br/>
            </Box>
        </Paper>
    );
};

export default DeviceShowContent;
