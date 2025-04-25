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
        deleteDevice(projectIdentifier, device.id as number).then((result) =>
            navigate(-1),
        ).catch(ex => setError(ex));
        handleClose()
    };

    const node = state.nodes?.find(n => n.id === device.nodeId);

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <Dialog
                open={open}
                onClose={handleClose}
                aria-labelledby="alert-dialog-title"
                aria-describedby="alert-dialog-description"
            >
                <DialogTitle id="alert-dialog-title">
                    {"Delete device?"}
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
                            platformType={device.platformType}/> {`Device: ${device.name} (${device.deviceIdentifier})`}
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
                                    {device.id}
                                </Grid>
                                <Grid size={{xs: 12, md: 2}}>
                                    Name:
                                </Grid>
                                <Grid size={{xs: 12, md: 10}}>
                                    {device.name}
                                </Grid>
                                {device.alias.length > 0 && <>
                                    <Grid size={{xs: 12, md: 2}}>
                                        Alias:
                                    </Grid>
                                    <Grid size={{xs: 12, md: 10}}>
                                        {device.alias}
                                    </Grid>
                                </>}
                                <Grid size={{xs: 12, md: 2}}>
                                    Operation System:
                                </Grid>
                                <Grid size={{xs: 12, md: 10}}>
                                    {device.os} {device.osVersion}
                                </Grid>
                                <Grid size={{xs: 12, md: 2}}>
                                    Identifier:
                                </Grid>
                                <Grid size={{xs: 12, md: 10}}>
                                    {device.deviceIdentifier}
                                </Grid>
                                <Grid size={{xs: 12, md: 2}}>
                                    Type:
                                </Grid>
                                <Grid size={{xs: 12, md: 10}}>
                                    {DeviceType[device.deviceType]}
                                </Grid>
                                {
                                    /*
                                    <Grid size={{xs: 12, md: 2}}>
                                    Acknowledged:
                                </Grid>
                                <Grid size={{xs: 12, md: 10}}>
                                    {device.isAcknowledged ? 'Yes' : 'No'}
                                </Grid>
                                     */
                                }

                            </Grid>
                        </TitleCard>
                    </Grid>
                    <Grid size={12}>
                        {
                            device.connectionParameter && <TitleCard title={'Connection'}>
                                <Grid container={true} spacing={1}>
                                    <Grid size={{xs: 12, md: 2}}>
                                        Type
                                    </Grid>
                                    <Grid size={{xs: 12, md: 10}}>
                                        {DeviceConnectionType[device.connectionParameter.connectionType]}
                                    </Grid>
                                    {
                                        device.connectionParameter.connectionType == DeviceConnectionType.HubNode && <>
                                            <Grid size={{xs: 12, md: 2}}>
                                                Node
                                            </Grid>
                                            <Grid size={{xs: 12, md: 10}}>
                                                {node?.name}
                                            </Grid>
                                        </>
                                    }
                                    {
                                        device.connectionParameter && device.connectionParameter.connectionType == DeviceConnectionType.Remote && device.connectionParameter.ip.length > 0 && device.connectionParameter.port > 0 && (
                                            <Grid size={{xs: 12, md: 12}} container={true}>
                                                <Grid size={{xs: 12, md: 2}}>
                                                    IP:
                                                </Grid>
                                                <Grid size={{xs: 12, md: 10}}>
                                                    {device.connectionParameter.ip}
                                                </Grid>
                                                <Grid size={{xs: 12, md: 2}}>
                                                    Port:
                                                </Grid>
                                                <Grid size={{xs: 12, md: 10}}>
                                                    {device.connectionParameter.port}
                                                </Grid>
                                            </Grid>
                                        )
                                    }
                                </Grid>
                            </TitleCard>
                        }
                    </Grid>
                    <Grid size={12}>
                        <TitleCard title={'Device parameter'}>
                            <Grid container={true} spacing={1}>
                                {device.deviceParameter.map((d, i) => (
                                    <Grid key={`device_parameter_${i}`} container={true} size={12}>
                                        <Grid size={{xs: 12, md: 2}}>
                                            {d.key}
                                        </Grid>
                                        <Grid size={{xs: 12, md: 10}}>
                                            {d.value}
                                        </Grid>
                                    </Grid>)
                                )}
                            </Grid>
                        </TitleCard>
                    </Grid>
                    <Grid size={12}>
                        {device.customParameter.length > 0 && <TitleCard title={'Custom User parameter'}>
                            <Grid container={true} spacing={1}>
                                {device.customParameter.map((value, i) => (
                                        <Grid key={`custom_parameter_${i}`} container={true} size={12}>
                                            <Grid size={{xs: 12, md: 2}}>
                                                {value.key}
                                            </Grid>
                                            <Grid size={{xs: 12, md: 10}}>
                                                {value.value}
                                            </Grid>
                                        </Grid>
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
