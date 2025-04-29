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

    const node = state.nodes?.find(n => n.id === device.nodeId);

    const [uiState, setUiState] = React.useState<{
        alias: string,
        customParameter: IParameter[],
        acknowledged: boolean,
        connectionType: DeviceConnectionType,
        ipAddress: string,
        port: number
        currentKey: string
        currentValue: string
    }>({
        alias: device.alias,
        customParameter: device.customParameter,
        acknowledged: device.isAcknowledged,
        connectionType: device.connectionParameter ? device.connectionParameter.connectionType : DeviceConnectionType.USB,
        ipAddress: device.connectionParameter ? device.connectionParameter.ip : "",
        port: device.connectionParameter ? device.connectionParameter.port : 0,
        currentKey: '',
        currentValue: '',
    })

    const updateParameterKey = (index: number, value: string) => {
        setUiState(prevState => ({
            ...prevState,
            customParameter: [...prevState.customParameter.map((value1, index1) => index1 == index ? {
                ...value1,
                key: value
            } : value1)]
        }))
    };
    const updateParameterValue = (index: number, value: string) => {
        setUiState(prevState => ({
            ...prevState,
            customParameter: [...prevState.customParameter.map((value1, index1) => index1 == index ? {
                ...value1,
                value: value
            } : value1)]
        }))
    };

    const addParameter = (): void => {
        setUiState(prevState => ({...prevState, customParameter: [...prevState.customParameter, {key: '', value: ''}]}))
    };

    const removeParameter = (param: string) => {
        setUiState(prevState => ({
            ...prevState, customParameter: [...prevState.customParameter.filter(value => value.key !== param)]
        }))
    };

    const connectionTypes = getConnectionType();

    const handleConnectionTypeChange = (event: SelectChangeEvent) => {
        setUiState(prevState => ({
            ...prevState, connectionType: +event.target.value as DeviceConnectionType
        }))
    };

    const onSaveClick = () => {
        device.customParameter = uiState.customParameter;
        device.isAcknowledged = uiState.acknowledged;
        device.alias = uiState.alias;
        device.connectionParameter = {
            connectionType: uiState.connectionType,
            ip: uiState.ipAddress,
            port: uiState.port,
        };
        updateDevice(projectIdentifier, device, device.id).then(response => {
            navigateBack()
        }).catch(ex => setError(ex));
    };

    const navigateBack = () => {
        navigate(`/project/${projectIdentifier}/device/${device.id}`);
    }

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <TitleCard titleElement={
                <Box sx={{display: 'flex', justifyContent: 'space-between', width: '100%'}}>
                    <Typography component="h2" variant="h6">
                        <PlatformTypeIcon
                            platformType={device.platformType}/> {`Device: ${device.name} (${device.deviceIdentifier})`}
                    </Typography>
                    <ButtonGroup variant="contained" aria-label="text button group">
                        <Button variant="outlined" size={'small'} onClick={navigateBack}>Cancel</Button>
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
                                    {device.id}
                                </Grid>
                                <Grid size={{xs: 12, md: 2}}>
                                    Name:
                                </Grid>
                                <Grid size={{xs: 12, md: 10}}>
                                    {device.name}
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
                                {device.deviceParameter.map(d => (
                                    <>
                                        <Grid size={{xs: 12, md: 2}}>
                                            {d.key}
                                        </Grid>
                                        <Grid size={{xs: 12, md: 10}}>
                                            {d.value}
                                        </Grid>
                                    </>
                                ))}
                            </Grid>
                        </TitleCard>
                    </Grid>
                    <Grid size={12}>
                        <TitleCard title={'Custom User parameter'}>
                            <Grid container={true} spacing={1}>
                                {uiState.customParameter.map((value, index) => (
                                        <Grid size={{xs: 12, md: 12}} container={true} spacing={1}>
                                            <Grid size={{xs: 12, md: 2}}>
                                                <TextField id={`update_parameter_key_${index}`} value={value.key}
                                                           placeholder={"Key"} variant={"outlined"} size={'small'} fullWidth={true}
                                                           onChange={event => updateParameterKey(index, event.target.value)}/>
                                            </Grid>
                                            <Grid size={{xs: 12, md: 2}}>
                                                <TextField id={`update_parameter_value_${index}`} value={value.value} fullWidth={true}
                                                           placeholder={"Value"} variant={"outlined"} size={'small'}
                                                           onChange={event => updateParameterValue(index, event.target.value)}/>
                                            </Grid>
                                            <Grid size={{xs: 12, md: 8}}>
                                                <IconButton color={"error"} aria-label={"remove"} component={"span"}
                                                            size={"small"}
                                                            onClick={() => removeParameter(value.key)}>
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
