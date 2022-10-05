import React, { useEffect, useState } from 'react';
import Paper from '@mui/material/Paper';
import Grid from '@mui/material/Grid';
import { useNavigate } from 'react-router-dom';
import TableContainer from '@mui/material/TableContainer';
import Table from '@mui/material/Table';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TableCell from '@mui/material/TableCell';
import TableBody from '@mui/material/TableBody';
import { deleteAppBundle, getAppBundles } from '../../services/app.service';
import { IAppBinaryData, prettySize } from '../../types/app';
import Moment from 'react-moment';
import { ButtonGroup, Divider, FormControl, MenuItem, Select, SelectChangeEvent, Typography } from '@mui/material';
import { AndroidRounded, Apple } from '@mui/icons-material';
import DownloadIcon from '@mui/icons-material/Download';
import IconButton from '@mui/material/IconButton';
import DeleteForeverIcon from '@mui/icons-material/DeleteForever';
import { ApplicationProps } from "../../application/application.props";
import { useProjectAppContext } from "../../project/app.context";
import { TitleCard } from "../../components/title.card.component";
import { PlatformType } from "../../types/platform.type.enum";

const AppBundlesPage: React.FC<ApplicationProps> = (props: ApplicationProps) => {

    const {appState, dispatch} = props;

    const {projectId, appId} = useProjectAppContext();

    const navigate = useNavigate();

    function newAppClick(): void {
        navigate('new');
    }

    const app = appState.project?.Apps.find(a => a.ID === appId);

    const [bundles, setBundles] = useState<IAppBinaryData[]>([]);

    useEffect(() => {
        if (appId === 0) {
            const value = appState.project?.Apps === undefined ? null : appState.project?.Apps.length === 0 ? null : appState.project?.Apps[ 0 ].ID;
            if (value !== null) {
                navigate(`/project/${ projectId }/app/${ value }/bundles`)
            }
        }
    }, [appId, appState.project?.Apps])

    useEffect(() => {
        if (projectId !== null && appId > 0) {
            getAppBundles(projectId, appId).then(response => {
                setBundles(response.data);
            }).catch(e => {
                console.log(e);
            });
        }
    }, [projectId, appId]);

    const handleDeleteApp = (bundleId: number): void => {
        deleteAppBundle(projectId, app?.ID as number, bundleId).then(value => {
            setBundles(prevState => {
                const newState = [...prevState];
                const index = newState.findIndex(value1 => value1.ID == bundleId);
                if (index > -1) {
                    newState.splice(index, 1);
                }
                return newState;
            });
        });
    };

    const handleChange = (event: SelectChangeEvent): void => {
        if (appState.appId != +event.target.value) {
            navigate(`/project/${projectId}/app/${event.target.value}/bundles`)
        }
    };
    return (
        <Grid container={ true } spacing={ 2 }>
            <Grid item={ true } xs={ 12 }>
                <Typography variant={ "h1" }>Test Cases <FormControl variant="standard">
                    <Select
                        id="app-select"
                        label="App"
                        defaultValue={ `${ appId }` }
                        value={ `${ appId }` }
                        autoWidth={ true }
                        disableUnderline={ true }
                        sx={ {color: "black"} }
                        onChange={ handleChange }
                    >
                        {
                            appState.project?.Apps.map(app => (
                                <MenuItem key={ `app_item_${ app.ID }` } value={ app.ID }>{ app.Name }</MenuItem>
                            ))
                        }
                    </Select>
                </FormControl></Typography>
            </Grid>
            <Grid item={ true } xs={ 12 }>
                <Divider/>
            </Grid>
            <Grid item={ true } container={ true } xs={ 12 } alignItems={ "center" } justifyContent={ "center" }>
                <Grid
                    item={ true }
                    xs={ 12 }
                    style={ {maxWidth: 800} }
                >
                    <TitleCard title={ "Bundles" }>
                        <Paper sx={ {margin: 'auto', overflow: 'hidden'} }>
                            <Grid container={ true }>
                                <Grid item={ true } xs={ 12 } container={ true } spacing={1} sx={ {
                                    padding: 1,
                                    backgroundColor: '#fafafa',
                                    borderBottom: '1px solid rgba(0, 0, 0, 0.12)'
                                } }>
                                    <Grid item={true}>
                                        { app?.Platform === PlatformType.Android ? (<AndroidRounded/>) : (<Apple/>) }
                                    </Grid>
                                    <Grid item={true} xs={true}>
                                        <Typography variant={"body1"}>{ app?.Name }{' / '}{ app?.Identifier }</Typography>
                                    </Grid>
                                    <Grid item={true} xs={true} container={true} justifyContent={ "flex-end" }>
                                    </Grid>
                                </Grid>
                                <Grid item={ true } container={ true } xs={ 12 }>
                                    <TableContainer component={ Paper }>
                                        <Table size="small" aria-label="a dense table">
                                            <TableHead>
                                                <TableRow>
                                                    <TableCell>ID</TableCell>
                                                    <TableCell>Created</TableCell>
                                                    <TableCell >Version</TableCell>
                                                    <TableCell align="right">Size</TableCell>
                                                    <TableCell>Tags</TableCell>
                                                    <TableCell align="right">Actions</TableCell>
                                                </TableRow>
                                            </TableHead>
                                            <TableBody>
                                                { bundles.map((bundle) => <TableRow key={ bundle.ID }>
                                                    <TableCell component="th" scope="row">
                                                        { bundle.ID }
                                                    </TableCell>
                                                    <TableCell><Moment
                                                        format="YYYY/MM/DD HH:mm:ss">{ bundle.CreatedAt }</Moment></TableCell>
                                                    <TableCell>{ bundle.Version }</TableCell>
                                                    <TableCell align="right">{ prettySize(bundle.Size) }</TableCell>
                                                    <TableCell>{ bundle.Tags }</TableCell>
                                                    <TableCell><ButtonGroup>
                                                        <IconButton color="primary" size="small"
                                                                    href={ `/${ bundle.AppPath }` }>
                                                            <DownloadIcon/>
                                                        </IconButton>
                                                        <IconButton color="secondary" size="small" onClick={ () => {
                                                            handleDeleteApp(bundle.ID as number);
                                                        } }>
                                                            <DeleteForeverIcon/>
                                                        </IconButton>
                                                    </ButtonGroup>
                                                    </TableCell>
                                                </TableRow>) }
                                            </TableBody>
                                        </Table>
                                    </TableContainer>
                                </Grid>
                            </Grid>
                        </Paper>
                    </TitleCard>
                </Grid>
            </Grid>
        </Grid>
    );
};

export default AppBundlesPage;
