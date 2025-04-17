import React from 'react';
import {
    Box,
    ButtonGroup,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Typography
} from '@mui/material';
import Grid from "@mui/material/Grid";
import Button from '@mui/material/Button';
import IDeviceData from '../../types/device';
import {deleteDevice} from '../../services/device.service';
import {useNavigate} from 'react-router-dom';
import {DeviceType} from '../../types/device.type.enum';
import {DeviceConnectionType} from '../../types/device.connection.type.enum';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {TitleCard} from "../../components/title.card.component";
import {useError} from "../../ErrorProvider";
import PlatformTypeIcon from "../../components/PlatformTypeIcon";
import {useHubState} from "../../hooks/HubStateProvider";

interface DeviceShowPageProps {
    device: IDeviceData
}

const DeviceShowPage: React.FC<DeviceShowPageProps> = (props: DeviceShowPageProps) => {

    const {projectIdentifier} = useProjectContext();

    const {state} = useHubState()

    const navigate = useNavigate();
    const {setError} = useError()

    const {device} = props;

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
        ).catch(ex => setError(ex));
        handleClose()
    };

    const node = state.nodes?.find(n => n.ID === device.NodeID);

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
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

            <TitleCard titleElement={
                <Box sx={{display: 'flex', justifyContent: 'space-between', width: '100%'}}>
                    <Typography component="h2" variant="h6">
                        <PlatformTypeIcon
                            platformType={device.PlatformType}/> {`Device: ${device.Name} (${device.DeviceIdentifier})`}
                    </Typography>
                    <ButtonGroup variant="contained" aria-label="text button group">
                        <Button variant="contained" size={'small'} onClick={() => navigate('edit')}>Edit</Button>
                        <Button variant="contained" color="error" size={'small'}
                                onClick={handleClickOpen}> Delete</Button>
                    </ButtonGroup>
                </Box>}>
                <Grid container={true} spacing={2}>
                    <Grid size={12}>
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
                                {device.Alias.length > 0 && <>
                                    <Grid size={{xs: 12, md: 2}}>
                                        Alias:
                                    </Grid>
                                    <Grid size={{xs: 12, md: 10}}>
                                        {device.Alias}
                                    </Grid>
                                </>}
                                <Grid size={{xs: 12, md: 2}}>
                                    Operation System:
                                </Grid>
                                <Grid size={{xs: 12, md: 10}}>
                                    {device.OS} {device.OSVersion}
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
                                {
                                    /*
                                    <Grid size={{xs: 12, md: 2}}>
                                    Acknowledged:
                                </Grid>
                                <Grid size={{xs: 12, md: 10}}>
                                    {device.IsAcknowledged ? 'Yes' : 'No'}
                                </Grid>
                                     */
                                }

                            </Grid>
                        </TitleCard>
                    </Grid>
                    <Grid size={12}>
                        {
                            device.ConnectionParameter && <TitleCard title={'Connection'}>
                                <Grid container={true} spacing={1}>
                                    <Grid size={{xs: 12, md: 2}}>
                                        Type
                                    </Grid>
                                    <Grid size={{xs: 12, md: 10}}>
                                        {DeviceConnectionType[device.ConnectionParameter.ConnectionType]}
                                    </Grid>
                                    {
                                        device.ConnectionParameter.ConnectionType == DeviceConnectionType.HubNode && <>
                                            <Grid size={{xs: 12, md: 2}}>
                                                Node
                                            </Grid>
                                            <Grid size={{xs: 12, md: 10}}>
                                                {node?.Name}
                                            </Grid>
                                        </>
                                    }
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
                        }
                    </Grid>
                    <Grid size={12}>
                        <TitleCard title={'Device Parameter'}>
                            <Grid container={true} spacing={1}>
                                {device.DeviceParameter.map(d => (
                                    <>
                                        <Grid size={{xs: 12, md: 2}}>
                                            {d.Key}
                                        </Grid>
                                        <Grid size={{xs: 12, md: 10}}>
                                            {d.Value}
                                        </Grid>
                                    </>
                                ))}
                            </Grid>
                        </TitleCard>
                    </Grid>
                    <Grid size={12}>
                        {device.CustomParameter.length > 0 && <TitleCard title={'Custom User Parameter'}>
                            <Grid container={true} spacing={1}>
                                {device.CustomParameter.map(value => (
                                        <>
                                            <Grid size={{xs: 12, md: 2}}>
                                                {value.Key}
                                            </Grid>
                                            <Grid size={{xs: 12, md: 10}}>
                                                {value.Value}
                                            </Grid>
                                        </>
                                    ),
                                )}
                            </Grid>
                        </TitleCard>}
                    </Grid>
                </Grid>
            </TitleCard>
        </Box>
    )
        ;
};

export default DeviceShowPage;
