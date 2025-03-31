import React, {useEffect, useState} from 'react';
import Paper from '@mui/material/Paper';
import {useNavigate} from 'react-router-dom';
import TableContainer from '@mui/material/TableContainer';
import Table from '@mui/material/Table';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TableCell from '@mui/material/TableCell';
import TableBody from '@mui/material/TableBody';
import {deleteAppBundle, getAppBundles} from '../../services/app.service';
import {IAppBinaryData, prettySize} from '../../types/app';
import Moment from 'react-moment';
import {ButtonGroup, Typography} from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import IconButton from '@mui/material/IconButton';
import DeleteForeverIcon from '@mui/icons-material/DeleteForever';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {TitleCard} from "../../components/title.card.component";
import {PlatformType} from "../../types/platform.type.enum";
import {useApplicationContext} from "../../hooks/ApplicationProvider";
import {Box} from "@mui/system";
import Grid from "@mui/material/Grid2";
import PlatformTypeIcon from "../../components/PlatformTypeIcon";
import {useError} from "../../ErrorProvider";

const AppBundlesPage: React.FC = () => {

    const {project, projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const {setError} = useError()

    const navigate = useNavigate();

    function newAppClick(): void {
        navigate('new');
    }

    const app = project.Apps.find(a => a.ID === appId);

    const [bundles, setBundles] = useState<IAppBinaryData[]>([]);

    useEffect(() => {
        if (projectIdentifier !== null && appId != null) {
            getAppBundles(projectIdentifier, appId as number).then(response => {
                setBundles(response.data);
            }).catch(ex => {
                setError(ex);
            });
        }
    }, [projectIdentifier, appId]);

    const handleDeleteApp = (bundleId: number): void => {
        deleteAppBundle(projectIdentifier, app?.ID as number, bundleId).then(value => {
            setBundles(prevState => {
                const newState = [...prevState];
                const index = newState.findIndex(value1 => value1.ID == bundleId);
                if (index > -1) {
                    newState.splice(index, 1);
                }
                return newState;
            });
        }).catch(ex => setError(ex));
    };

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <TitleCard title={'App Bundles'}>
                <Paper sx={{width: '100%', margin: 'auto', overflow: 'hidden'}}>
                    <Grid container={true}>
                        <Grid size={{xs: 12, md: 12}} container={true} spacing={1} sx={{
                            padding: 1,
                            borderBottom: '1px solid rgba(0, 0, 0, 0.12)'
                        }}>
                            <Grid>
                                <PlatformTypeIcon platformType={app?.Platform as PlatformType}/>
                            </Grid>
                            <Grid>
                                <Typography variant={"body1"}>{app?.Name}{' / '}{app?.Identifier}</Typography>
                            </Grid>
                            <Grid container={true} justifyContent={"flex-end"}>
                            </Grid>
                        </Grid>
                        <Grid container={true} size={{xs: 12, md: 12}}>
                            <TableContainer component={Paper}>
                                <Table size="small" aria-label="a dense table">
                                    <TableHead>
                                        <TableRow>
                                            <TableCell>ID</TableCell>
                                            <TableCell>Created</TableCell>
                                            <TableCell>Version</TableCell>
                                            <TableCell align="right">Size</TableCell>
                                            <TableCell>Tags</TableCell>
                                            <TableCell align="right">Actions</TableCell>
                                        </TableRow>
                                    </TableHead>
                                    <TableBody>
                                        {bundles.map((bundle) => <TableRow key={bundle.ID}>
                                            <TableCell component="th" scope="row">
                                                {bundle.ID}
                                            </TableCell>
                                            <TableCell><Moment
                                                format="YYYY/MM/DD HH:mm:ss">{bundle.CreatedAt}</Moment></TableCell>
                                            <TableCell>{bundle.Version}</TableCell>
                                            <TableCell align="right">{prettySize(bundle.Size)}</TableCell>
                                            <TableCell>{bundle.Tags}</TableCell>
                                            <TableCell align="right"><ButtonGroup>
                                                <IconButton color="primary" size="small"
                                                            href={`/upload/${bundle.AppPath}`}>
                                                    <DownloadIcon/>
                                                </IconButton>
                                                <IconButton color="secondary" size="small" onClick={() => {
                                                    handleDeleteApp(bundle.ID as number);
                                                }}>
                                                    <DeleteForeverIcon/>
                                                </IconButton>
                                            </ButtonGroup>
                                            </TableCell>
                                        </TableRow>)}
                                    </TableBody>
                                </Table>
                            </TableContainer>
                        </Grid>
                    </Grid>
                </Paper>
            </TitleCard>
        </Box>
    );
};

export default AppBundlesPage;
