import React, { FC } from 'react';
import Paper from '@material-ui/core/Paper';
import { createStyles, Theme, withStyles, WithStyles } from '@material-ui/core/styles';
import { Box, Checkbox, Divider, FormControlLabel, Typography } from '@material-ui/core';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Grid from '@material-ui/core/Grid';
import Button from '@material-ui/core/Button';
import IDeviceData from '../../types/device';
import { deleteDevice } from '../../services/device.service';
import { useHistory } from 'react-router-dom';
import { DeviceType } from '../../types/device.type.enum';
import { DeviceConnectionType } from '../../types/device.connection.type.enum';

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

interface DeviceShowProps extends WithStyles<typeof styles> {
    device: IDeviceData
}

const DeviceShowContent: FC<DeviceShowProps> = props => {
    const history = useHistory();

    const { device, classes } = props;
    const [expanded, setExpanded] = React.useState<string | false>(false);

    const handleChange = (panel: string) => (event: React.ChangeEvent<{}>, isExpanded: boolean) => {
        setExpanded(isExpanded ? panel : false);
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
                            <Button variant="contained" color="primary" size="small"
                                href={ `${ device.ID }/edit` }>Edit</Button>
                        </Grid>
                        <Button variant="contained" color="secondary" size="small" onClick={ () => {
                            deleteDevice(device.ID as number).then((result) =>
                                history.push('/web/devices'),
                            );
                        } }>
                            Delete
                        </Button>
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
                                    control={<Checkbox readOnly={ true } checked={device.IsAcknowledged} name="ack" />}
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
                                { DeviceConnectionType[ device.ConnectionParameter.ConnectionType ] }
                            </Grid>
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
                    </Grid>
                </Grid>
                <Grid container={ true }>
                    <Grid item={ true } xs={ 12 }>
                        <Typography variant={ 'h6' }>Parameter</Typography>
                        <Divider/>
                        <br/>
                        <Grid container={ true }>
                            { device.Parameter.map(value => (
                                <>
                                    <Grid item={ true } xs={ 2 }>
                                        { value.Key }
                                    </Grid>
                                    <Grid item={ true } xs={ 10 }>
                                        { value.Value }
                                    </Grid>
                                </>
                            ),
                            ) }
                        </Grid>
                    </Grid>
                </Grid>
                <br/>
                <Grid container={ true }>
                    <Grid item={ true } xs={ 12 }>
                        <Typography className={ classes.heading }>Hardware</Typography>
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
                        <Typography className={ classes.heading }>Graphic</Typography>
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
            </Box>
        </Paper>
    );
};

export default withStyles(styles)(DeviceShowContent);
