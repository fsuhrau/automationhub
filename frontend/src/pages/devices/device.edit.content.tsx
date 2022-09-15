import React, { useState } from 'react';
import Paper from '@mui/material/Paper';
import {
    Box,
    Checkbox,
    Divider,
    FormControl,
    FormControlLabel,
    IconButton,
    Input,
    InputLabel,
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

const StringIsNumber = (value: any): boolean => !isNaN(Number(value));

function ToArray(en: any): Array<Object> {
    return Object.keys(en).filter(StringIsNumber).map(key => ({ id: key, name: en[ key ] }));
}

function getConnectionType(): Array<{ id: DeviceConnectionType, name: string }> {
    return ToArray(DeviceConnectionType) as Array<{ id: DeviceConnectionType, name: string }>;
}

interface DeviceEditProps {
    device: IDeviceData
}

const DeviceEditContent: React.FC<DeviceEditProps> = props => {

    const navigate = useNavigate();

    const { device } = props;

    const [parameter, setParameter] = useState<IDeviceParameter[]>(device.Parameter);
    const updateParameterKey = (index: number, value: string) => {
        const newParams = [...parameter];
        newParams[ index ].Key = value;
        setParameter(newParams);
    };
    const updateParameterValue = (index: number, value: string) => {
        const newParams = [...parameter];
        newParams[ index ].Value = value;
        setParameter(newParams);
    };

    const addParameter = (): void => {
        setParameter([...parameter, { Key: '', Value: '' }]);
    };
    const removeParameter = (param: string) => {
        parameter.forEach((element, index) => {
            if (element.Key === param) parameter.splice(index, 1);
        });

        setParameter([...parameter]);
    };

    const [acknowledged, setAcknowledged] = useState<boolean>(device.IsAcknowledged);

    const connectionTypes = getConnectionType();
    const [connectionType, setConnectionType] = useState<DeviceConnectionType>(device.ConnectionParameter.ConnectionType);
    const handleConnectionTypeChange = (event: SelectChangeEvent) => {
        setConnectionType(+event.target.value as DeviceConnectionType);
    };

    const [ipAddress, setIPAddress] = useState<string>(device.ConnectionParameter.IP);
    const [port, setPort] = useState<number>(device.ConnectionParameter.Port);

    const saveChanges = () => {
        device.Parameter = parameter;
        device.IsAcknowledged = acknowledged;
        device.ConnectionParameter.ConnectionType = connectionType;
        device.ConnectionParameter.IP = ipAddress;
        device.ConnectionParameter.Port = port;
        updateDevice(device, device.ID).then(response => {
            navigate(`/web/device/${ device.ID }`);
        });
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
                                    control={ <Checkbox checked={ acknowledged }
                                        onChange={ (event, checked) => setAcknowledged(checked) }
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
                                <FormControl variant="filled" sx={ { m: 1, minWidth: 120 } }>
                                    <InputLabel id="demo-simple-select-filled-label">Connection Type</InputLabel>
                                    <Select
                                        labelId="demo-simple-select-filled-label"
                                        id="demo-simple-select-filled"
                                        value={ connectionType.toString() }
                                        onChange={ handleConnectionTypeChange }
                                    >
                                        { connectionTypes.map((value: { id: DeviceConnectionType, name: string }) => (
                                            <MenuItem value={ value.id }>{ value.name }</MenuItem>
                                        )) }
                                    </Select>
                                </FormControl>
                            </Grid>
                            { connectionType == DeviceConnectionType.Remote && (
                                <>
                                    <Grid item={ true } xs={ 2 }>
                                        IP:
                                    </Grid>
                                    <Grid item={ true } xs={ 10 }>
                                        <FormControl>
                                            <Input id="connection_ip" aria-describedby="ip-address"
                                                value={ ipAddress }
                                                onChange={ event => setIPAddress(event.target.value) }/>
                                        </FormControl>
                                    </Grid>
                                    <Grid item={ true } xs={ 2 }>
                                        Port:
                                    </Grid>
                                    <Grid item={ true } xs={ 10 }>
                                        <FormControl>
                                            <Input id="connection_ip" aria-describedby="ip-address"
                                                value={ port }
                                                onChange={ event => setPort(+event.target.value) }/>
                                        </FormControl>
                                    </Grid>
                                </>
                            )}
                        </Grid>
                    </Grid>
                </Grid>
                <Grid container={ true }>
                    <Grid item={ true } xs={ 12 } component="form"
                        sx={ {
                            '& .MuiTextField-root': { m: 1, width: '25ch' },
                        } }>
                        <Typography variant={ 'h6' }>Parameter</Typography>
                        <Divider/>
                        <br/>
                        <Grid container={ true }>
                            { parameter.map((value, index) => (
                                <>
                                    <Grid item={ true } xs={ 2 }>
                                        <TextField id={ `update_parameter_key_${ index }` } value={ value.Key }
                                            onChange={ event => updateParameterKey(index, event.target.value) }/>
                                    </Grid>
                                    <Grid item={ true } xs={ 2 }>
                                        <TextField id={ `update_parameter_value_${ index }` } value={ value.Value }
                                            onChange={ event => updateParameterValue(index, event.target.value) }/>
                                    </Grid>
                                    <Grid item={ true } xs={ 8 }>
                                        <IconButton color="secondary" aria-label="remove" component="span"
                                            onClick={ () => removeParameter(value.Key) }><Remove/></IconButton>
                                    </Grid>
                                </>
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
                <Grid container={ true } justifyContent="flex-end">
                    <Button variant="contained" color="primary" size="small" onClick={ saveChanges }>Save</Button>
                </Grid>
            </Box>
        </Paper>
    );
};

export default DeviceEditContent;
