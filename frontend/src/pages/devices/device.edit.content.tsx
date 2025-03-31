import React, {useState} from 'react';
import Paper from '@mui/material/Paper';
import {
    Box,
    Checkbox,
    FormControl,
    FormControlLabel,
    IconButton,
    MenuItem,
    Select,
    SelectChangeEvent,
    TextField,
    Typography,
} from '@mui/material';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Grid from "@mui/material/Grid2";
import Button from '@mui/material/Button';
import IDeviceData from '../../types/device';
import {updateDevice} from '../../services/device.service';
import {useNavigate} from 'react-router-dom';
import {DeviceConnectionType} from '../../types/device.connection.type.enum';
import IDeviceParameter from '../../types/device.parameter';
import {Add, Remove} from '@mui/icons-material';
import {DeviceType} from '../../types/device.type.enum';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {TitleCard} from "../../components/title.card.component";
import {useError} from "../../ErrorProvider";

const StringIsNumber = (value: any): boolean => !isNaN(Number(value));

function ToArray(en: any): Array<Object> {
    return Object.keys(en).filter(StringIsNumber).map(key => ({id: key, name: en[key]}));
}

function getConnectionType(): Array<{ id: DeviceConnectionType, name: string }> {
    return ToArray(DeviceConnectionType) as Array<{ id: DeviceConnectionType, name: string }>;
}

interface DeviceEditProps {
    device: IDeviceData
}

