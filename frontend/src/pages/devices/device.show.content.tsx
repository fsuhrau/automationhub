import React from 'react';
import Paper from '@mui/material/Paper';
import {
    Box,
    ButtonGroup,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Divider,
    Typography
} from '@mui/material';
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

    const {device} = props;
    const [expanded, setExpanded] = React.useState<string | false>(false);

    const handleChange = (panel: string) => (event: React.ChangeEvent<{}>, isExpanded: boolean) => {
        setExpanded(isExpanded ? panel : false);
    };

    const [open, setOpen] = React.useState(false);

    const handleClickOpen = () => {
        setOpen(true);
    };

    const handleClose = () => {
        setOpen(false);
    };

    const deleteAndClose = () => {
        deleteDevice(projectId as string, device.ID as number).then((result) =>
            navigate(-1),
        );
        handleClose()
    };

    return (
        <Paper sx={ {maxWidth: 1200, margin: 'auto', overflow: 'hidden'} }>
            <Dialog
                open={ open }
                onClose={ handleClose }
                aria-labelledby="alert-dialog-title"
                aria-describedby="alert-dialog-description"
            >
                <DialogTitle id="alert-dialog-title">
                    { "Delete Device?" }
                </DialogTitle>
                <DialogContent>
                    <DialogContentText id="alert-dialog-description">
                        You are going to delete the device are your sure?
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button onClick={ handleClose }>No</Button>
                    <Button onClick={ deleteAndClose } autoFocus>
                        Yes Delete it!
                    </Button>
                </DialogActions>
            </Dialog>
            <AppBar
                position="static"
                color="default"
                elevation={ 0 }
                sx={ {borderBottom: '1px solid rgba(0, 0, 0, 0.12)'} }
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
                                <Button color="secondary" onClick={ handleClickOpen }> Delete</Button>
                            </ButtonGroup>
                        </Grid>
                    </Grid>
                </Toolbar>
            </AppBar>
            <Box sx={ {p: 2, m: 2} }>
                <Grid container={ true } spacing={ 2 }>
                    <Grid item={ true } xs={ 12 } container={ true } spacing={ 1 }>
                        <Grid item={ true } xs={ 12 }>
                            <Typography variant={ 'h4' }>Device Infos</Typography>
                            <Divider/>
                        </Grid>
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
                            Alias:
                        </Grid>
                        <Grid item={ true } xs={ 10 }>
                            { device.Alias }
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
                            Acknowledged:
                        </Grid>
                        <Grid item={ true } xs={ 10 }>
                            { device.IsAcknowledged ? 'Yes' : 'No' }
                        </Grid>
                    </Grid>
                    <Grid item={ true } xs={ 12 } container={ true } spacing={ 1 }>
                        <Grid item={ true } xs={ 6 } container={ true } spacing={ 1 }>
                            <Grid item={ true } xs={ 12 }>
                                <Typography variant={ 'h4' }>Connection</Typography>
                                <Divider/>
                            </Grid>
                            <Grid item={ true } xs={ 2 }>
                                Type
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { device.ConnectionParameter && DeviceConnectionType[ device.ConnectionParameter.ConnectionType ] }
                            </Grid>
                            {
                                device.ConnectionParameter && device.ConnectionParameter.ConnectionType == DeviceConnectionType.Remote && device.ConnectionParameter.IP.length > 0 && device.ConnectionParameter.Port > 0 && (
                                    <Grid item={ true } xs={ 12 } container={ true }>
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
                                    </Grid>
                                )
                            }
                        </Grid>
                        <Grid item={ true } xs={ 6 } container={ true } spacing={ 1 }>
                            <Grid item={ true } xs={ 12 }>
                                <Typography variant={ 'h4' }>Parameter</Typography>
                                <Divider/>
                            </Grid>
                            { device.Parameter.map(value => (
                                    <Grid item={ true } container={ true } spacing={ 1 }>
                                        <Grid item={ true } xs={ 2 }>
                                            { value.Key }
                                        </Grid>
                                        <Grid item={ true } xs={ 10 }>
                                            { value.Value }
                                        </Grid>
                                    </Grid>
                                ),
                            ) }
                        </Grid>
                    </Grid>
                    <Grid item={ true } xs={ 6 } container={ true }>
                        <Grid item={ true } xs={ 12 }>
                            <Typography variant={ 'h4' }>Hardware</Typography>
                            <Divider/>
                        </Grid>
                        <Grid item={ true } xs={ 12 } container={ true } spacing={ 1 }>
                            <Grid item={ true } xs={ 4 }>
                                RAM:
                            </Grid>
                            <Grid item={ true } xs={ 8 }>
                                { device.RAM }
                            </Grid>
                            <Grid item={ true } xs={ 4 }>
                                SOC:
                            </Grid>
                            <Grid item={ true } xs={ 8 }>
                                { device.SOC }
                            </Grid>
                            <Grid item={ true } xs={ 4 }>
                                GPU:
                            </Grid>
                            <Grid item={ true } xs={ 8 }>
                                { device.GPU }
                            </Grid>
                            <Grid item={ true } xs={ 4 }>
                                ABI:
                            </Grid>
                            <Grid item={ true } xs={ 8 }>
                                { device.ABI }
                            </Grid>
                        </Grid>
                    </Grid>
                    <Grid item={ true } xs={ 6 } container={ true }>
                        <Grid item={ true } xs={ 12 }>
                            <Typography variant={ 'h4' }>Graphic</Typography>
                            <Divider/>
                        </Grid>
                        <Grid item={ true } container={ true } spacing={ 1 }>
                            <Grid item={ true } xs={ 4 }>
                                Display:
                            </Grid>
                            <Grid item={ true } xs={ 8 }>
                                { device.DisplaySize }
                            </Grid>
                            <Grid item={ true } xs={ 4 }>
                                DPI:
                            </Grid>
                            <Grid item={ true } xs={ 8 }>
                                { device.DPI }
                            </Grid>
                            <Grid item={ true } xs={ 4 }>
                                OpenGL Es Version
                            </Grid>
                            <Grid item={ true } xs={ 8 }>
                                { device.OpenGLESVersion }
                            </Grid>
                        </Grid>
                    </Grid>
                </Grid>
            </Box>
        </Paper>
    )
        ;
};

export default DeviceShowContent;
