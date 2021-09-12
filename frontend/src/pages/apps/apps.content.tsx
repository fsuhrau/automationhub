import { FC, useEffect, useState } from 'react';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Paper from '@material-ui/core/Paper';
import Grid from '@material-ui/core/Grid';
import Button from '@material-ui/core/Button';
import { createStyles, Theme, withStyles, WithStyles } from '@material-ui/core/styles';
import SearchIcon from '@material-ui/icons/Search';
import { useHistory } from 'react-router-dom';
import TableContainer from "@material-ui/core/TableContainer";
import Table from "@material-ui/core/Table";
import TableHead from "@material-ui/core/TableHead";
import TableRow from "@material-ui/core/TableRow";
import TableCell from "@material-ui/core/TableCell";
import TableBody from "@material-ui/core/TableBody";
import AppDataService from "../../services/app.service";
import IAppData from "../../types/app";

const styles = (theme: Theme): ReturnType<typeof createStyles> =>
    createStyles({
        paper: {
            maxWidth: 936,
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
        AppDataService.getAll().then(response => {
            console.log(response.data);
            setApps(response.data);
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
                            <TableCell>Name</TableCell>
                            <TableCell>Platform</TableCell>
                            <TableCell>Bundle Identifier</TableCell>
                            <TableCell align="right">Version</TableCell>
                            <TableCell>Activity</TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {apps.map((app) => <TableRow key={app.ID}>
                            <TableCell component="th" scope="row">
                                {app.ID}
                            </TableCell>
                            <TableCell>{app.Name}</TableCell>
                            <TableCell>{app.Platform}</TableCell>
                            <TableCell>{app.Identifier}</TableCell>
                            <TableCell align="right">{app.Version}</TableCell>
                            <TableCell>{app.LaunchActivity}</TableCell>
                        </TableRow>)}
                    </TableBody>
                </Table>
            </TableContainer>
        </Paper>
    );
};

export default withStyles(styles)(AppsPage);
