import { FC, useEffect, useState } from 'react';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Paper from '@material-ui/core/Paper';
import Grid from '@material-ui/core/Grid';
import Button from '@material-ui/core/Button';
import { createStyles, Theme, withStyles, WithStyles } from '@material-ui/core/styles';
import SearchIcon from '@material-ui/icons/Search';
import { useHistory } from 'react-router-dom';
import TableContainer from '@material-ui/core/TableContainer';
import Table from '@material-ui/core/Table';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import TableCell from '@material-ui/core/TableCell';
import TableBody from '@material-ui/core/TableBody';
import { getAllApps, deleteApp } from '../../services/app.service';
import IAppData from '../../types/app';
import Moment from 'react-moment';

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

export type AppsProps = WithStyles<typeof styles>;

const AppsPage: FC<AppsProps> = (props) => {
    const { classes } = props;

    const history = useHistory();

    function newAppClick(): void {
        history.push('/app/new');
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

    const ppSize = (bytes: number): string => {
        const KB = 1024;
        const MB = KB * 1024;
        const GB = MB * 1024;

        if (bytes >= GB) {
            return `${(bytes / GB).toFixed(2)}GB`;
        }

        if (bytes >= MB) {
            return `${(bytes / MB).toFixed(2)}MB`;
        }

        if (bytes >= KB) {
            return `${(bytes / KB).toFixed(2)}KB`;
        }

        return `${bytes}B`;
    };

    return (
        <Paper className={ classes.paper }>
            <AppBar className={ classes.searchBar } position="static" color="default" elevation={ 0 }>
                <Toolbar>
                    <Grid container={ true } spacing={ 2 } alignItems="center">
                        <Grid item={ true }>
                            <SearchIcon className={ classes.block } color="inherit"/>
                        </Grid>
                        <Grid item={ true } xs={ true }>
                        </Grid>
                        <Grid item={ true }>
                            <Button color="primary" variant="contained"
                                onClick={ newAppClick }>
                                New App
                            </Button>
                        </Grid>
                    </Grid>
                </Toolbar>
            </AppBar>
            <TableContainer component={Paper}>
                <Table className={classes.table} size="small" aria-label="a dense table">
                    <TableHead>
                        <TableRow>
                            <TableCell>ID</TableCell>
                            <TableCell>Created</TableCell>
                            <TableCell>Name</TableCell>
                            <TableCell>Platform</TableCell>
                            <TableCell>Bundle Identifier</TableCell>
                            <TableCell align="right">Version</TableCell>
                            <TableCell>Activity</TableCell>
                            <TableCell align="right">Size</TableCell>
                            <TableCell>Actions</TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {apps.map((app) => <TableRow key={app.ID}>
                            <TableCell component="th" scope="row">
                                {app.ID}
                            </TableCell>
                            <TableCell><Moment format="YYYY/MM/DD HH:mm:ss">{app.CreatedAt}</Moment></TableCell>
                            <TableCell>{app.Name}</TableCell>
                            <TableCell>{app.Platform}</TableCell>
                            <TableCell>{app.Identifier}</TableCell>
                            <TableCell align="right">{app.Version}</TableCell>
                            <TableCell>{app.LaunchActivity}</TableCell>
                            <TableCell align="right">{ppSize(app.Size)}</TableCell>
                            <TableCell>
                                <Button variant="contained" color="secondary" size="small" onClick={ () => {
                                    handleDeleteApp(app.ID as number);
                                }}>
                                    Delete
                                </Button>
                            </TableCell>
                        </TableRow>)}
                    </TableBody>
                </Table>
            </TableContainer>
        </Paper>
    );
};

export default withStyles(styles)(AppsPage);