const DeviceEditContent: React.FC<DeviceEditProps> = props => {

    const {projectIdentifier} = useProjectContext()

    const navigate = useNavigate();
    const {setError} = useError()

    const {device} = props;

    const [state, setState] = useState<{
        alias: string,
        parameter: IDeviceParameter[],
        acknowledged: boolean,
        connectionType: DeviceConnectionType,
        ipAddress: string,
        port: number
    }>({
        alias: device.Alias,
        parameter: device.Parameter,
        acknowledged: device.IsAcknowledged,
        connectionType: device.ConnectionParameter ? device.ConnectionParameter.ConnectionType : DeviceConnectionType.USB,
        ipAddress: device.ConnectionParameter ? device.ConnectionParameter.IP : "",
        port: device.ConnectionParameter ? device.ConnectionParameter.Port : 0
    })

    const updateParameterKey = (index: number, value: string) => {
        setState(prevState => ({
            ...prevState,
            parameter: prevState.parameter.map((value1, index1) => index1 == index ? {...value1, Key: value} : value1)
        }))
    };
    const updateParameterValue = (index: number, value: string) => {
        setState(prevState => ({
            ...prevState,
            parameter: prevState.parameter.map((value1, index1) => index1 == index ? {...value1, Value: value} : value1)
        }))
    };

    const addParameter = (): void => {
        setState(prevState => ({...prevState, parameter: [...prevState.parameter, {Key: '', Value: ''}]}))
    };

    const removeParameter = (param: string) => {
        setState(prevState => ({
            ...prevState, parameter: [...prevState.parameter.filter(value => value.Key !== param)]
        }))
    };

    const connectionTypes = getConnectionType();

    const handleConnectionTypeChange = (event: SelectChangeEvent) => {
        setState(prevState => ({
            ...prevState, connectionType: +event.target.value as DeviceConnectionType
        }))
    };

    const saveChanges = () => {
        device.Parameter = state.parameter;
        device.IsAcknowledged = state.acknowledged;
        device.Alias = state.alias;
        device.ConnectionParameter = {
            ConnectionType: state.connectionType,
            IP: state.ipAddress,
            Port: state.port,
        };
        updateDevice(projectIdentifier, device, device.ID).then(response => {
            navigate(`/project/${projectIdentifier}/device/${device.ID}`);
        }).catch(ex => setError(ex));
    };

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
                                Device: {device.Name} ({device.DeviceIdentifier})
                            </Typography>
                        </Grid>
                        <Grid>
                        </Grid>
                        <Grid>
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
                            <TextField id="alias_edit" placeholder="Alias" variant="standard" value={state.alias}
                                       onChange={event => {
                                           setState(prevState => ({...prevState, alias: event.target.value}))
                                       }}/>
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
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            <FormControlLabel
                                control={<Checkbox checked={state.acknowledged}
                                                   onChange={(event, checked) => {
                                                       setState(prevState => ({
                                                           ...prevState, acknowledged: checked
                                                       }))
                                                   }}
                                                   name="ack"/>}
                                label="Acknowledged"
                            />
                        </Grid>
                    </Grid>
                </TitleCard>
                <TitleCard title={'Connection'}>
                    <Grid container={true} spacing={1}>
                        <Grid size={{xs: 12, md: 2}}>
                            Type
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            <FormControl sx={{m: 1, minWidth: 120}}>
                                <Select
                                    value={state.connectionType.toString()}
                                    onChange={handleConnectionTypeChange}
                                    displayEmpty
                                    variant={"standard"}
                                    inputProps={{'aria-label': 'Without label'}}
                                >
                                    <MenuItem value="">
                                        <em>None</em>
                                    </MenuItem>
                                    {connectionTypes.map((value: { id: DeviceConnectionType, name: string }) => (
                                        <MenuItem key={`device_connection_type_${value.id}`}
                                                  value={value.id}>{value.name}</MenuItem>
                                    ))}
                                </Select>
                            </FormControl>
                        </Grid>
                        {state.connectionType == DeviceConnectionType.Remote && (
                            <Grid size={{xs: 12, md: 12}} container={true}>
                                <Grid size={{xs: 12, md: 2}}>
                                    IP:
                                </Grid>
                                <Grid size={{xs: 12, md: 10}}>
                                    <TextField id="connection_address"
                                               placeholder="Address"
                                               variant="standard"
                                               size={"small"}
                                               value={state.ipAddress}
                                               fullWidth={true}
                                               onChange={event => {
                                                   setState(prevState => ({
                                                       ...prevState, ipAddress: event.target.value
                                                   }))
                                               }}
                                    />
                                </Grid>
                                <Grid size={{xs: 12, md: 2}}>
                                    Port:
                                </Grid>
                                <Grid size={{xs: 12, md: 10}}>
                                        <TextField id="connection_port"
                                                   placeholder="Port"
                                                   variant="standard"
                                                   size={"small"}
                                                   value={state.port}
                                                   fullWidth={true}
                                                   onChange={event => {
                                                       setState(prevState => ({
                                                           ...prevState, port: +event.target.value
                                                       }))
                                                   }}
                                        />
                                </Grid>
                            </Grid>
                        )}
                    </Grid>
                </TitleCard>
                <TitleCard title={'Parameter'}>
                    <Grid container={true} spacing={1}>
                        {state.parameter.map((value, index) => (
                                <Grid size={{xs: 12, md: 12}} container={true} spacing={1}>
                                    <Grid size={{xs: 12, md: 2}}>
                                        <TextField id={`update_parameter_key_${index}`} value={value.Key}
                                                   placeholder={"Key"} variant={"standard"}
                                                   onChange={event => updateParameterKey(index, event.target.value)}/>
                                    </Grid>
                                    <Grid size={{xs: 12, md: 2}}>
                                        <TextField id={`update_parameter_value_${index}`} value={value.Value}
                                                   placeholder={"Value"} variant={"standard"}
                                                   onChange={event => updateParameterValue(index, event.target.value)}/>
                                    </Grid>
                                    <Grid size={{xs: 12, md: 8}}>
                                        <IconButton color="secondary" aria-label="remove" component="span"
                                                    size={"small"}
                                                    onClick={() => removeParameter(value.Key)}><Remove/></IconButton>
                                    </Grid>
                                </Grid>
                            ),
                        )}
                        <Grid size={{xs: 12, md: 2}}>
                        </Grid>
                        <Grid size={{xs: 12, md: 2}}>
                        </Grid>
                        <Grid size={{xs: 12, md: 8}}>
                            <IconButton color="primary" aria-label="remove" component="span"
                                        onClick={addParameter}>
                                <Add/>
                            </IconButton>
                        </Grid>
                    </Grid>
                </TitleCard>
                <TitleCard title={'Hardware'}>
                    <Grid container={true} spacing={1}>
                        <Grid size={{xs: 12, md: 2}}>
                            RAM:
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            {device.RAM}
                        </Grid>
                        <Grid size={{xs: 12, md: 2}}>
                            SOC:
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            {device.SOC}
                        </Grid>
                        <Grid size={{xs: 12, md: 2}}>
                            GPU:
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            {device.GPU}
                        </Grid>
                        <Grid size={{xs: 12, md: 2}}>
                            ABI:
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            {device.ABI}
                        </Grid>
                    </Grid>
                </TitleCard>
                <TitleCard title={'Graphic'}>
                    <Grid container={true} spacing={1}>
                        <Grid size={{xs: 12, md: 2}}>
                            Display:
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            {device.DisplaySize}
                        </Grid>
                        <Grid size={{xs: 12, md: 2}}>
                            DPI:
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            {device.DPI}
                        </Grid>
                        <Grid size={{xs: 12, md: 2}}>
                            OpenGL Es Version
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            {device.OpenGLESVersion}
                        </Grid>
                    </Grid>
                </TitleCard>
                <Grid size={{xs: 12, md: 12}} container={true} justifyContent="flex-end">
                    <Button variant="contained" color="primary" size="small" onClick={saveChanges}>Save</Button>
                </Grid>
            </Box>
        </Paper>
    );
};

export default DeviceEditContent;
