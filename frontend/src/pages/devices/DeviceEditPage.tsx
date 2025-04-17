import React from 'react';
import {Box, ButtonGroup, IconButton, SelectChangeEvent, TextField, Typography} from '@mui/material';
import Grid from "@mui/material/Grid";
import Button from '@mui/material/Button';
import IDeviceData from '../../types/device';
import {updateDevice} from '../../services/device.service';
import {useNavigate} from 'react-router-dom';
import {DeviceType} from '../../types/device.type.enum';
import {DeviceConnectionType} from '../../types/device.connection.type.enum';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {TitleCard} from "../../components/title.card.component";
import {useError} from "../../ErrorProvider";
import PlatformTypeIcon from "../../components/PlatformTypeIcon";
import {useHubState} from "../../hooks/HubStateProvider";
import IParameter from '../../types/device.parameter';
import {Add, Remove} from '@mui/icons-material';

const StringIsNumber = (value: any): boolean => !isNaN(Number(value));

function ToArray(en: any): Array<Object> {
    return Object.keys(en).filter(StringIsNumber).map(key => ({id: key, name: en[key]}));
}

function getConnectionType(): Array<{ id: DeviceConnectionType, name: string }> {
    return ToArray(DeviceConnectionType) as Array<{ id: DeviceConnectionType, name: string }>;
}

interface DeviceEditPageProps {
    device: IDeviceData
}

const DeviceEditPage: React.FC<DeviceEditPageProps> = (props: DeviceEditPageProps) => {

    const {device} = props;

    const navigate = useNavigate();
    const {projectIdentifier} = useProjectContext();
    const {state} = useHubState()
    const {setError} = useError()

    const node = state.nodes?.find(n => n.ID === device.NodeID);

    const [uiState, setUiState] = React.useState<{
        alias: string,
        customParameter: IParameter[],
        acknowledged: boolean,
        connectionType: DeviceConnectionType,
        ipAddress: string,
        port: number
    }>({
        alias: device.Alias,
        customParameter: device.CustomParameter,
        acknowledged: device.IsAcknowledged,
        connectionType: device.ConnectionParameter ? device.ConnectionParameter.ConnectionType : DeviceConnectionType.USB,
        ipAddress: device.ConnectionParameter ? device.ConnectionParameter.IP : "",
        port: device.ConnectionParameter ? device.ConnectionParameter.Port : 0
    })


    const updateParameterKey = (index: number, value: string) => {
        setUiState(prevState => ({
            ...prevState,
            customParameter: prevState.customParameter.map((value1, index1) => index1 == index ? {
                ...value1,
                Key: value
            } : value1)
        }))
    };
    const updateParameterValue = (index: number, value: string) => {
        setUiState(prevState => ({
            ...prevState,
            customParameter: prevState.customParameter.map((value1, index1) => index1 == index ? {
                ...value1,
                Value: value
            } : value1)
        }))
    };

    const addParameter = (): void => {
        setUiState(prevState => ({...prevState, customParameter: [...prevState.customParameter, {Key: '', Value: ''}]}))
    };

    const removeParameter = (param: string) => {
        setUiState(prevState => ({
            ...prevState, customParameter: [...prevState.customParameter.filter(value => value.Key !== param)]
        }))
    };

    const connectionTypes = getConnectionType();

    const handleConnectionTypeChange = (event: SelectChangeEvent) => {
        setUiState(prevState => ({
            ...prevState, connectionType: +event.target.value as DeviceConnectionType
        }))
    };

    const onSaveClick = () => {
        device.CustomParameter = uiState.customParameter;
        device.IsAcknowledged = uiState.acknowledged;
        device.Alias = uiState.alias;
        device.ConnectionParameter = {
            ConnectionType: uiState.connectionType,
            IP: uiState.ipAddress,
            Port: uiState.port,
        };
        updateDevice(projectIdentifier, device, device.ID).then(response => {
            navigateBack()
        }).catch(ex => setError(ex));
    };

    const navigateBack = () => {
        navigate(`/project/${projectIdentifier}/device/${device.ID}`);
    }

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <TitleCard titleElement={
                <Box sx={{display: 'flex', justifyContent: 'space-between', width: '100%'}}>
                    <Typography component="h2" variant="h6">
                        <PlatformTypeIcon
                            platformType={device.PlatformType}/> {`Device: ${device.Name} (${device.DeviceIdentifier})`}
                    </Typography>
                    <ButtonGroup variant="contained" aria-label="text button group">
                        <Button variant="contained" size={'small'} onClick={navigateBack}>Cancel</Button>
                        <Button color={'primary'} variant="contained" size={'small'} onClick={onSaveClick}>Save</Button>
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

                                <Grid size={{xs: 12, md: 2}}>
                                    Alias:
                                </Grid>
                                <Grid size={{xs: 12, md: 10}}>
                                    <TextField id={"alias_edit"} placeholder={"Alias"} variant={'outlined'}
                                               size={'small'} value={uiState.alias}
                                               onChange={event => {
                                                   setUiState(prevState => ({...prevState, alias: event.target.value}))
                                               }}/>
                                </Grid>

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
                        <TitleCard title={'Custom User Parameter'}>
                            <Grid container={true} spacing={1}>
                                {uiState.customParameter.map((value, index) => (
                                        <Grid size={{xs: 12, md: 12}} container={true} spacing={1}>
                                            <Grid size={{xs: 12, md: 2}}>
                                                <TextField id={`update_parameter_key_${index}`} value={value.Key}
                                                           placeholder={"Key"} variant={"outlined"} size={'small'} fullWidth={true}
                                                           onChange={event => updateParameterKey(index, event.target.value)}/>
                                            </Grid>
                                            <Grid size={{xs: 12, md: 2}}>
                                                <TextField id={`update_parameter_value_${index}`} value={value.Value} fullWidth={true}
                                                           placeholder={"Value"} variant={"outlined"} size={'small'}
                                                           onChange={event => updateParameterValue(index, event.target.value)}/>
                                            </Grid>
                                            <Grid size={{xs: 12, md: 8}}>
                                                <IconButton color={"error"} aria-label={"remove"} component={"span"}
                                                            size={"small"}
                                                            onClick={() => removeParameter(value.Key)}>
                                                    <Remove/>
                                                </IconButton>
                                            </Grid>
                                        </Grid>
                                    ),
                                )}
                                <Grid size={{xs: 12, md: 2}}>
                                </Grid>
                                <Grid size={{xs: 12, md: 8}}>
                                    <IconButton color="primary" aria-label="remove" component="span" size={'small'}
                                                onClick={addParameter}>
                                        <Add/>
                                    </IconButton>
                                </Grid>
                            </Grid>
                        </TitleCard>
                    </Grid>
                </Grid>
            </TitleCard>
        </Box>
    );
};

export default DeviceEditPage;
