import React, { FC } from 'react';
import Paper from '@material-ui/core/Paper';
import { createStyles, Theme, withStyles, WithStyles } from '@material-ui/core/styles';
import {
    Box, Checkbox,
    Divider,
    FormControl, FormControlLabel,
    IconButton,
    Select,
    TextField,
    Typography,
} from '@material-ui/core';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Grid from '@material-ui/core/Grid';
import Button from '@material-ui/core/Button';
import IDeviceData from '../../types/device';
import { updateDevice } from '../../services/device.service';
import { useHistory } from 'react-router-dom';
import { DeviceConnectionType } from '../../types/device.connection.type.enum';
import IDeviceParameter from '../../types/device.parameter';
import { Add, Remove } from '@material-ui/icons';
import { DeviceType } from '../../types/device.type.enum';

const styles = (theme: Theme): ReturnType<typeof createStyles> =>
    createStyles({
        paper: {
            maxWidth: 1200,
            margin: 'auto',
            overflow: 'hidden',
        },
        searchBar: {
            borderBottom: '1px solid rgba(0, 0, 0, 0.12)',
        },
        searchInput: {
            fontSize: theme.typography.fontSize,
        },
        block: {
            display: 'block',
        },
        addUser: {
            marginRight: theme.spacing(1),
        },
        contentWrapper: {
            margin: '40px 16px',
        },
        heading: {
            fontSize: theme.typography.pxToRem(15),
            flexBasis: '33.33%',
            flexShrink: 0,
        },
        secondaryHeading: {
            fontSize: theme.typography.pxToRem(15),
            color: theme.palette.text.secondary,
        },
    });

const StringIsNumber = (value: any): boolean => !isNaN(Number(value));

function ToArray(en: any): Array<Object> {
    return Object.keys(en).filter(StringIsNumber).map(key => ({ id: key, name: en[ key ] }));
}

function getConnectionType(): Array<{ id: DeviceConnectionType, name: string }> {
    return ToArray(DeviceConnectionType) as Array<{ id: DeviceConnectionType, name: string }>;
}

interface DeviceEditProps extends WithStyles<typeof styles> {
    device: IDeviceData
}

const DeviceEditContent: FC<DeviceEditProps> = props => {
    const history = useHistory();

    const { device, classes } = props;

    const connectionTypes = getConnectionType();

    const [parameter, setParameter] = React.useState<IDeviceParameter[]>(device.Parameter);
    const updateParameterKey = (index: number, value: string) => {
        const newParams = [...parameter];
        newParams[index].Key = value;
        setParameter(newParams);
    };
    const updateParameterValue = (index: number, value: string) => {
        const newParams = [...parameter];
        newParams[index].Value = value;
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

    const [acknowledged, setAcknowledged] = React.useState<boolean>(device.IsAcknowledged);

    const saveChanges = () => {
        device.Parameter = parameter;
        device.IsAcknowledged = acknowledged;
        updateDevice(device, device.ID).then(response => {
            history.push(`/web/device/${device.ID}`);
        });
    };

    return (
        <Paper className={ classes.paper }>
            <AppBar className={ classes.searchBar } position="static" color="default" elevation={ 0 }>
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
                        <Typography variant={ 'h6' }>Device Infos</Typography>
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
                                { DeviceType[device.DeviceType] }
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
                                    control={<Checkbox checked={acknowledged} onChange={(event, checked)  => setAcknowledged(checked) } name="ack" />}
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
                                <FormControl className={ classes.formControl }>
                                    <Select native={ true }
                                        value={ device.ConnectionParameter.ConnectionType }
                                        name={ 'connection-type-selection' }
                                        onChange={ event => device.ConnectionParameter.ConnectionType = (+(event.target.value as string) as DeviceConnectionType) }
                                        inputProps={ {
                                            name: 'Connection Type',
                                            id: 'connection-types',
                                        } }>
                                        { connectionTypes.map((value: { id: DeviceConnectionType, name: string }) => (
                                            <option key={ 'ct_' + value.id.toString() }
                                                value={ value.id.toString() }>{ value.name }</option>
                                        )) }
                                    </Select>
                                </FormControl>
                            </Grid>
                            <Grid item={ true } xs={ 2 }>
                                IP:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                <TextField id="connection_ip" value={ device.ConnectionParameter.IP }
                                    onChange={ event => device.ConnectionParameter.IP = event.target.value }/>
                            </Grid>
                            <Grid item={ true } xs={ 2 }>
                                Port:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                <TextField id="connection_port" value={ device.ConnectionParameter.Port }
                                    onChange={ event => device.ConnectionParameter.Port = +event.target.value }/>
                            </Grid>
                        </Grid>
                    </Grid>
                </Grid>
                <Grid container={ true }>
                    <Grid item={ true } xs={ 12 }>
                        <Typography variant={ 'h6' }>Parameter</Typography>
                        <Divider/>
                        <br />
                        <Grid container={ true }>
                            { parameter.map((value, index)  => (
                                <>
                                    <Grid item={ true } xs={ 2 }>
                                        <TextField id={`update_parameter_key_${index}`} value={ value.Key }  onChange={ event => updateParameterKey(index, event.target.value) }/>
                                    </Grid>
                                    <Grid item={ true } xs={ 2 }>
                                        <TextField id={`update_parameter_value_${index}`} value={ value.Value }  onChange={ event => updateParameterValue(index, event.target.value) }/>
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
                        <Typography className={ classes.heading }>Hardware</Typography>
                        <Divider/>
                        <br />
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
                        <Typography className={ classes.heading }>Graphic</Typography>
                        <Divider/>
                        <br />
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
                <Grid container={ true } justify="flex-end">
                    <Button variant="contained" color="primary" size="small" onClick={saveChanges}>Save</Button>
                </Grid>
            </Box>
        </Paper>
    );
};

export default withStyles(styles)(DeviceEditContent);
