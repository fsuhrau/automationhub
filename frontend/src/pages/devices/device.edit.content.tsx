import React, { useState } from 'react';
import Paper from '@mui/material/Paper';
import {
    Box,
    Checkbox,
    Divider,
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
import Grid from '@mui/material/Grid';
import Button from '@mui/material/Button';
import IDeviceData from '../../types/device';
import { updateDevice } from '../../services/device.service';
import { useNavigate } from 'react-router-dom';
import { DeviceConnectionType } from '../../types/device.connection.type.enum';
import IDeviceParameter from '../../types/device.parameter';
import { Add, Remove } from '@mui/icons-material';
import { DeviceType } from '../../types/device.type.enum';
import {useProjectContext} from "../../hooks/ProjectProvider";

const StringIsNumber = (value: any): boolean => !isNaN(Number(value));

function ToArray(en: any): Array<Object> {
    return Object.keys(en).filter(StringIsNumber).map(key => ({id: key, name: en[ key ]}));
}

function getConnectionType(): Array<{ id: DeviceConnectionType, name: string }> {
    return ToArray(DeviceConnectionType) as Array<{ id: DeviceConnectionType, name: string }>;
}

interface DeviceEditProps {
    device: IDeviceData
}

const DeviceEditContent: React.FC<DeviceEditProps> = props => {

    const {projectId} = useProjectContext()

    const navigate = useNavigate();

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
        updateDevice(projectId as string, device, device.ID).then(response => {
            navigate(`/project/${ projectId }/device/${ device.ID }`);
        });
    };

    return (
        <Paper sx={ {maxWidth: 1200, margin: 'auto', overflow: 'hidden'} }>
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
                            <TextField id="alias_edit" label="Alias" variant="standard" value={ state.alias }
                                       onChange={ event => {
                                           setState(prevState => ({...prevState, alias: event.target.value}))
                                       } }/>
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
                                control={ <Checkbox checked={ state.acknowledged }
                                                    onChange={ (event, checked) => {
                                                        setState(prevState => ({
                                                            ...prevState, acknowledged: checked
                                                        }))
                                                    } }
                                                    name="ack"/> }
                                label="Acknowledged"
                            />
                        </Grid>
                    </Grid>
                    <Grid item={ true } xs={ 12 } container={ true } spacing={ 1 }>
                        <Grid item={ true } xs={ 12 }>
                            <Typography variant={ 'h6' }>Connection</Typography>
                            <Divider/>
                        </Grid>
                        <Grid item={ true } xs={ 2 }>
                            Type
                        </Grid>
                        <Grid item={ true } xs={ 10 }>
                            <FormControl sx={ {m: 1, minWidth: 120} }>
                                <Select
                                    value={ state.connectionType.toString() }
                                    onChange={ handleConnectionTypeChange }
                                    displayEmpty
                                    variant={ "standard" }
                                    inputProps={ {'aria-label': 'Without label'} }
                                >
                                    <MenuItem value="">
                                        <em>None</em>
                                    </MenuItem>
                                    { connectionTypes.map((value: { id: DeviceConnectionType, name: string }) => (
                                        <MenuItem key={ `device_connection_type_${ value.id }` }
                                                  value={ value.id }>{ value.name }</MenuItem>
                                    )) }
                                </Select>
                            </FormControl>
                        </Grid>
                        { state.connectionType == DeviceConnectionType.Remote && (
                            <Grid item={ true } xs={ 12 } container={ true }>
                                <Grid item={ true } xs={ 2 }>
                                    IP:
                                </Grid>
                                <Grid item={ true } xs={ 10 }>
                                    <FormControl>
                                        <TextField id="connection_address"
                                                   label="Address"
                                                   variant="standard"
                                                   size={"small"}
                                                   value={ state.ipAddress }
                                                   fullWidth={ true }
                                                   onChange={ event => {
                                                       setState(prevState => ({
                                                           ...prevState, ipAddress: event.target.value
                                                       }))
                                                   } }
                                        />
                                    </FormControl>
                                </Grid>
                                <Grid item={ true } xs={ 2 }>
                                    Port:
                                </Grid>
                                <Grid item={ true } xs={ 10 }>
                                    <FormControl>
                                        <TextField id="connection_port"
                                                   label="Port"
                                                   variant="standard"
                                                   size={"small"}
                                                   value={ state.port }
                                                   fullWidth={ true }
                                                   onChange={ event => {
                                                       setState(prevState => ({
                                                           ...prevState, port: +event.target.value
                                                       }))
                                                   } }
                                        />
                                    </FormControl>
                                </Grid>
                            </Grid>
                        ) }
                    </Grid>
                    <Grid item={ true } xs={ 12 } container={ true } spacing={ 1 }>
                        <Grid item={ true } xs={ 12 }>
                            <Typography variant={ 'h6' }>Parameter</Typography>
                            <Divider/>
                        </Grid>
                        { state.parameter.map((value, index) => (
                                <Grid item={ true } xs={ 12 } container={ true } spacing={ 1 }>
                                    <Grid item={ true } xs={ 2 }>
                                        <TextField id={ `update_parameter_key_${ index }` } value={ value.Key }
                                                   label={ "Key" } variant={ "standard" }
                                                   onChange={ event => updateParameterKey(index, event.target.value) }/>
                                    </Grid>
                                    <Grid item={ true } xs={ 2 }>
                                        <TextField id={ `update_parameter_value_${ index }` } value={ value.Value }
                                                   label={ "Value" } variant={ "standard" }
                                                   onChange={ event => updateParameterValue(index, event.target.value) }/>
                                    </Grid>
                                    <Grid item={ true } xs={ 8 }>
                                        <IconButton color="secondary" aria-label="remove" component="span" size={ "small" }
                                                    onClick={ () => removeParameter(value.Key) }><Remove/></IconButton>
                                    </Grid>
                                </Grid>
                            ),
                        ) }
                        <Grid item={ true } xs={ 2 }>
                        </Grid>
                        <Grid item={ true } xs={ 2 }>
                        </Grid>
                        <Grid item={ true } xs={ 8 }>
                            <IconButton color="primary" aria-label="remove" component="span"
                                        onClick={ addParameter }>
                                <Add/>
                            </IconButton>
                        </Grid>
                    </Grid>
                    <Grid item={ true } xs={ 12 } container={ true } spacing={ 1 }>
                        <Grid item={ true } xs={ 12 }>
                            <Typography variant={ 'h1' }>Hardware</Typography>
                            <Divider/>
                        </Grid>
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
                    <Grid item={ true } xs={ 12 } container={ true } spacing={ 1 }>
                        <Grid item={ true } xs={ 12 }>
                            <Typography variant={ 'h1' }>Graphic</Typography>
                            <Divider/>
                        </Grid>
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
                    <Grid item={ true } xs={ 12 } container={ true } justifyContent="flex-end">
                        <Button variant="contained" color="primary" size="small" onClick={ saveChanges }>Save</Button>
                    </Grid>
                </Grid>
            </Box>
        </Paper>
    );
};

export default DeviceEditContent;
