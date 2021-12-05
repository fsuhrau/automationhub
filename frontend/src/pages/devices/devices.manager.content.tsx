import { FC, useEffect, useState } from 'react';
import Paper from '@material-ui/core/Paper';
import { createStyles, Theme, withStyles, WithStyles } from '@material-ui/core/styles';
import DeviceTableComponent from '../../components/device-table.component';
import {
    Avatar,
    Button, IconButton,
    List,
    ListItem,
    ListItemAvatar,
    ListItemSecondaryAction,
    ListItemText,
    Typography,
} from '@material-ui/core';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Grid from '@material-ui/core/Grid';
import IDeviceData from '../../types/device';
import { getAllDevices } from '../../services/device.service';
import TableRow from '@material-ui/core/TableRow';
import TableCell from '@material-ui/core/TableCell';
import { Add, ArrowForward, Edit, PlayArrow } from '@material-ui/icons';
import IRealDeviceData from '../../types/real.device';
import IRealDeviceConnectionData from '../../types/real.device.connection';
import TableContainer from '@material-ui/core/TableContainer';
import Table from '@material-ui/core/Table';
import TableHead from '@material-ui/core/TableHead';
import TableBody from '@material-ui/core/TableBody';
import { useHistory } from 'react-router-dom';

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
    });

export type DevicesManagerProps = WithStyles<typeof styles>;

const DevicesManagerContent: FC<DevicesManagerProps> = props => {
    const { classes } = props;
    const history = useHistory();

    const [devices, setDevices] = useState<IDeviceData[]>([]);

    function openDetails(id: number): void {
        history.push(`/web/device/${id}`);
    }

    useEffect(() => {
        getAllDevices().then(response => {
            setDevices(response.data);
        }).catch(e => {
            console.log(e);
        });
    }, []);

    return (
        <Paper className={ classes.paper }>
            <AppBar className={ classes.searchBar } position="static" color="default" elevation={ 0 }>
                <Toolbar>
                    <Grid container={ true } spacing={ 2 } alignItems="center">
                        <Grid item={ true }>
                            <Typography variant={ 'h6' }>
                                Device Manager
                            </Typography>
                        </Grid>
                        <Grid item={ true } xs={ true }>
                        </Grid>
                        <Grid item={ true }>
                            <IconButton color="primary" size={'small'}
                                onClick={ (e) => {

                                } }>
                                <Add/>
                            </IconButton>

                        </Grid>
                    </Grid>
                </Toolbar>
            </AppBar>
            <TableContainer component={ Paper }>
                <Table className={ classes.table } size="small" aria-label="a dense table">
                    <TableHead>
                        <TableRow>
                            <TableCell>Name</TableCell>
                            <TableCell>Model</TableCell>
                            <TableCell>RAM</TableCell>
                            <TableCell>SOC</TableCell>
                            <TableCell>Status</TableCell>
                            <TableCell align="right"></TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        { devices.map((device) => <TableRow key={ device.ID }>
                            <TableCell component="th" scope="row">
                                { device.Name }
                            </TableCell>
                            <TableCell>
                                { device.HardwareModel }
                            </TableCell>
                            <TableCell>
                                { device.RAM }
                            </TableCell>
                            <TableCell>
                                { device.SOC }
                            </TableCell>
                            <TableCell>
                                { device.Status }
                            </TableCell>
                            <TableCell align="right">
                                <IconButton color="primary" size={'small'}
                                    onClick={ (e) => {
                                        openDetails(device.ID);
                                    } }>
                                    <ArrowForward/>
                                </IconButton>
                            </TableCell>
                        </TableRow>) }
                    </TableBody>
                </Table>
            </TableContainer>
        </Paper>
    );
};

export default withStyles(styles)(DevicesManagerContent);
