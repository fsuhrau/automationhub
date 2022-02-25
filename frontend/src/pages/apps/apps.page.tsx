import React, { useEffect, useState } from 'react';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Paper from '@mui/material/Paper';
import Grid from '@mui/material/Grid';
import Button from '@mui/material/Button';
import { useHistory } from 'react-router-dom';
import TableContainer from '@mui/material/TableContainer';
import Table from '@mui/material/Table';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TableCell from '@mui/material/TableCell';
import TableBody from '@mui/material/TableBody';
import { deleteApp, getAllApps } from '../../services/app.service';
import IAppData, { prettySize } from '../../types/app';
import Moment from 'react-moment';
import { ButtonGroup, Typography } from '@mui/material';
import { AndroidRounded, Apple } from '@mui/icons-material';
import DownloadIcon from '@mui/icons-material/Download';
import IconButton from "@mui/material/IconButton";
import DeleteForeverIcon from '@mui/icons-material/DeleteForever';

const AppsPage: React.FC = () => {

    const history = useHistory();

    function newAppClick(): void {
        history.push('/web/app/new');
    }

    const [apps, setApps] = useState<IAppData[]>([]);

    useEffect(() => {
        getAllApps().then(response => {
            setApps(response.data);
        }).catch(e => {
            console.log(e);
        });
    }, []);

    const handleDeleteApp = (appId: number): void => {
        deleteApp(appId).then(value => {
            setApps(prevState => {
                const newState = [...prevState];
                const index = newState.findIndex(value1 => value1.ID == appId);
                if (index > -1) {
                    newState.splice(index, 1);
                }
                return newState;
            });
        });
    };

    return (
        <div>
            <Paper sx={{ maxWidth: 1200, margin: 'auto', overflow: 'hidden' }}>
                <AppBar
                    position="static"
                    color="default"
                    elevation={0}
                    sx={{ borderBottom: '1px solid rgba(0, 0, 0, 0.12)' }}
                >
                    <Toolbar>
                        <Grid container={ true } spacing={ 2 } alignItems="center">
                            <Grid item={ true }>
                                <Typography variant={ 'h6' }>
                                    Apps
                                </Typography>
                            </Grid>
                            <Grid item={ true } xs={ true }>
                            </Grid>
                            <Grid item={ true }>
                                { false && <Button color="primary" variant="contained"
                                    onClick={ newAppClick }>
                                    New App
                                </Button>
                                }
                            </Grid>
                        </Grid>
                    </Toolbar>
                </AppBar>
                <TableContainer component={ Paper }>
                    <Table size="small" aria-label="a dense table">
                        <TableHead>
                            <TableRow>
                                <TableCell>ID</TableCell>
                                <TableCell>Created</TableCell>
                                <TableCell>OS</TableCell>
                                <TableCell>Name</TableCell>
                                <TableCell>Bundle Identifier / Activity</TableCell>
                                <TableCell align="right">Version</TableCell>
                                <TableCell align="right">Size</TableCell>
                                <TableCell>Tags</TableCell>
                                <TableCell align="right">Actions</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            { apps.map((app) => <TableRow key={ app.ID }>
                                <TableCell component="th" scope="row">
                                    { app.ID }
                                </TableCell>
                                <TableCell><Moment format="YYYY/MM/DD HH:mm:ss">{ app.CreatedAt }</Moment></TableCell>
                                <TableCell>{ app.Platform === 'android' ? (<AndroidRounded />) : (<Apple/>) }</TableCell>
                                <TableCell>{ app.Name }</TableCell>
                                <TableCell>{ app.Identifier } { app.LaunchActivity }</TableCell>
                                <TableCell align="right">{ app.Version }</TableCell>
                                <TableCell align="right">{ prettySize(app.Size) }</TableCell>
                                <TableCell>{ app.Tags }</TableCell>
                                <TableCell><ButtonGroup>
                                    <IconButton color="primary" size="small" href={`/${app.AppPath}`}>
                                        <DownloadIcon/>
                                    </IconButton>
                                    <IconButton color="secondary" size="small" onClick={ () => {
                                        handleDeleteApp(app.ID as number);
                                    } }>
                                        <DeleteForeverIcon/>
                                    </IconButton>
                                    </ButtonGroup>
                                </TableCell>
                            </TableRow>) }
                        </TableBody>
                    </Table>
                </TableContainer>
            </Paper>
        </div>
    );
};

export default AppsPage;
