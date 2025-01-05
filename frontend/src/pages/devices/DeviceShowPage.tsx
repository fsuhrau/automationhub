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
import Grid from "@mui/material/Grid2";
import Button from '@mui/material/Button';
import IDeviceData from '../../types/device';
import {deleteDevice} from '../../services/device.service';
import {useNavigate} from 'react-router-dom';
import {DeviceType} from '../../types/device.type.enum';
import {DeviceConnectionType} from '../../types/device.connection.type.enum';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {TitleCard} from "../../components/title.card.component";

interface DeviceShowPageProps {
    device: IDeviceData
}

const DeviceShowPage: React.FC<DeviceShowPageProps> = (props: DeviceShowPageProps) => {

    const {projectIdentifier} = useProjectContext();

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
        deleteDevice(projectIdentifier, device.ID as number).then((result) =>
            navigate(-1),
        );
        handleClose()
    };

    return (
        <Paper sx={{maxWidth: 1200, margin: 'auto', overflow: 'hidden'}}>
            <Dialog
                open={open}
                onClose={handleClose}
                aria-labelledby="alert-dialog-title"
                aria-describedby="alert-dialog-description"
            >
                <DialogTitle id="alert-dialog-title">
                    {"Delete Device?"}
                </DialogTitle>
                <DialogContent>
                    <DialogContentText id="alert-dialog-description">
                        You are going to delete the device are your sure?
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button onClick={handleClose}>No</Button>
                    <Button onClick={deleteAndClose} autoFocus>
                        Yes Delete it!
                    </Button>
                </DialogActions>
            </Dialog>
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
                                Device: {device.Name} ({device.DeviceIdentifier})
                            </Typography>
                        </Grid>
                        <Grid>
                        </Grid>
                        <Grid>
                            <ButtonGroup variant="text" aria-label="text button group">
                                <Button onClick={() => navigate('edit')}>Edit</Button>
                                <Button color="secondary" onClick={handleClickOpen}> Delete</Button>
                            </ButtonGroup>
                        </Grid>
                    </Grid>
                </Toolbar>
            </AppBar>
            <Box sx={{p: 2, m: 2}}>
                <TitleCard title={'Device Infos'}>
                    <Grid container={true} spacing={1}>
                        <Grid size={{xs: 12, md: 2}}>
                            ID:
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            {device.ID}
                        </Grid>
                        <Grid size={{xs: 12, md: 2}}>
                            Name:
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            {device.Name}
                        </Grid>
                        <Grid size={{xs: 12, md: 2}}>
                            Alias:
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            {device.Alias}
                        </Grid>
                        <Grid size={{xs: 12, md: 2}}>
                            Model:
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            {device.HardwareModel}
                        </Grid>
                        <Grid size={{xs: 12, md: 2}}>
                            Identifier:
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            {device.DeviceIdentifier}
                        </Grid>
                        <Grid size={{xs: 12, md: 2}}>
                            Type:
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            {DeviceType[device.DeviceType]}
                        </Grid>
                        <Grid size={{xs: 12, md: 2}}>
                            Operation System:
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            {device.OS}<br/>
                            {device.OSVersion}
                        </Grid>
                        <Grid size={{xs: 12, md: 2}}>
                            Acknowledged:
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            {device.IsAcknowledged ? 'Yes' : 'No'}
                        </Grid>
                    </Grid>
                </TitleCard>
                <TitleCard title={'Connection'}>
                    <Grid container={true} spacing={1}>
                        <Grid size={{xs: 12, md: 2}}>
                            Type
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            {device.ConnectionParameter && DeviceConnectionType[device.ConnectionParameter.ConnectionType]}
                        </Grid>
                        {
                            device.ConnectionParameter && device.ConnectionParameter.ConnectionType == DeviceConnectionType.Remote && device.ConnectionParameter.IP.length > 0 && device.ConnectionParameter.Port > 0 && (
                                <Grid size={{xs: 12, md: 12}} container={true}>
                                    <Grid size={{xs: 12, md: 2}}>
                                        IP:
                                    </Grid>
                                    <Grid size={{xs: 12, md: 10}}>
                                        {device.ConnectionParameter.IP}
                                    </Grid>
                                    <Grid size={{xs: 12, md: 2}}>
                                        Port:
                                    </Grid>
                                    <Grid size={{xs: 12, md: 10}}>
                                        {device.ConnectionParameter.Port}
                                    </Grid>
                                </Grid>
                            )
                        }
                    </Grid>
                </TitleCard>
                <TitleCard title={'Parameter'}>
                    <Grid container={true} spacing={1}>
                        {device.Parameter.map(value => (
                                <Grid container={true} spacing={1}>
                                    <Grid size={{xs: 12, md: 2}}>
                                        {value.Key}
                                    </Grid>
                                    <Grid size={{xs: 12, md: 10}}>
                                        {value.Value}
                                    </Grid>
                                </Grid>
                            ),
                        )}
                    </Grid>
                </TitleCard>
                <TitleCard title={'Hardware'}>
                    <Grid container={true} spacing={1}>
                        <Grid size={{xs: 12, md: 4}}>
                            RAM:
                        </Grid>
                        <Grid size={{xs: 12, md: 8}}>
                            {device.RAM}
                        </Grid>
                        <Grid size={{xs: 12, md: 4}}>
                            SOC:
                        </Grid>
                        <Grid size={{xs: 12, md: 8}}>
                            {device.SOC}
                        </Grid>
                        <Grid size={{xs: 12, md: 4}}>
                            GPU:
                        </Grid>
                        <Grid size={{xs: 12, md: 8}}>
                            {device.GPU}
                        </Grid>
                        <Grid size={{xs: 12, md: 4}}>
                            ABI:
                        </Grid>
                        <Grid size={{xs: 12, md: 8}}>
                            {device.ABI}
                        </Grid>
                    </Grid>
                </TitleCard>
                <TitleCard title={'Graphic'}>
                    <Grid container={true} spacing={1}>
                        <Grid size={{xs: 12, md: 4}}>
                            Display:
                        </Grid>
                        <Grid size={{xs: 12, md: 8}}>
                            {device.DisplaySize}
                        </Grid>
                        <Grid size={{xs: 12, md: 4}}>
                            DPI:
                        </Grid>
                        <Grid size={{xs: 12, md: 8}}>
                            {device.DPI}
                        </Grid>
                        <Grid size={{xs: 12, md: 4}}>
                            OpenGL Es Version
                        </Grid>
                        <Grid size={{xs: 12, md: 8}}>
                            {device.OpenGLESVersion}
                        </Grid>
                    </Grid>
                </TitleCard>
            </Box>
        </Paper>
    );
};

export default DeviceShowPage;
